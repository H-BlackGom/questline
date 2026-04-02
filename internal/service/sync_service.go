package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
)

type syncService struct {
	db *sql.DB
}

func NewSyncService(repo *repository.Repository) SyncService {
	return &syncService{db: repo.DB()}
}

func (s *syncService) EvaluateLazySync() (*SyncResult, error) {
	now := time.Now().UTC()
	if _, err := s.db.Exec(
		"UPDATE player SET last_synced_at = ?, last_evaluated = ?, updated_at = datetime('now') WHERE id = 1",
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		return nil, fmt.Errorf("failed to persist lazy sync checkpoint: %w", err)
	}

	return &SyncResult{EvaluatedAt: now}, nil
}

func (s *syncService) CalculateFlowGrade(date time.Time) (*FlowGrade, error) {
	dateKey := date.Format("2006-01-02")

	var total int
	var completed int
	err := s.db.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0)
		 FROM quests
		 WHERE type = ? AND (scheduled_date = ? OR date(created_at) = ?)`,
		string(domain.StatusCompleted),
		string(domain.QuestTypeDaily),
		dateKey,
		dateKey,
	).Scan(&total, &completed)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate flow grade: %w", err)
	}

	rate := 0.0
	if total > 0 {
		rate = float64(completed) / float64(total)
	}

	grade := engine.EvaluateFlowStatus(total, completed)
	multiplier := 1.0
	switch grade {
	case domain.FlowStatusBurning:
		multiplier = 1.5
	case domain.FlowStatusHazy:
		multiplier = 0.5
	}

	return &FlowGrade{
		Date:           date,
		CompletionRate: rate,
		Grade:          grade,
		Multiplier:     multiplier,
	}, nil
}
