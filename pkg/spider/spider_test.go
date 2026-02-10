package spider_test

import (
	"os"
	"path/filepath"
	"testing"

	"spuderman/pkg/matcher"
	"spuderman/pkg/spider"
	"spuderman/pkg/utils"
)

func TestSpiderLocal(t *testing.T) {
	// Setup temp dir
	tmpDir := t.TempDir()

	// Create files
	secrets := filepath.Join(tmpDir, "secrets.txt")
	err := os.WriteFile(secrets, []byte("This file contains a password123!"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	boring := filepath.Join(tmpDir, "boring.txt")
	err = os.WriteFile(boring, []byte("Just some random text"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Setup Config
	lootDir := filepath.Join(tmpDir, "loot")
	cfg := spider.Config{
		MaxDepth:   5,
		Threads:    1,
		LootDir:    lootDir,
		NoDownload: false,
	}

	// Setup Matcher (Match "password")
	mConfig := matcher.MatchConfig{
		Content: []string{"password"},
	}
	m, err := matcher.NewMatcher(mConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Setup Spider
	m.ExcludeRegex = nil    // Fix: TempDir is in AppData which is excluded by default
	fs := &spider.LocalFS{} // Use real local FS on temp dir
	dedup := utils.NewDeduplicator()
	s := spider.NewSpider(cfg, m, fs, dedup, nil)

	// Run
	s.Walk(tmpDir)

	// Verify Loot
	// Should have downloaded secrets.txt (sanitized path)
	// Original path: C:\...\secrets.txt -> C____secrets.txt (based on our sanitize logic)
	// Let's check if *any* file exists in loot dir
	entries, err := os.ReadDir(lootDir)
	if err != nil {
		t.Fatalf("Failed to read loot dir: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected loot to be found, but loot dir is empty")
	} else {
		t.Logf("Found matches: %d", len(entries))
		for _, e := range entries {
			t.Logf("Loot: %s", e.Name())
		}
	}
}

func TestKnownGoods(t *testing.T) {
	t.Log("Starting TestKnownGoods...")
	// Setup mock structure mirroring user request
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, "Users", "bryce.miller", "Desktop")

	files := map[string]string{
		"cornhuliopassword.txt":         "content with password 1",
		"flexible/Not_a password.txt":   "content with password 2",
		"mr-worldwide/this.txt":         "content with password 3",
		"mr-worldwide/folder1/that.txt": "content with password 4",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(base, relPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Setup Config
	lootDir := filepath.Join(tmpDir, "loot")
	cfg := spider.Config{
		MaxDepth:   10,
		Threads:    1,
		LootDir:    lootDir,
		NoDownload: false,
	}

	// Match "password"
	mConfig := matcher.MatchConfig{
		Content:   []string{"password"},
		Filenames: []string{"password"},
		OrLogic:   true, // Match filename OR content
	}
	m, err := matcher.NewMatcher(mConfig)
	if err != nil {
		t.Fatal(err)
	}
	// Disable Default Exclusions for this test to be safe?
	// The paths are "Users\bryce.miller\Desktop..." which might look like "Users" (not excluded)
	// But "Desktop" is not excluded.
	// Wait, "Users" is not in DefaultExcludes. "Windows" and "Program Files" are.
	// Maybe "AppData" is interfering if TempDir is in AppData?
	// t.TempDir() usually returns C:\Users\User\AppData\Local\Temp\...
	// If we are scanning t.TempDir(), and "AppData" is in excludes...
	// AND the spider checks path against regex.
	// matcher.CheckExclude checks if regex matches the filename/path.
	// If path contains "AppData", it might be excluded!

	// FIX: Clear default excludes for test located in AppData
	m.ExcludeRegex = nil

	fs := &spider.LocalFS{}
	dedup := utils.NewDeduplicator()
	s := spider.NewSpider(cfg, m, fs, dedup, nil)

	s.Walk(tmpDir)

	// Verify all 4 files are in loot
	// Limitation: Local loot structure flattens names unless structured
	// With default settings (flat), we expect 4 files.
	// But duplicate names might overwrite if logic isn't unique?
	// Spider.downloadFile uses: safeName := strings.ReplaceAll(path, "\\", "_") ...
	// So they should be unique.

	entries, err := os.ReadDir(lootDir)
	if err != nil {
		t.Fatalf("Failed to read loot dir: %v", err)
	}

	if len(entries) != 4 {
		t.Errorf("Expected 4 files in loot, found %d", len(entries))
		for _, e := range entries {
			t.Logf("Found: %s", e.Name())
		}
	}
}
