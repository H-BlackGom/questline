package tui

import (
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeTUIQuestService struct {
	quests       []*domain.Quest
	completedIDs []string
	err          error
}

func (f *fakeTUIQuestService) GetQuestTree() ([]*domain.Quest, error) {
	return f.quests, nil
}

func (f *fakeTUIQuestService) CompleteQuest(questID string) (*service.CompletionResult, error) {
	f.completedIDs = append(f.completedIDs, questID)
	if f.err != nil {
		return nil, f.err
	}
	return &service.CompletionResult{Quest: &domain.Quest{ID: questID}}, nil
}

func TestUpdateNavigation(t *testing.T) {
	now := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)
	daily := newRootQuest("daily-1", "아침 루틴", domain.QuestTypeDaily, now)
	epic := newRootQuest(
		"epic-1",
		"장기 프로젝트",
		domain.QuestTypeEpic,
		now.Add(time.Hour),
		newSubQuest("sub-1", "설계 확정", now.Add(2*time.Hour)),
		newSubQuest("sub-2", "리뷰 반영", now.Add(3*time.Hour)),
	)

	model := NewModel([]*QuestNode{daily, epic}, nil)
	questService := &fakeTUIQuestService{}
	model.questService = questService

	model, _ = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeyDown})
	if model.MasterCursor != 1 {
		t.Fatalf("expected master cursor to move down, got %d", model.MasterCursor)
	}

	model, _ = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.CurrentFocus != FocusDetail {
		t.Fatalf("expected focus to move to detail, got %v", model.CurrentFocus)
	}
	if model.DetailCursor != 0 {
		t.Fatalf("expected detail cursor to reset on enter, got %d", model.DetailCursor)
	}

	model, _ = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.DetailCursor != 1 {
		t.Fatalf("expected detail cursor to move down, got %d", model.DetailCursor)
	}

	model, cmd := updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("expected space on sub quest to return toggle command")
	}
	msg := cmd()
	toggled, ok := msg.(QuestToggledMsg)
	if !ok {
		t.Fatalf("expected QuestToggledMsg, got %T", msg)
	}
	if toggled.QuestID != "sub-2" {
		t.Fatalf("expected sub quest toggle for sub-2, got %q", toggled.QuestID)
	}
	if len(questService.completedIDs) != 1 || questService.completedIDs[0] != "sub-2" {
		t.Fatalf("expected sub quest completion path to use service CompleteQuest, got %v", questService.completedIDs)
	}
	if !model.Loading {
		t.Fatal("expected model to enter loading state when toggling a sub quest")
	}

	model, _ = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if model.CurrentFocus != FocusMaster {
		t.Fatalf("expected esc to return focus to master, got %v", model.CurrentFocus)
	}

	model, _ = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeyUp})
	if model.MasterCursor != 0 {
		t.Fatalf("expected master cursor to move back up, got %d", model.MasterCursor)
	}

	model, cmd = updateModelForTest(t, model, tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("expected space on root quest to return toggle command")
	}
	msg = cmd()
	toggled, ok = msg.(QuestToggledMsg)
	if !ok {
		t.Fatalf("expected QuestToggledMsg, got %T", msg)
	}
	if toggled.QuestID != "daily-1" {
		t.Fatalf("expected root quest toggle for daily-1, got %q", toggled.QuestID)
	}
	if len(questService.completedIDs) != 2 || questService.completedIDs[1] != "daily-1" {
		t.Fatalf("expected root quest completion path to use service CompleteQuest, got %v", questService.completedIDs)
	}
}

func TestToggleQuestCmd(t *testing.T) {
	questService := &fakeTUIQuestService{}
	cmd := ToggleQuestCmd(questService, "quest-1")
	if cmd == nil {
		t.Fatal("expected toggle quest command")
	}

	msg := cmd()
	toggled, ok := msg.(QuestToggledMsg)
	if !ok {
		t.Fatalf("expected QuestToggledMsg, got %T", msg)
	}
	if toggled.QuestID != "quest-1" {
		t.Fatalf("expected quest-1 to be toggled, got %q", toggled.QuestID)
	}
	if len(questService.completedIDs) != 1 || questService.completedIDs[0] != "quest-1" {
		t.Fatalf("expected CompleteQuest to be called once, got %v", questService.completedIDs)
	}
}

func TestWindowResize(t *testing.T) {
	now := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)
	model := NewModel([]*QuestNode{
		newRootQuest(
			"epic-1",
			"장기 프로젝트",
			domain.QuestTypeEpic,
			now,
			newSubQuest("sub-1", "설계 확정", now.Add(time.Hour)),
			newSubQuest("sub-2", "리뷰 반영", now.Add(2*time.Hour)),
		),
	}, nil)
	model.CurrentFocus = FocusDetail
	model.MasterCursor = 99
	model.DetailCursor = 99

	model, _ = updateModelForTest(t, model, tea.WindowSizeMsg{Width: 120, Height: 40})

	if model.Width != 120 || model.Height != 40 {
		t.Fatalf("expected resize to update viewport to 120x40, got %dx%d", model.Width, model.Height)
	}
	if model.MasterCursor != 0 {
		t.Fatalf("expected master cursor to clamp to 0, got %d", model.MasterCursor)
	}
	if model.DetailCursor != 1 {
		t.Fatalf("expected detail cursor to clamp to last sub quest, got %d", model.DetailCursor)
	}
	if model.CurrentFocus != FocusDetail {
		t.Fatalf("expected focus to remain detail when sub quests exist, got %v", model.CurrentFocus)
	}
}

func updateModelForTest(t *testing.T, model Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()

	next, cmd := model.Update(msg)
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("expected updated model type, got %T", next)
	}
	return updated, cmd
}
