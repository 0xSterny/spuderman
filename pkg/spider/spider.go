package spider

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"spuderman/pkg/extractor"
	"spuderman/pkg/matcher"
	"spuderman/pkg/utils"
)

type FileSystem interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
	Open(name string) (fs.File, error)
}

// LocalFS wrapper
type LocalFS struct{}

func (l *LocalFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

func (l *LocalFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

type Config struct {
	MaxDepth   int
	Threads    int
	LootDir    string
	NoDownload bool

	// Structured Loot options
	Structured bool
	Host       string
	Share      string
}

type DownloadJob struct {
	Path   string
	Reason string
}

type Spider struct {
	Config    Config
	Matcher   *matcher.Matcher
	FS        FileSystem
	Dedup     *utils.Deduplicator
	Semaphore chan struct{} // If set, use this semaphore for concurrency limitation
	Reporter  Reporter

	// Async Download
	downloadChan chan DownloadJob
	downloadWG   sync.WaitGroup
}

func NewSpider(cfg Config, m *matcher.Matcher, fs FileSystem, dedup *utils.Deduplicator, reporter Reporter) *Spider {
	if reporter == nil {
		reporter = &ConsoleReporter{}
	}
	return &Spider{
		Config:   cfg,
		Matcher:  m,
		FS:       fs,
		Dedup:    dedup,
		Reporter: reporter,
	}
}

func (s *Spider) Walk(target string) {
	utils.LogInfo("Starting walk on: %s", target)

	// Semaphore for concurrency
	// If shared semaphore is provided, use it. Otherwise create local based on threads config.
	var sem chan struct{}
	if s.Semaphore != nil {
		sem = s.Semaphore
	} else {
		// Minimum 1 thread
		t := s.Config.Threads
		if t < 1 {
			t = 1
		}
		sem = make(chan struct{}, t)
	}

	// Init Async Downloads
	if !s.Config.NoDownload {
		s.downloadChan = make(chan DownloadJob, 100) // Buffer 100

		// Start workers (Fixed 4 workers? or based on threads?)
		// Let's use 2 workers just to offload.
		numWorkers := 2
		s.downloadWG.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go s.downloadWorker()
		}
	}

	var wg sync.WaitGroup

	err := s.FS.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			utils.LogWarning("Error accessing %s: %v", path, err)
			return nil // Continue walking
		}

		if d.IsDir() {
			// Check depth (naive implementation, just check separator count)
			// TODO: Better depth check
			return nil
		}

		// Check exclusion first
		if s.Matcher.CheckExclude(path) {
			utils.LogDebug("Skipping excluded file (or dir): %s", path)
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Check directory filter
		if !s.Matcher.CheckDir(path) {
			return nil
		}

		// 3. Check Extension Filter (Extension is a Filter, not a Search Term)
		if !s.Matcher.CheckExtension(d.Name()) {
			return nil
		}

		// Search Logic: (NameMatch || ContentMatch)
		// If no search terms provided (no -f, no -c), we consider it a match (if extension matched). (Dump all mode)

		hasNameTerm := len(s.Matcher.Config.Filenames) > 0
		hasContentTerm := len(s.Matcher.Config.Content) > 0

		// Optimization: If no search terms, match immediately
		if !hasNameTerm && !hasContentTerm {
			s.handleMatch(path, "Extension/All")
			return nil
		}

		// Check Filename Regex
		if hasNameTerm {
			if s.Matcher.CheckFilenameRegex(d.Name()) {
				s.handleMatch(path, "Filename")
				// Short-circuit: OR logic means if name matches, we are done.
				return nil
			}
		}

		// Check Content Regex (only if needed)
		if hasContentTerm {
			// If we are here, Name checks failed (or weren't present).
			// We MUST check content.

			wg.Add(1)
			sem <- struct{}{}
			go func(fPath string, fEntry fs.DirEntry) {
				defer wg.Done()
				defer func() { <-sem }()

				// Open file
				f, err := s.FS.Open(fPath)
				if err != nil {
					return
				}
				defer f.Close()

				// Extract
				extEngine := extractor.GetExtractor(fPath)
				// Limit extraction to 10MB to prevent OOM
				limitReader := io.LimitReader(f, 10*1024*1024)
				text, err := extEngine.Extract(limitReader, fPath)
				if err != nil {
					return
				}

				matched, snippet := s.Matcher.CheckContent(text)
				if matched {
					reason := "Content"
					if snippet != "" {
						reason = "Content: " + utils.Bold(snippet)
					}
					s.handleMatch(fPath, reason)
				}
			}(path, d)
		}

		return nil
	})

	if err != nil {
		utils.LogError("Error walking %s: %v", target, err)
	}

	wg.Wait()

	// Cleanup Downloads
	if !s.Config.NoDownload {
		close(s.downloadChan)
		s.downloadWG.Wait()
	}
}

