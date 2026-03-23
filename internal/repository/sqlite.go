package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Repository handles database operations
type Repository struct {
	db *sql.DB
}

// New creates a new Repository instance
func New(dbPath string) (*Repository, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &Repository{db: db}
	if err := repo.initSchema(); err != nil {
		return nil, err
	}

	return repo, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// initSchema creates tables if they don't exist
func (r *Repository) initSchema() error {
	// Quests table
	questsTable := `
	CREATE TABLE IF NOT EXISTS quests (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		status TEXT DEFAULT 'TODO' CHECK (status IN ('TODO', 'DONE', 'DROPPED')),
		due_date TEXT,
		created_at TEXT NOT NULL,
		completed_at TEXT
	);`
	if _, err := r.db.Exec(questsTable); err != nil {
		return fmt.Errorf("failed to create quests table: %w", err)
	}

	// Player table (singleton - always 1 row)
	playerTable := `
	CREATE TABLE IF NOT EXISTS player (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		level INTEGER DEFAULT 1 NOT NULL,
		current_xp INTEGER DEFAULT 0 NOT NULL,
		total_xp_earned INTEGER DEFAULT 0 NOT NULL,
		quests_completed INTEGER DEFAULT 0 NOT NULL,
		updated_at TEXT NOT NULL
	);`
	if _, err := r.db.Exec(playerTable); err != nil {
		return fmt.Errorf("failed to create player table: %w", err)
	}

	// Ensure singleton player exists
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM player").Scan(&count); err != nil {
		return fmt.Errorf("failed to check player: %w", err)
	}
	if count == 0 {
		_, err := r.db.Exec("INSERT INTO player (id, level, current_xp, total_xp_earned, quests_completed, updated_at) VALUES (1, 1, 0, 0, 0, datetime('now'))")
		if err != nil {
			return fmt.Errorf("failed to create initial player: %w", err)
		}
	}

	return nil
}

// DB returns the underlying database connection
func (r *Repository) DB() *sql.DB {
	return r.db
}
