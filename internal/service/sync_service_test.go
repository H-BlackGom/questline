package service

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
)

func TestFlowGrade(t *testing.T) {
	repo := newSyncTestRepository(t)
	defer repo.Close()

	date := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	dateKey := date.Format("2006-01-02")

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, type, scheduled_date)
		 VALUES
		 ('daily_done', 'daily_done', ?, ?, ?, ?),
		 ('daily_pending', 'daily_pending', ?, ?, ?, ?)`,
		string(domain.StatusCompleted), date.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
		string(domain.StatusPending), date.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
	); err != nil {
		t.Fatalf("failed to seed daily quests: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(date))
	grade, err := svc.CalculateFlowGrade(date)
	if err != nil {
		t.Fatalf("CalculateFlowGrade failed: %v", err)
	}
	if grade.Grade != domain.FlowStatusSmooth {
		t.Fatalf("expected smooth grade, got %s", grade.Grade)
	}
	if grade.Multiplier != 1.0 {
		t.Fatalf("expected smooth multiplier 1.0, got %v", grade.Multiplier)
	}
}

func TestFlowGrade_ExcludesSoftDeletedDailyQuests(t *testing.T) {
	repo := newSyncTestRepository(t)
	defer repo.Close()

	date := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	dateKey := date.Format("2006-01-02")

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, type, scheduled_date, deleted_at)
		 VALUES
		 ('daily_active', 'daily_active', ?, ?, ?, ?, NULL),
		 ('daily_deleted', 'daily_deleted', ?, ?, ?, ?, datetime('now'))`,
		string(domain.StatusCompleted), date.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
		string(domain.StatusPending), date.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
	); err != nil {
		t.Fatalf("failed to seed daily quests with deleted row: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(date))
	grade, err := svc.CalculateFlowGrade(date)
	if err != nil {
		t.Fatalf("CalculateFlowGrade failed: %v", err)
	}
	if grade.Grade != domain.FlowStatusSingularity {
		t.Fatalf("expected singularity grade with only active daily counted, got %s", grade.Grade)
	}
	if grade.Multiplier != 2.0 {
		t.Fatalf("expected singularity multiplier 2.0, got %v", grade.Multiplier)
	}
}

func TestEvaluateLazySync(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)
	now := time.Date(2026, 4, 5, 5, 0, 0, 0, loc)
	lastSynced := time.Date(2026, 4, 3, 10, 0, 0, 0, loc)
	dateKey := "2026-04-03"

	repo := newSyncTestRepository(t)
	defer repo.Close()

	if _, err := repo.DB().Exec(
		"UPDATE player SET last_synced_at = ?, last_evaluated = ? WHERE id = 1",
		lastSynced.Format(time.RFC3339),
		lastSynced.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("failed to set last_synced_at fixture: %v", err)
	}

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, type, scheduled_date)
		 VALUES ('daily_old', 'daily_old', ?, ?, ?, ?)`,
		string(domain.StatusPending),
		lastSynced.Format(time.RFC3339),
		string(domain.QuestTypeDaily),
		dateKey,
	); err != nil {
		t.Fatalf("failed to seed stale daily quest: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(now))
	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("EvaluateLazySync failed: %v", err)
	}

	var count int
	if err := repo.DB().QueryRow("SELECT COUNT(*) FROM daily_evaluation WHERE date = ?", dateKey).Scan(&count); err != nil {
		t.Fatalf("failed to count daily evaluations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one daily evaluation for stale date, got %d", count)
	}
}

func TestEvaluateLazySync_LongInactivitySinglePenalty(t *testing.T) {
	now := time.Date(2026, 4, 8, 8, 0, 0, 0, time.UTC)
	lastSynced := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)
	dateKey := "2026-04-03"

	repo := newSyncTestRepository(t)
	defer repo.Close()

	if _, err := repo.DB().Exec(
		"UPDATE player SET flow_status = ?, last_synced_at = ?, last_evaluated = ? WHERE id = 1",
		string(domain.FlowStatusBurning),
		lastSynced.Format(time.RFC3339),
		lastSynced.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("failed to set player fixture: %v", err)
	}

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, type, scheduled_date)
		 VALUES
		 ('hist_done', 'hist_done', ?, ?, ?, ?),
		 ('hist_pending', 'hist_pending', ?, ?, ?, ?)`,
		string(domain.StatusCompleted), lastSynced.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
		string(domain.StatusPending), lastSynced.Format(time.RFC3339), string(domain.QuestTypeDaily), dateKey,
	); err != nil {
		t.Fatalf("failed to seed quests for history: %v", err)
	}

	if _, err := repo.DB().Exec(
		`INSERT INTO quest_history (quest_id, date, status, completed, xp_earned)
		 VALUES
		 ('hist_done', ?, ?, 1, 50),
		 ('hist_pending', ?, ?, 0, 0)`,
		dateKey,
		string(domain.StatusCompleted),
		dateKey,
		string(domain.StatusPending),
	); err != nil {
		t.Fatalf("failed to seed quest history: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(now))
	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("first EvaluateLazySync failed: %v", err)
	}

	flowStatus := mustString(t, repo, "SELECT flow_status FROM player WHERE id = 1")
	if flowStatus != string(domain.FlowStatusHazy) {
		t.Fatalf("expected single inactivity penalty to downgrade to hazy, got %s", flowStatus)
	}

	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("second EvaluateLazySync failed: %v", err)
	}

	flowStatusAfterSecondSync := mustString(t, repo, "SELECT flow_status FROM player WHERE id = 1")
	if flowStatusAfterSecondSync != flowStatus {
		t.Fatalf("expected single penalty behavior to be idempotent, first=%s second=%s", flowStatus, flowStatusAfterSecondSync)
	}
}

