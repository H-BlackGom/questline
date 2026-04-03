package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	schemaVersionV1 = 1
	schemaVersionV2 = 2
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

func (r *Repository) initSchema() error {
	version, err := r.getUserVersion()
	if err != nil {
		return err
	}

	switch version {
	case schemaVersionV2:
		if err := r.ensurePlayerSingleton(); err != nil {
			return err
		}
		return r.ensureV2SupportTables()
	case schemaVersionV1:
		return r.migrateV1ToV2()
	case 0:
		hasQuests, err := r.hasTable("quests")
		if err != nil {
			return err
		}
		hasPlayer, err := r.hasTable("player")
		if err != nil {
			return err
		}

		if hasQuests || hasPlayer {
			return r.migrateV1ToV2()
		}
		return r.createFreshV2Schema()
	default:
		return fmt.Errorf("unsupported schema version: %d", version)
	}
}

func (r *Repository) getUserVersion() (int, error) {
	var version int
	if err := r.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("failed to read user_version: %w", err)
	}
	return version, nil
}

func (r *Repository) hasTable(name string) (bool, error) {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to inspect schema tables: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) createFreshV2Schema() (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin schema transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = createQuestsV2Table(tx, "quests"); err != nil {
		return err
	}
	if err = createPlayerV2Table(tx, "player"); err != nil {
		return err
	}
	if err = createQuestHistoryV2Table(tx, "quest_history"); err != nil {
		return err
	}
	if err = createDailyEvaluationV2Table(tx, "daily_evaluation"); err != nil {
		return err
	}
	if err = ensurePlayerSingletonTx(tx); err != nil {
		return err
	}
	if _, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersionV2)); err != nil {
		return fmt.Errorf("failed to set user_version to v2: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit v2 schema bootstrap: %w", err)
	}
	return nil
}

func (r *Repository) migrateV1ToV2() (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = createQuestsV2Table(tx, "quests_new"); err != nil {
		return err
	}
	if err = createPlayerV2Table(tx, "player_new"); err != nil {
		return err
	}
	if err = createQuestHistoryV2Table(tx, "quest_history"); err != nil {
		return err
	}
	if err = createDailyEvaluationV2Table(tx, "daily_evaluation"); err != nil {
		return err
	}

	if _, err = tx.Exec(`
		INSERT INTO quests_new (
			id, title, status, due_date, created_at, completed_at,
			type, parent_id, scheduled_date, deleted_at
		)
		SELECT
			id,
			title,
			CASE status
				WHEN 'TODO' THEN 'pending'
				WHEN 'DONE' THEN 'completed'
				WHEN 'DROPPED' THEN 'archived'
				ELSE 'pending'
			END,
			due_date,
			created_at,
			completed_at,
			'daily',
			NULL,
			NULL,
			NULL
		FROM quests;
	`); err != nil {
		return fmt.Errorf("failed to migrate quests data: %w", err)
	}

	if _, err = tx.Exec(`
		INSERT INTO player_new (
			id, level, current_xp, total_xp_earned, quests_completed, updated_at,
			flow_status, last_synced_at, last_evaluated, streak_days
		)
		SELECT
			id,
			level,
			current_xp,
			total_xp_earned,
			quests_completed,
			updated_at,
			'smooth',
			updated_at,
			updated_at,
			0
		FROM player;
	`); err != nil {
		return fmt.Errorf("failed to migrate player data: %w", err)
	}

	if _, err = tx.Exec("ALTER TABLE quests RENAME TO quests_v1_backup"); err != nil {
		return fmt.Errorf("failed to rename legacy quests table: %w", err)
	}
	if _, err = tx.Exec("ALTER TABLE quests_new RENAME TO quests"); err != nil {
		return fmt.Errorf("failed to activate migrated quests table: %w", err)
	}
	if _, err = tx.Exec("DROP TABLE quests_v1_backup"); err != nil {
		return fmt.Errorf("failed to remove legacy quests backup: %w", err)
	}

	if _, err = tx.Exec("ALTER TABLE player RENAME TO player_v1_backup"); err != nil {
		return fmt.Errorf("failed to rename legacy player table: %w", err)
	}
	if _, err = tx.Exec("ALTER TABLE player_new RENAME TO player"); err != nil {
		return fmt.Errorf("failed to activate migrated player table: %w", err)
	}
	if _, err = tx.Exec("DROP TABLE player_v1_backup"); err != nil {
		return fmt.Errorf("failed to remove legacy player backup: %w", err)
	}

	if err = ensurePlayerSingletonTx(tx); err != nil {
		return err
	}

	if _, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersionV2)); err != nil {
		return fmt.Errorf("failed to set user_version to v2: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit v1→v2 migration: %w", err)
	}
	return nil
}

