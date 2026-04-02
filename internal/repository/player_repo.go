package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

// PlayerRepository provides player CRUD operations
type PlayerRepository struct {
	repo *Repository
}

// NewPlayerRepository creates a new PlayerRepository
func NewPlayerRepository(repo *Repository) *PlayerRepository {
	return &PlayerRepository{repo: repo}
}

// Get retrieves the singleton player
func (pr *PlayerRepository) Get() (*domain.Player, error) {
	var player domain.Player
	var updatedAt string
	var flowStatus sql.NullString
	var lastSyncedAt sql.NullString
	var lastEvaluated sql.NullString
	err := pr.repo.db.QueryRow(
		`SELECT id, level, current_xp, total_xp_earned, quests_completed, updated_at,
		        flow_status, last_synced_at, last_evaluated, streak_days
		 FROM player WHERE id = 1`,
	).Scan(
		&player.ID,
		&player.Level,
		&player.CurrentXP,
		&player.TotalXPEarned,
		&player.QuestsCompleted,
		&updatedAt,
		&flowStatus,
		&lastSyncedAt,
		&lastEvaluated,
		&player.StreakDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if flowStatus.Valid {
		player.FlowStatus = domain.FlowStatus(flowStatus.String)
	}
	if t, ok := parseDBTime(updatedAt); ok {
		player.UpdatedAt = t
	}
	if lastSyncedAt.Valid {
		if t, ok := parseDBTime(lastSyncedAt.String); ok {
			player.LastSyncedAt = &t
		}
	}
	if lastEvaluated.Valid {
		if t, ok := parseDBTime(lastEvaluated.String); ok {
			player.LastEvaluated = &t
		}
	}

	return &player, nil
}

// Update updates the player
func (pr *PlayerRepository) Update(player *domain.Player) error {
	var lastSynced interface{}
	if player.LastSyncedAt != nil {
		lastSynced = player.LastSyncedAt.Format(time.RFC3339)
	}
	var lastEvaluated interface{}
	if player.LastEvaluated != nil {
		lastEvaluated = player.LastEvaluated.Format(time.RFC3339)
	}

	result, err := pr.repo.db.Exec(
		`UPDATE player SET level = ?, current_xp = ?, total_xp_earned = ?, 
		 quests_completed = ?, flow_status = ?, last_synced_at = ?, last_evaluated = ?,
		 streak_days = ?, updated_at = datetime('now') WHERE id = 1`,
		player.Level,
		player.CurrentXP,
		player.TotalXPEarned,
		player.QuestsCompleted,
		string(player.FlowStatus),
		lastSynced,
		lastEvaluated,
		player.StreakDays,
	)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	if err := assertSinglePlayerUpdated(result); err != nil {
		return err
	}
	return nil
}

func (pr *PlayerRepository) UpdateFlowStatus(status domain.FlowStatus) error {
	result, err := pr.repo.db.Exec(
		`UPDATE player SET flow_status = ?, last_evaluated = ?, updated_at = datetime('now') WHERE id = 1`,
		string(status),
		time.Now().UTC().Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("failed to update player flow status: %w", err)
	}
	if err := assertSinglePlayerUpdated(result); err != nil {
		return err
	}
	return nil
}

func (pr *PlayerRepository) UpdateSyncTime(syncedAt time.Time) error {
	result, err := pr.repo.db.Exec(
		`UPDATE player SET last_synced_at = ?, updated_at = datetime('now') WHERE id = 1`,
		syncedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to update player sync time: %w", err)
	}
	if err := assertSinglePlayerUpdated(result); err != nil {
		return err
	}
	return nil
}

func assertSinglePlayerUpdated(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect player update rows: %w", err)
	}
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func parseDBTime(value string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return t.UTC(), true
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}
