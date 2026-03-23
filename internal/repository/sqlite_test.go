package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRepository(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Verify tables exist by querying
	var count int
	err = repo.DB().QueryRow("SELECT COUNT(*) FROM player").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query player table: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 player row, got %d", count)
	}

	err = repo.DB().QueryRow("SELECT COUNT(*) FROM quests").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query quests table: %v", err)
	}
}

func TestBootstrap(t *testing.T) {
	// Test that bootstrap creates directory and DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "questline", "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to bootstrap: %v", err)
	}
	defer repo.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Data directory was not created")
	}

	// Verify player singleton exists
	var level int
	err = repo.DB().QueryRow("SELECT level FROM player WHERE id = 1").Scan(&level)
	if err != nil {
		t.Fatalf("Player singleton not created: %v", err)
	}
	if level != 1 {
		t.Errorf("Expected level 1, got %d", level)
	}
}
