package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
)

type playerService struct {
	playerRepo *repository.PlayerRepository
	db         *sql.DB
}

type PlayerWithFlow struct {
	*domain.Player
	CurrentFlow    domain.FlowStatus
	FlowMultiplier float64
	NextEvaluation time.Time
}

func NewPlayerService(playerRepo *repository.PlayerRepository, repo *repository.Repository) PlayerService {
	return &playerService{playerRepo: playerRepo, db: repo.DB()}
}

func (s *playerService) GetPlayer() (*domain.Player, error) {
	return s.playerRepo.Get()
}

func (s *playerService) GetPlayerWithFlow() (*PlayerWithFlow, error) {
	player, err := s.playerRepo.Get()
	if err != nil {
		return nil, err
	}

	flowStatus := player.FlowStatus
	if !flowStatus.IsValid() {
		flowStatus = domain.FlowStatusSmooth
	}

	return &PlayerWithFlow{
		Player:         player,
		CurrentFlow:    flowStatus,
		FlowMultiplier: flowMultiplierForStatus(flowStatus),
		NextEvaluation: nextFlowEvaluation(time.Now()),
	}, nil

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

func flowMultiplierForStatus(status domain.FlowStatus) float64 {
	switch status {
	case domain.FlowStatusSingularity:
		return 2.0
	case domain.FlowStatusBurning:
		return 1.5
	case domain.FlowStatusHazy:
		return 0.5
	default:
		return 1.0
	}
}

func nextFlowEvaluation(now time.Time) time.Time {
	locNow := now.Local()
	next := time.Date(locNow.Year(), locNow.Month(), locNow.Day(), 4, 0, 0, 0, locNow.Location())
	if !locNow.Before(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
