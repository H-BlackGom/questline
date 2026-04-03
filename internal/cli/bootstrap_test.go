package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/service"
)

type fakeQuestService struct{}

func (f *fakeQuestService) CreateQuest(title string, questType domain.QuestType, parentID *string, dueDate *time.Time) (*domain.Quest, error) {
	return &domain.Quest{ID: "q1234567", Title: title, Type: questType, ParentID: parentID, DueDate: dueDate, Status: domain.StatusPending}, nil
}

func (f *fakeQuestService) CompleteQuest(questID string) (*service.CompletionResult, error) {
	return &service.CompletionResult{XPAfter: 50, LevelAfter: 1, LevelBefore: 1}, nil
}

func (f *fakeQuestService) UpdateQuestStatus(questID string, newStatus domain.QuestStatus) error {
	return nil
}

func (f *fakeQuestService) GetQuest(questID string) (*domain.Quest, error) {
	return &domain.Quest{ID: questID, Status: domain.StatusPending}, nil
}

func (f *fakeQuestService) ListQuests(filter service.QuestFilter) ([]*domain.Quest, error) {
	return []*domain.Quest{{ID: "q1234567", Title: "stub", Status: domain.StatusPending}}, nil
}

func (f *fakeQuestService) GetQuestTree() ([]*domain.Quest, error) {
	return []*domain.Quest{}, nil
}

type fakePlayerService struct{}

func (f *fakePlayerService) GetPlayer() (*domain.Player, error) {
	return &domain.Player{Level: 1, CurrentXP: 0, TotalXPEarned: 0, QuestsCompleted: 0}, nil
}

func (f *fakePlayerService) GetPlayerWithFlow() (*service.PlayerWithFlow, error) {
	player, err := f.GetPlayer()
	if err != nil {
		return nil, err
	}
	return &service.PlayerWithFlow{Player: player, CurrentFlow: domain.FlowStatusSmooth, FlowMultiplier: 1.0, NextEvaluation: time.Now().Add(24 * time.Hour)}, nil
}

func (f *fakePlayerService) AwardXP(baseXP int, questID string) (*service.AwardResult, error) {
	return &service.AwardResult{}, nil
}

type fakeSyncService struct{}

func (f *fakeSyncService) EvaluateLazySync() (*service.SyncResult, error) {
	return &service.SyncResult{EvaluatedAt: time.Now()}, nil
}

func (f *fakeSyncService) CalculateFlowGrade(date time.Time) (*service.FlowGrade, error) {
	return &service.FlowGrade{Date: date, Grade: domain.FlowStatusSmooth, Multiplier: 1.0}, nil
}

func TestCommandsUseSharedBootstrap(t *testing.T) {
	originalBootstrap := bootstrap
	defer func() { bootstrap = originalBootstrap }()

	calls := 0
	bootstrap = func(_ string) (*service.Services, error) {
		calls++
		return &service.Services{
			Quest:  &fakeQuestService{},
			Player: &fakePlayerService{},
			Sync:   &fakeSyncService{},
		}, nil
	}

	commands := [][]string{
		{"add", "shared bootstrap"},
		{"done", "q1234567"},
		{"ls"},
		{"me"},
	}

	for _, args := range commands {
		dueDate = ""
		listAll = false
		listDone = false
		listType = ""
		addQuestType = "daily"
		addParentID = ""
		meFlow = false

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("command %v failed: %v", args, err)
		}
	}

	if calls != len(commands) {
		t.Fatalf("expected bootstrap to be called %d times, got %d", len(commands), calls)
	}
}