func (s *Spider) downloadWorker() {
	defer s.downloadWG.Done()
	for job := range s.downloadChan {
		hash, err := s.downloadFile(job.Path)
		if err != nil {
			utils.LogDebug("Download failed: %v", err)
			// Report failure?
		}

		// Report match after download (to include hash)
		s.Reporter.Report(MatchResult{
			Path:   job.Path,
			Reason: job.Reason,
			Host:   s.Config.Host,
			Share:  s.Config.Share,
			Hash:   hash,
		})
	}
}

func (s *Spider) handleMatch(path string, reason string) {
	utils.LogSuccess("Match found (%s): %s", reason, path)

	if !s.Config.NoDownload {
		// Queue for async download
		s.downloadChan <- DownloadJob{
			Path:   path,
			Reason: reason,
		}
	} else {
		// Report match immediately
		s.Reporter.Report(MatchResult{
			Path:   path,
			Reason: reason,
			Host:   s.Config.Host,
			Share:  s.Config.Share,
		})
	}
}

func (s *Spider) downloadFile(path string) (string, error) {
	// Create loot dir if not exists
	if err := os.MkdirAll(s.Config.LootDir, 0755); err != nil {
		utils.LogError("Failed to create loot dir: %v", err)
		return "", err
	}

	// Open source
	src, err := s.FS.Open(path)
	if err != nil {
		utils.LogError("Failed to open file for download %s: %v", path, err)
		return "", err
	}
	defer src.Close()

	// Create dest
	var destPath string
	if s.Config.Structured {
		// LootDir/Host/Share/Path...
		// Ensure Host/Share are safe
		safeHost := strings.ReplaceAll(s.Config.Host, ":", "")
		safeShare := strings.ReplaceAll(s.Config.Share, "\\", "")
		safeShare = strings.ReplaceAll(safeShare, "/", "")

		// Join effectively sanitizes middle segments? No, we rely on Join.
		// Path comes in as "foo/bar.txt".
		destPath = filepath.Join(s.Config.LootDir, safeHost, safeShare, path)
	} else {
		// Old flat behavior
		// Sanitize path for local fs
		// Let's try to keep directory structure if possible, but path separators might differ.
		// Making it unique: Replace separators with _
		safeName := strings.ReplaceAll(path, "\\", "_")
		safeName = strings.ReplaceAll(safeName, "/", "_")
		safeName = strings.ReplaceAll(safeName, ":", "")
		destPath = filepath.Join(s.Config.LootDir, safeName)
	}

	// Create parent dirs
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		utils.LogError("Failed to create loot subdirs for %s: %v", destPath, err)
		return "", err
	}

	dst, err := os.Create(destPath)
	if err != nil {
		utils.LogError("Failed to create destination file %s: %v", destPath, err)
		return "", err
	}
	defer dst.Close()

	// Hash while downloading
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)

	// Download
	_, err = io.Copy(dst, tee)
	if err != nil {
		dst.Close() // Ensure closed before removing
		utils.LogError("Failed to download file %s: %v", path, err)
		os.Remove(destPath) // Cleanup partial
		return "", err
	} else {
		dst.Close() // Close successful file

		// Check hash
		hash := hex.EncodeToString(hasher.Sum(nil))
		if s.Dedup.IsDuplicate(hash) {
			utils.LogInfo("Duplicate file detected (Hash: %s), removing: %s", hash[:8], path)
			os.Remove(destPath)
			// We still return hash? It was a match, just duplicate content.
			// Maybe report it?
			return hash, nil
		} else {
			utils.LogDownload("Downloaded to: %s [Hash: %s]", destPath, hash[:8])
			return hash, nil
		}
	}
}
