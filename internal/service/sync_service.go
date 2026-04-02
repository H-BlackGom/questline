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
	db    *sql.DB
	clock engine.Clock
}

func NewSyncService(repo *repository.Repository) SyncService {
	return newSyncServiceWithClock(repo, engine.SystemClock{})
}

func newSyncServiceWithClock(repo *repository.Repository, clock engine.Clock) *syncService {
	if clock == nil {
		clock = engine.SystemClock{}
	}
	return &syncService{db: repo.DB(), clock: clock}
}

func (s *syncService) EvaluateLazySync() (*SyncResult, error) {
	now := s.clock.Now()
	if now.IsZero() {
		return nil, fmt.Errorf("clock returned zero time")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start lazy sync transaction: %w", err)
	}
	defer tx.Rollback()

	var lastSyncedRaw string
	err = tx.QueryRow("SELECT last_synced_at FROM player WHERE id = 1").Scan(&lastSyncedRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load player sync checkpoint: %w", err)
	}
	lastSynced, ok := repositoryTime(lastSyncedRaw)
	if !ok {
		lastSynced = now
	}

	plan, err := engine.EvaluateLazySync(lastSynced, now, s.clock)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate lazy sync plan: %w", err)
	}

	if plan.EvaluationDate != nil {
		if err := s.evaluateDateIfNeeded(tx, *plan.EvaluationDate); err != nil {
			return nil, err
		}
	}

	if err := s.rolloverWeekly(tx, plan.WeeklyRolloverTo); err != nil {
		return nil, err
	}
	if err := s.archiveExpiredQuests(tx, plan.LogicalDate); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		"UPDATE player SET last_synced_at = ?, updated_at = datetime('now') WHERE id = 1",
		now.Format(time.RFC3339),
	); err != nil {
		return nil, fmt.Errorf("failed to persist lazy sync checkpoint: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit lazy sync transaction: %w", err)
	}

	return &SyncResult{EvaluatedAt: now}, nil
}

func (s *syncService) evaluateDateIfNeeded(tx *sql.Tx, date time.Time) error {
	dateKey := date.Format("2006-01-02")

	var exists int
	if err := tx.QueryRow("SELECT COUNT(*) FROM daily_evaluation WHERE date = ?", dateKey).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check existing daily evaluation: %w", err)
	}
	if exists > 0 {
		return nil
	}

	var total int
	var completed int
	err := tx.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0)
		 FROM quests
		 WHERE deleted_at IS NULL
		   AND type = ?
		   AND (scheduled_date = ? OR date(created_at) = ?)`,
		string(domain.StatusCompleted),
		string(domain.QuestTypeDaily),
		dateKey,
		dateKey,
	).Scan(&total, &completed)
	if err != nil {
		return fmt.Errorf("failed to evaluate daily flow stats: %w", err)
	}

	grade := engine.EvaluateFlowStatus(total, completed)
	rate := 0.0
	if total > 0 {
		rate = float64(completed) / float64(total)
	}

	if _, err := tx.Exec(
		`INSERT INTO daily_evaluation (date, total_routines, completed_routines, completion_rate, flow_grade)
		 VALUES (?, ?, ?, ?, ?)`,
		dateKey,
		total,
		completed,
		rate,
		string(grade),
	); err != nil {
		return fmt.Errorf("failed to persist daily evaluation: %w", err)
	}

	if _, err := tx.Exec(
		`UPDATE player
		 SET flow_status = ?, last_evaluated = ?, updated_at = datetime('now')
		 WHERE id = 1`,
		string(grade),
		date.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("failed to persist flow status after evaluation: %w", err)
	}

	return nil
}

func (s *syncService) rolloverWeekly(tx *sql.Tx, rolloverDate time.Time) error {
	dateKey := rolloverDate.Format("2006-01-02")
	if _, err := tx.Exec(
		`UPDATE quests
		 SET due_date = ?
		 WHERE deleted_at IS NULL
		   AND type = ?
		   AND status IN (?, ?)
		   AND due_date IS NOT NULL
		   AND date(due_date) < ?`,
		dateKey,
		string(domain.QuestTypeWeekly),
		string(domain.StatusPending),
		string(domain.StatusInProgress),
		dateKey,
	); err != nil {
		return fmt.Errorf("failed to rollover overdue weekly quests: %w", err)
	}
	return nil
}

func (s *syncService) archiveExpiredQuests(tx *sql.Tx, logicalDate time.Time) error {
	dateKey := logicalDate.Format("2006-01-02")
	if _, err := tx.Exec(
		`UPDATE quests
		 SET status = ?
		 WHERE deleted_at IS NULL
		   AND status IN (?, ?)
		   AND type IN (?, ?)
		   AND due_date IS NOT NULL
		   AND date(due_date) < ?`,
		string(domain.StatusArchived),
		string(domain.StatusPending),
		string(domain.StatusInProgress),
		string(domain.QuestTypeDaily),
		string(domain.QuestTypeWeekly),
		dateKey,
	); err != nil {
		return fmt.Errorf("failed to archive expired quests: %w", err)
	}
	return nil
}

func repositoryTime(value string) (time.Time, bool) {
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
