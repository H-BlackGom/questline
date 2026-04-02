package service

import (
	"database/sql"
	"fmt"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
)

type playerService struct {
	playerRepo *repository.PlayerRepository
	db         *sql.DB
}

func NewPlayerService(playerRepo *repository.PlayerRepository, repo *repository.Repository) PlayerService {
	return &playerService{playerRepo: playerRepo, db: repo.DB()}
}

func (s *playerService) GetPlayer() (*domain.Player, error) {
	return s.playerRepo.Get()
}

func (s *playerService) AwardXP(baseXP int, _ string) (*AwardResult, error) {
	if baseXP <= 0 {
		player, err := s.playerRepo.Get()
		if err != nil {
			return nil, err
		}
		return &AwardResult{
			XPBefore:        player.CurrentXP,
			XPAfter:         player.CurrentXP,
			LevelBefore:     player.Level,
			LevelAfter:      player.Level,
			LevelUpOccurred: false,
		}, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start xp transaction: %w", err)
	}
	defer tx.Rollback()

	var level int
	var currentXP int
	var totalXPEarned int
	err = tx.QueryRow("SELECT level, current_xp, total_xp_earned FROM player WHERE id = 1").Scan(&level, &currentXP, &totalXPEarned)
	if err != nil {
		return nil, fmt.Errorf("failed to load player for xp award: %w", err)
	}

	leveling := engine.AddXP(level, currentXP, baseXP)
	if _, err := tx.Exec("UPDATE player SET level = ?, current_xp = ?, total_xp_earned = ?, updated_at = datetime('now') WHERE id = 1", leveling.NewLevel, leveling.NewXP, totalXPEarned+baseXP); err != nil {
		return nil, fmt.Errorf("failed to persist xp award: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit xp award transaction: %w", err)
	}

	return &AwardResult{
		XPBefore:        currentXP,
		XPAfter:         leveling.NewXP,
		LevelBefore:     level,
		LevelAfter:      leveling.NewLevel,
		LevelUpOccurred: leveling.LeveledUp,
	}, nil
}
