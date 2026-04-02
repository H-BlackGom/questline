package repository

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateV1ToV2(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "mvp1-fixture.db")

	seeded := createMVP1FixtureDB(t, dbPath)
	if seeded != 3 {
		t.Fatalf("expected 3 seeded quests, got %d", seeded)
	}

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to bootstrap migrated repository: %v", err)
	}
	defer repo.Close()

	var userVersion int
	if err := repo.DB().QueryRow("PRAGMA user_version").Scan(&userVersion); err != nil {
		t.Fatalf("failed to read user_version: %v", err)
	}
	if userVersion != schemaVersionV2 {
		t.Fatalf("expected user_version=%d, got %d", schemaVersionV2, userVersion)
	}

	var migratedCount int
	if err := repo.DB().QueryRow("SELECT COUNT(*) FROM quests").Scan(&migratedCount); err != nil {
		t.Fatalf("failed to count migrated quests: %v", err)
	}
	if migratedCount != seeded {
		t.Fatalf("expected %d migrated quests, got %d", seeded, migratedCount)
	}

	assertQuestStatus(t, repo.DB(), "q_todo", "pending")
	assertQuestStatus(t, repo.DB(), "q_done", "completed")
	assertQuestStatus(t, repo.DB(), "q_dropped", "archived")

	var flowStatus string
	var streak int
	if err := repo.DB().QueryRow("SELECT flow_status, streak_days FROM player WHERE id = 1").Scan(&flowStatus, &streak); err != nil {
		t.Fatalf("failed to read migrated player columns: %v", err)
	}
	if flowStatus != "smooth" {
		t.Fatalf("expected default flow_status=smooth, got %q", flowStatus)
	}
	if streak != 0 {
		t.Fatalf("expected default streak_days=0, got %d", streak)
	}
}

func createMVP1FixtureDB(t *testing.T, dbPath string) int {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open fixture database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE quests (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT DEFAULT 'TODO' CHECK (status IN ('TODO', 'DONE', 'DROPPED')),
			due_date TEXT,
			created_at TEXT NOT NULL,
			completed_at TEXT
		);
	`); err != nil {
		t.Fatalf("failed to create v1 quests table: %v", err)
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
		t.Fatalf("failed to create v1 player table: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO quests (id, title, status, due_date, created_at, completed_at) VALUES
			('q_todo', 'todo quest', 'TODO', NULL, '2026-04-01T00:00:00Z', NULL),
			('q_done', 'done quest', 'DONE', '2026-04-02', '2026-04-01T00:00:00Z', '2026-04-01T04:00:00Z'),
			('q_dropped', 'dropped quest', 'DROPPED', NULL, '2026-04-01T00:00:00Z', NULL);
	`); err != nil {
		t.Fatalf("failed to insert fixture quests: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO player (id, level, current_xp, total_xp_earned, quests_completed, updated_at)
		VALUES (1, 3, 30, 180, 2, '2026-04-01T00:00:00Z');
	`); err != nil {
		t.Fatalf("failed to insert fixture player: %v", err)
	}

	if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
		t.Fatalf("failed to set fixture user_version: %v", err)
	}

	return 3
}

func assertQuestStatus(t *testing.T, db *sql.DB, questID, expected string) {
	t.Helper()

	var actual string
	if err := db.QueryRow("SELECT status FROM quests WHERE id = ?", questID).Scan(&actual); err != nil {
		t.Fatalf("failed to query migrated status for %s: %v", questID, err)
	}
	if actual != expected {
		t.Fatalf("expected quest %s status=%q, got %q", questID, expected, actual)
	}
}
