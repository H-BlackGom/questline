package repository

import (
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
	var flowStatus string
	var lastSyncedAt string
	var lastEvaluated string
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

	player.FlowStatus = domain.FlowStatus(flowStatus)
	if t, ok := parseDBTime(updatedAt); ok {
		player.UpdatedAt = t
	}
	if t, ok := parseDBTime(lastSyncedAt); ok {
		player.LastSyncedAt = &t
	}
	if t, ok := parseDBTime(lastEvaluated); ok {
		player.LastEvaluated = &t
	}

	return &player, nil
}

// Update updates the player
func (pr *PlayerRepository) Update(player *domain.Player) error {
	lastSynced := ""
	if player.LastSyncedAt != nil {
		lastSynced = player.LastSyncedAt.Format(time.RFC3339)
	} else {
		lastSynced = time.Now().UTC().Format(time.RFC3339)
	}
	lastEvaluated := ""
	if player.LastEvaluated != nil {
		lastEvaluated = player.LastEvaluated.Format(time.RFC3339)
	} else {
		lastEvaluated = time.Now().UTC().Format(time.RFC3339)
	}

	_, err := pr.repo.db.Exec(
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
	return nil
}

func parseDBTime(value string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}
