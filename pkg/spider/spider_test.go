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
	fs := &spider.LocalFS{} // Use real local FS on temp dir
	dedup := utils.NewDeduplicator()
	s := spider.NewSpider(cfg, m, fs, dedup)

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

	fs := &spider.LocalFS{}
	dedup := utils.NewDeduplicator()
	s := spider.NewSpider(cfg, m, fs, dedup)

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
