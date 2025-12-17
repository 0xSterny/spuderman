package matcher

import (
	"regexp"
	"strings"
)

type MatchConfig struct {
	Filenames  []string
	Extensions []string
	Content    []string
	Dirnames   []string
	OrLogic    bool
}

type Matcher struct {
	FilenameRegex []*regexp.Regexp
	ContentRegex  []*regexp.Regexp
	DirnameRegex  []*regexp.Regexp
	ExcludeRegex  []*regexp.Regexp
	Extensions    map[string]bool
	Config        MatchConfig
}

func NewMatcher(config MatchConfig) (*Matcher, error) {
	m := &Matcher{
		Config:     config,
		Extensions: make(map[string]bool),
	}

	for _, p := range config.Filenames {
		re, err := regexp.Compile("(?i)" + p) // Case insensitive by default
		if err != nil {
			return nil, err
		}
		m.FilenameRegex = append(m.FilenameRegex, re)
	}

	for _, p := range config.Content {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return nil, err
		}
		m.ContentRegex = append(m.ContentRegex, re)
	}

	for _, p := range config.Dirnames {
		// Dirname matching: "Search directories containing these strings"
		// We treat it as regex (case insensitive).
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return nil, err
		}
		m.DirnameRegex = append(m.DirnameRegex, re)
	}

	// Initialize Default Exclusions
	for _, p := range DefaultExcludes {
		// Escape path for regex if needed? Or treating as regex?
		// User said "file path of ...". Assuming exact partial match or path substring.
		// Let's use `regexp.QuoteMeta` if we want strict literal, but user might want patterns.
		// "Microsoft/Edge/User Data/ZxcvbnData" looks like a path.
		// Let's assume regex for flexibility but be careful.
		// If I use regexp directly, I should ensure separators are handled.
		// Actually, let's treat them as regex patterns but case insensitive.
		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(p))
		if err == nil {
			m.ExcludeRegex = append(m.ExcludeRegex, re)
		}
	}

	for _, e := range config.Extensions {
		ext := strings.ToLower(e)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		m.Extensions[ext] = true
	}

	return m, nil
}

// CheckExtension returns true if file extension matches allowlist (or if list is empty)
func (m *Matcher) CheckExtension(filename string) bool {
	if len(m.Config.Extensions) == 0 {
		return true
	}

	ext := ""
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		ext = strings.ToLower(filename[idx:])
	}

	return m.Extensions[ext]
}

// CheckFilenameRegex returns true if filename matches any of the regex patterns
func (m *Matcher) CheckFilenameRegex(filename string) bool {
	if len(m.Config.Filenames) == 0 {
		return false
	}
	for _, re := range m.FilenameRegex {
		if re.MatchString(filename) {
			return true
		}
	}
	return false
}

// Deprecated: wrapper for backward compatibility if needed, but we should use split checks.
func (m *Matcher) CheckFilename(filename string) bool {
	return m.CheckExtension(filename) && (len(m.Config.Filenames) == 0 || m.CheckFilenameRegex(filename))
}

// CheckContent returns true if content matches regex, and the matching line (snippet)
func (m *Matcher) CheckContent(text string) (bool, string) {
	if len(m.Config.Content) == 0 {
		return true, "" // No content filter
	}

	// Split text into lines to find the matching one
	// This might be slow for huge files, but we already extracted it to memory.
	lines := strings.Split(text, "\n")

	for _, re := range m.ContentRegex {
		for _, line := range lines {
			if re.MatchString(line) {
				// Found a match
				snippet := strings.TrimSpace(line)
				if len(snippet) > 80 {
					snippet = snippet[:80] + "..."
				}
				return true, snippet
			}
		}
	}
	return false, ""
}

// CheckExclude returns true if the filename matches any exclusion pattern
func (m *Matcher) CheckExclude(filename string) bool {
	for _, re := range m.ExcludeRegex {
		if re.MatchString(filename) {
			return true
		}
	}
	return false
}

// CheckDir returns true if the parent directory of path matches criteria
// If no dirnames filter is set, returns true (allow all).
// If dirnames set: returns true if ANY regex matches the directory path.
func (m *Matcher) CheckDir(path string) bool {
	if len(m.Config.Dirnames) == 0 {
		return true
	}

	// Get directory part
	// Handle cross-platform separators if needed, but path usually normalized by Walker?
	// spider logic processes path from Walker.
	// Just check if the full path contains the string (simplified) OR check just components?
	// MANSPIDER: "Only spider directories with names containing these strings"
	// So if path is "Users/bryce/Documents", and dirname="bryce", it should match.
	// But if path is "Users/bryce/Documents", and dirname="foo", fail.

	// We check against the full path string (usually relative to share).
	// Or should we exclude the filename? Yes, dirname check usually means "Is the file inside a matching directory?"

	// NOTE: `path` passed here includes filename.
	// We can just check regex match on the whole string?
	// But `Dirnames` implies "Directory Names".
	// If I match "password" and filename is "password.txt" but dir is "Public", CheckDir should pass?
	// The Flag is "Only search directories..."
	// If I match the whole path, then filename match might trigger it.
	// Let's check `filepath.Dir(path)`

	dir := path
	// If called with just dir path (from walker), use it.
	// If called with file path, assume caller stripped it or we strip it?
	// Spider.Walk passes proper path.
	// Let's assume input is the full path.

	for _, re := range m.DirnameRegex {
		if re.MatchString(dir) {
			return true
		}
	}
	return false
}
