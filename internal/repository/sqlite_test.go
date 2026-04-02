package repository

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewRepository(t *testing.T) {
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

	var userVersion int
	if err := repo.DB().QueryRow("PRAGMA user_version").Scan(&userVersion); err != nil {
		t.Fatalf("Failed to read user_version: %v", err)
	}
	if userVersion != schemaVersionV2 {
		t.Fatalf("Expected user_version=%d, got %d", schemaVersionV2, userVersion)
	}

	err = repo.DB().QueryRow("SELECT COUNT(*) FROM quests").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query quests table: %v", err)
	}
}

func TestBootstrap(t *testing.T) {
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

	var legacyStatusCount int
	err = repo.DB().QueryRow("SELECT COUNT(*) FROM quests WHERE status IN ('TODO','DONE','DROPPED')").Scan(&legacyStatusCount)
	if err != nil {
		t.Fatalf("Failed to evaluate status enum: %v", err)
	}
	if legacyStatusCount != 0 {
		t.Fatalf("Expected no legacy statuses in fresh v2 schema, got %d", legacyStatusCount)
	}

	var hasFlowStatus int
	err = repo.DB().QueryRow("SELECT COUNT(*) FROM pragma_table_info('player') WHERE name = 'flow_status'").Scan(&hasFlowStatus)
	if err != nil {
		t.Fatalf("Failed to inspect player columns: %v", err)
	}
	if hasFlowStatus != 1 {
		t.Fatalf("Expected player.flow_status column in v2 schema")
	}

	var userVersion int
	err = repo.DB().QueryRow("PRAGMA user_version").Scan(&userVersion)
	if err != nil {
		t.Fatalf("Failed to read user_version: %v", err)
	}
	if userVersion != schemaVersionV2 {
		t.Fatalf("Expected user_version=%d, got %d", schemaVersionV2, userVersion)
	}
}

func TestMigrationRollback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "broken-v1.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open fixture db: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE quests (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT DEFAULT 'TODO' CHECK (status IN ('TODO', 'DONE', 'DROPPED')),
			due_date TEXT,
			created_at TEXT,
			completed_at TEXT
		);
	`); err != nil {
		t.Fatalf("failed to create broken quests table: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE player (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			level INTEGER DEFAULT 1 NOT NULL,
			current_xp INTEGER DEFAULT 0 NOT NULL,
			total_xp_earned INTEGER DEFAULT 0 NOT NULL,
			quests_completed INTEGER DEFAULT 0 NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		t.Fatalf("failed to create player table: %v", err)
	}

	if _, err := db.Exec("INSERT INTO quests (id, title, status, created_at) VALUES ('q1', 'broken row', 'TODO', NULL)"); err != nil {
		t.Fatalf("failed to seed broken quest: %v", err)
	}
	if _, err := db.Exec("INSERT INTO player (id, level, current_xp, total_xp_earned, quests_completed, updated_at) VALUES (1, 1, 0, 0, 0, datetime('now'))"); err != nil {
		t.Fatalf("failed to seed player: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
		t.Fatalf("failed to set user_version=1: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close fixture db: %v", err)
	}

	_, err = New(dbPath)
	if err == nil {
		t.Fatalf("expected migration to fail")
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to reopen fixture db: %v", err)
	}
	defer db.Close()

	var userVersion int
	if err := db.QueryRow("PRAGMA user_version").Scan(&userVersion); err != nil {
		t.Fatalf("failed to read user_version after rollback: %v", err)
	}
	if userVersion != schemaVersionV1 {
		t.Fatalf("expected user_version to remain v1 after rollback, got %d", userVersion)
	}

	var hasLegacyQuests int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='quests'").Scan(&hasLegacyQuests); err != nil {
		t.Fatalf("failed to inspect quests table after rollback: %v", err)
	}
	if hasLegacyQuests != 1 {
		t.Fatalf("expected legacy quests table to remain after rollback")
	}

	var hasTempQuests int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='quests_new'").Scan(&hasTempQuests); err != nil {
		t.Fatalf("failed to inspect temporary table after rollback: %v", err)
	}
	if hasTempQuests != 0 {
		t.Fatalf("expected temporary migration tables to rollback")
	}
}
