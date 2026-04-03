package repository

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

func TestQuestTreeQueries(t *testing.T) {
	repo := newTestRepository(t)
	defer repo.Close()

	questRepo := NewQuestRepository(repo)

	createdAt := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	daily := seedQuest(t, questRepo, "daily_root", domain.QuestTypeDaily, nil, createdAt)
	weekly := seedQuest(t, questRepo, "weekly_root", domain.QuestTypeWeekly, nil, createdAt.Add(1*time.Minute))
	epic := seedQuest(t, questRepo, "epic_root", domain.QuestTypeEpic, nil, createdAt.Add(2*time.Minute))
	guild := seedQuest(t, questRepo, "guild_root", domain.QuestTypeGuild, nil, createdAt.Add(3*time.Minute))
	_ = seedQuest(t, questRepo, "daily_child", domain.QuestTypeSub, &daily.ID, createdAt.Add(4*time.Minute))
	epicChild := seedQuest(t, questRepo, "epic_child", domain.QuestTypeSub, &epic.ID, createdAt.Add(5*time.Minute))
	deletedRoot := seedQuest(t, questRepo, "deleted_root", domain.QuestTypeDaily, nil, createdAt.Add(6*time.Minute))

	if err := questRepo.Delete(deletedRoot.ID); err != nil {
		t.Fatalf("failed to soft-delete quest fixture: %v", err)
	}

	roots, err := questRepo.GetRootQuests()
	if err != nil {
		t.Fatalf("GetRootQuests failed: %v", err)
	}
	if len(roots) != 4 {
		t.Fatalf("expected 4 visible roots, got %d", len(roots))
	}

	if roots[0].ID != daily.ID || roots[1].ID != weekly.ID || roots[2].ID != epic.ID || roots[3].ID != guild.ID {
		t.Fatalf("unexpected root ordering: got [%s %s %s %s]", roots[0].ID, roots[1].ID, roots[2].ID, roots[3].ID)
	}

	tree, err := questRepo.GetQuestTree()
	if err != nil {
		t.Fatalf("GetQuestTree failed: %v", err)
	}

	ids := make(map[string]struct{}, len(tree))
	for _, quest := range tree {
		ids[quest.ID] = struct{}{}
	}
	if _, exists := ids[deletedRoot.ID]; exists {
		t.Fatalf("soft-deleted quest should be excluded from tree")
	}
	if _, exists := ids[epicChild.ID]; !exists {
		t.Fatalf("epic child should be included in tree")
	}

	children, err := questRepo.GetByParent(epic.ID)
	if err != nil {
		t.Fatalf("GetByParent failed: %v", err)
	}
	if len(children) != 1 || children[0].ID != epicChild.ID {
		t.Fatalf("expected epic child from GetByParent, got %+v", children)
	}

	dailyQuests, err := questRepo.GetByType(domain.QuestTypeDaily)
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}
	if len(dailyQuests) != 1 || dailyQuests[0].ID != daily.ID {
		t.Fatalf("expected only active daily root from GetByType")
	}
}

func TestPlayerFlowPersistence(t *testing.T) {
	repo := newTestRepository(t)
	defer repo.Close()

	playerRepo := NewPlayerRepository(repo)

	syncAt := time.Date(2026, 4, 3, 9, 30, 0, 0, time.UTC)
	if err := playerRepo.UpdateFlowStatus(domain.FlowStatusBurning); err != nil {
		t.Fatalf("UpdateFlowStatus failed: %v", err)
	}
	if err := playerRepo.UpdateSyncTime(syncAt); err != nil {
		t.Fatalf("UpdateSyncTime failed: %v", err)
	}

	player, err := playerRepo.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if player.FlowStatus != domain.FlowStatusBurning {
		t.Fatalf("expected flow_status=burning, got %s", player.FlowStatus)
	}
	if player.LastSyncedAt == nil || !player.LastSyncedAt.Equal(syncAt) {
		t.Fatalf("expected last_synced_at=%s, got %+v", syncAt.Format(time.RFC3339), player.LastSyncedAt)
	}
	if player.LastEvaluated == nil || player.LastEvaluated.Format("2006-01-02") != time.Now().UTC().Format("2006-01-02") {
		t.Fatalf("expected last_evaluated set to today, got %+v", player.LastEvaluated)
	}

	var count int
	if err := repo.DB().QueryRow("SELECT COUNT(*) FROM player").Scan(&count); err != nil {
		t.Fatalf("failed to verify singleton player row: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected player singleton row count=1, got %d", count)
	}
}

func TestDailyEvaluationPersistence(t *testing.T) {
	repo := newTestRepository(t)
	defer repo.Close()

	questRepo := NewQuestRepository(repo)
	historyRepo := NewHistoryRepository(repo)

	createdAt := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	q1 := seedQuest(t, questRepo, "daily_1", domain.QuestTypeDaily, nil, createdAt)
	q2 := seedQuest(t, questRepo, "daily_2", domain.QuestTypeDaily, nil, createdAt.Add(time.Minute))

	date := "2026-04-03"
	if err := historyRepo.SaveQuestHistory(q1.ID, date, domain.StatusCompleted, true, 50); err != nil {
		t.Fatalf("SaveQuestHistory q1 failed: %v", err)
	}
	if err := historyRepo.SaveQuestHistory(q2.ID, date, domain.StatusPending, false, 0); err != nil {
		t.Fatalf("SaveQuestHistory q2 failed: %v", err)
	}

	stats, err := historyRepo.GetDailyStats(date)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}
	if stats.TotalRoutines != 2 || stats.CompletedRoutines != 1 || stats.TotalXPEarned != 50 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	if err := historyRepo.SaveDailyEvaluation(date, 2, 1, 0.5, domain.FlowStatusSmooth); err != nil {
		t.Fatalf("SaveDailyEvaluation initial failed: %v", err)
	}
	if err := historyRepo.SaveDailyEvaluation(date, 2, 2, 1.0, domain.FlowStatusSingularity); err != nil {
		t.Fatalf("SaveDailyEvaluation upsert failed: %v", err)
	}

	evaluations, err := historyRepo.GetEvaluations(10)
	if err != nil {
		t.Fatalf("GetEvaluations failed: %v", err)
	}
	if len(evaluations) != 1 {
		t.Fatalf("expected 1 evaluation row, got %d", len(evaluations))
	}
	if evaluations[0].FlowGrade != domain.FlowStatusSingularity || evaluations[0].CompletionRate != 1.0 {
		t.Fatalf("unexpected evaluation row: %+v", evaluations[0])
	}
}

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "repo-test.db")
	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test repository: %v", err)
	}
	return repo
}

func seedQuest(t *testing.T, questRepo *QuestRepository, id string, questType domain.QuestType, parentID *string, createdAt time.Time) *domain.Quest {
	t.Helper()

	quest := &domain.Quest{
		ID:        id,
		Title:     id,
		Status:    domain.StatusPending,
		Type:      questType,
		ParentID:  parentID,
		CreatedAt: createdAt,
	}
	if err := questRepo.Create(quest); err != nil {
		t.Fatalf("failed to seed quest %s: %v", id, err)
	}
	return quest
}