func TestPendingPreserved(t *testing.T) {
	now := time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC)
	lastSynced := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)

	repo := newSyncTestRepository(t)
	defer repo.Close()

	if _, err := repo.DB().Exec(
		"UPDATE player SET last_synced_at = ?, last_evaluated = ? WHERE id = 1",
		lastSynced.Format(time.RFC3339),
		lastSynced.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("failed to set last_synced_at fixture: %v", err)
	}

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, due_date, type)
		 VALUES ('parent_pending_completion', 'parent', ?, ?, ?, ?)`,
		string(domain.StatusPendingCompletion),
		lastSynced.Format(time.RFC3339),
		"2026-04-01",
		string(domain.QuestTypeWeekly),
	); err != nil {
		t.Fatalf("failed to seed pending completion quest: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(now))
	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("EvaluateLazySync failed: %v", err)
	}

	var status string
	if err := repo.DB().QueryRow("SELECT status FROM quests WHERE id = 'parent_pending_completion'").Scan(&status); err != nil {
		t.Fatalf("failed to load parent status: %v", err)
	}
	if status != string(domain.StatusPendingCompletion) {
		t.Fatalf("pending_completion should be preserved, got %s", status)
	}
}

func TestEvaluateLazySync_Idempotent(t *testing.T) {
	now := time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC)
	lastSynced := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	repo := newSyncTestRepository(t)
	defer repo.Close()

	if _, err := repo.DB().Exec(
		"UPDATE player SET last_synced_at = ?, last_evaluated = ? WHERE id = 1",
		lastSynced.Format(time.RFC3339),
		lastSynced.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("failed to set last_synced_at fixture: %v", err)
	}

	if _, err := repo.DB().Exec(
		`INSERT INTO quests (id, title, status, created_at, due_date, type, scheduled_date)
		 VALUES
		 ('weekly_old', 'weekly_old', ?, ?, ?, ?, ?),
		 ('daily_stale', 'daily_stale', ?, ?, NULL, ?, ?)`,
		string(domain.StatusPending), lastSynced.Format(time.RFC3339), "2026-04-01", string(domain.QuestTypeWeekly), "2026-04-03",
		string(domain.StatusPending), lastSynced.Format(time.RFC3339), string(domain.QuestTypeDaily), "2026-04-03",
	); err != nil {
		t.Fatalf("failed to seed idempotent fixture: %v", err)
	}

	svc := newSyncServiceWithClock(repo, engine.NewFixedClock(now))
	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("first EvaluateLazySync failed: %v", err)
	}

	firstEvalCount := mustCount(t, repo, "SELECT COUNT(*) FROM daily_evaluation")
	firstWeeklyDue := mustString(t, repo, "SELECT due_date FROM quests WHERE id = 'weekly_old'")

	if _, err := svc.EvaluateLazySync(); err != nil {
		t.Fatalf("second EvaluateLazySync failed: %v", err)
	}

	secondEvalCount := mustCount(t, repo, "SELECT COUNT(*) FROM daily_evaluation")
	secondWeeklyDue := mustString(t, repo, "SELECT due_date FROM quests WHERE id = 'weekly_old'")

	if secondEvalCount != firstEvalCount {
		t.Fatalf("expected idempotent daily evaluation count, first=%d second=%d", firstEvalCount, secondEvalCount)
	}
	if secondWeeklyDue != firstWeeklyDue {
		t.Fatalf("expected idempotent weekly rollover, first=%s second=%s", firstWeeklyDue, secondWeeklyDue)
	}
}

func newSyncTestRepository(t *testing.T) *repository.Repository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "sync-service-test.db")
	repo, err := repository.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	return repo
}

func mustCount(t *testing.T, repo *repository.Repository, query string, args ...any) int {
	t.Helper()
	var count int
	if err := repo.DB().QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("failed to execute count query %q: %v", query, err)
	}
	return count
}

func mustString(t *testing.T, repo *repository.Repository, query string, args ...any) string {
	t.Helper()
	var value string
	if err := repo.DB().QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("failed to execute string query %q: %v", query, err)
	}
	return value
}
