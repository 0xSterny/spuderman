package spider

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

type SMBFS struct {
	Share *smb2.Share
}

func (s *SMBFS) Open(name string) (fs.File, error) {
	// smb2.Share.Open returns *smb2.File which implements fs.File ?
	// Check smb2 docs: Open(name string) (*File, error). File implements Read, Close, Seek...
	// We need it to satisfy fs.File which requires Stat() as well.
	// go-smb2 File has Stat().
	return s.Share.Open(name)
}

func (s *SMBFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	// Implement simple recursive walker
	// root is relative to share
	return s.walk(root, fn)
}

func (s *SMBFS) walk(path string, fn fs.WalkDirFunc) error {
	// Normalize path
	if path == "" {
		path = "."
	}

	// Read dir
	infos, err := s.Share.ReadDir(path)
	if err != nil {
		return fn(path, nil, err)
	}

	// Process current dir?
	// Usually WalkDir calls fn for the root first.
	// We need to stat the root to pass DirEntry
	// Manually skipped for now to avoid unused var error

	for _, info := range infos {
		name := info.Name()
		if name == "." || name == ".." {
			continue
		}

		fullPath := filepath.Join(path, name)
		// Adjust for SMB path separators if needed (usually backslash)
		fullPath = strings.ReplaceAll(fullPath, "\\", "/")

		d := fs.FileInfoToDirEntry(info)

		if err := fn(fullPath, d, nil); err != nil {
			if err == fs.SkipDir {
				return nil
			}
			return err
		}

		if info.IsDir() {
			// Do not follow symlinks/reparse points (Junctions)
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			if err := s.walk(fullPath, fn); err != nil {
				return err
			}
		}
	}
	return nil
}