func createQuestsV2Table(tx *sql.Tx, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'pending_completion', 'completed', 'archived')),
			due_date TEXT,
			created_at TEXT NOT NULL,
			completed_at TEXT,
			type TEXT NOT NULL DEFAULT 'daily' CHECK (type IN ('daily', 'weekly', 'epic', 'guild', 'sub')),
			parent_id TEXT,
			scheduled_date TEXT,
			deleted_at TEXT,
			FOREIGN KEY (parent_id) REFERENCES quests(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_quests_type ON %s(type);
		CREATE INDEX IF NOT EXISTS idx_quests_status ON %s(status);
		CREATE INDEX IF NOT EXISTS idx_quests_parent ON %s(parent_id);
		CREATE INDEX IF NOT EXISTS idx_quests_scheduled ON %s(scheduled_date);
		CREATE INDEX IF NOT EXISTS idx_quests_deleted ON %s(deleted_at);
	`, tableName, tableName, tableName, tableName, tableName, tableName)

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to create %s table: %w", tableName, err)
	}
	return nil
}

func createPlayerV2Table(tx *sql.Tx, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			level INTEGER DEFAULT 1 NOT NULL,
			current_xp INTEGER DEFAULT 0 NOT NULL,
			total_xp_earned INTEGER DEFAULT 0 NOT NULL,
			quests_completed INTEGER DEFAULT 0 NOT NULL,
			updated_at TEXT NOT NULL,
			flow_status TEXT NOT NULL DEFAULT 'smooth' CHECK (flow_status IN ('singularity', 'burning', 'smooth', 'hazy')),
			last_synced_at TEXT NOT NULL,
			last_evaluated TEXT NOT NULL,
			streak_days INTEGER NOT NULL DEFAULT 0
		);
	`, tableName)

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to create %s table: %w", tableName, err)
	}
	return nil
}

func createQuestHistoryV2Table(tx *sql.Tx, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quest_id TEXT NOT NULL,
			date TEXT NOT NULL,
			status TEXT NOT NULL,
			completed BOOLEAN DEFAULT FALSE,
			xp_earned INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (quest_id) REFERENCES quests(id) ON DELETE CASCADE,
			UNIQUE (quest_id, date)
		);
		CREATE INDEX IF NOT EXISTS idx_history_quest ON %s(quest_id);
		CREATE INDEX IF NOT EXISTS idx_history_date ON %s(date);
	`, tableName, tableName, tableName)

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to create %s table: %w", tableName, err)
	}
	return nil
}

func createDailyEvaluationV2Table(tx *sql.Tx, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL UNIQUE,
			total_routines INTEGER NOT NULL,
			completed_routines INTEGER NOT NULL,
			completion_rate REAL NOT NULL,
			flow_grade TEXT NOT NULL CHECK (flow_grade IN ('singularity', 'burning', 'smooth', 'hazy')),
			evaluated_at TEXT DEFAULT (datetime('now'))
		);
		CREATE INDEX IF NOT EXISTS idx_evaluation_date ON %s(date);
	`, tableName, tableName)

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to create %s table: %w", tableName, err)
	}
	return nil
}

func (r *Repository) ensurePlayerSingleton() error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin player singleton transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensurePlayerSingletonTx(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit player singleton transaction: %w", err)
	}
	return nil
}

func ensurePlayerSingletonTx(tx *sql.Tx) error {
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM player").Scan(&count); err != nil {
		return fmt.Errorf("failed to check player: %w", err)
	}
	if count == 0 {
		if _, err := tx.Exec(`
			INSERT INTO player (
				id, level, current_xp, total_xp_earned, quests_completed, updated_at,
				flow_status, last_synced_at, last_evaluated, streak_days
			) VALUES (
				1, 1, 0, 0, 0, datetime('now'),
				'smooth', datetime('now'), datetime('now'), 0
			)
		`); err != nil {
			return fmt.Errorf("failed to create initial player: %w", err)
		}
	}
	return nil
}

func (r *Repository) ensureV2SupportTables() error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin support-table transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := createQuestHistoryV2Table(tx, "quest_history"); err != nil {
		return err
	}
	if err := createDailyEvaluationV2Table(tx, "daily_evaluation"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit support-table transaction: %w", err)
	}
	return nil
}

// DB returns the underlying database connection
func (r *Repository) DB() *sql.DB {
	return r.db
}
