package repository

import (
	"database/sql"
	"fmt"

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
	err := pr.repo.db.QueryRow(
		`SELECT id, level, current_xp, total_xp_earned, quests_completed, updated_at 
		 FROM player WHERE id = 1`,
	).Scan(
		&player.ID,
		&player.Level,
		&player.CurrentXP,
		&player.TotalXPEarned,
		&player.QuestsCompleted,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	_ = updatedAt
	return &player, nil
}

// Update updates the player
func (pr *PlayerRepository) Update(player *domain.Player) error {
	_, err := pr.repo.db.Exec(
		`UPDATE player SET level = ?, current_xp = ?, total_xp_earned = ?, 
		 quests_completed = ?, updated_at = datetime('now') WHERE id = 1`,
		player.Level,
		player.CurrentXP,
		player.TotalXPEarned,
		player.QuestsCompleted,
	)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	return nil
}

// CompleteQuest marks a quest as done and awards XP in a transaction
func (pr *PlayerRepository) CompleteQuest(questID string, xpAmount int) error {
	tx, err := pr.repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if quest exists and is not already done
	var status string
	err = tx.QueryRow("SELECT status FROM quests WHERE id = ?", questID).Scan(&status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("quest not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check quest: %w", err)
	}
	if status == "DONE" {
		return fmt.Errorf("quest already completed")
	}

	// Mark quest as done
	_, err = tx.Exec(
		"UPDATE quests SET status = 'DONE', completed_at = datetime('now') WHERE id = ?",
		questID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete quest: %w", err)
	}

	// Get player
	var player domain.Player
	err = tx.QueryRow(
		"SELECT level, current_xp, total_xp_earned, quests_completed FROM player WHERE id = 1",
	).Scan(&player.Level, &player.CurrentXP, &player.TotalXPEarned, &player.QuestsCompleted)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Update player XP
	player.CurrentXP += xpAmount
	player.TotalXPEarned += xpAmount
	player.QuestsCompleted++

	// Check for level up
	requiredXP := 100 + (player.Level * 50)
	for player.CurrentXP >= requiredXP {
		player.CurrentXP -= requiredXP
		player.Level++
		requiredXP = 100 + (player.Level * 50)
	}

	// Update player
	_, err = tx.Exec(
		"UPDATE player SET level = ?, current_xp = ?, total_xp_earned = ?, quests_completed = ?, updated_at = datetime('now') WHERE id = 1",
		player.Level, player.CurrentXP, player.TotalXPEarned, player.QuestsCompleted,
	)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return tx.Commit()
}
