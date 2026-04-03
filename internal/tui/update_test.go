package tui

import (
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeTUIQuestService struct {
	quests     map[string]*domain.Quest
	updatedIDs []string
	updatedTo  []domain.QuestStatus
	err        error
}

func (f *fakeTUIQuestService) GetQuestTree() ([]*domain.Quest, error) {
	quests := make([]*domain.Quest, 0, len(f.quests))
	for _, quest := range f.quests {
		quests = append(quests, quest)
	}
	return quests, nil
}

func (f *fakeTUIQuestService) GetQuest(questID string) (*domain.Quest, error) {
	if quest, ok := f.quests[questID]; ok {
		return quest, nil
	}
	return nil, nil
}

func (f *fakeTUIQuestService) UpdateQuestStatus(questID string, status domain.QuestStatus) error {
	f.updatedIDs = append(f.updatedIDs, questID)
	f.updatedTo = append(f.updatedTo, status)
	if f.err != nil {
		return f.err
	}
	if quest, ok := f.quests[questID]; ok {
		quest.Status = status
	}
	return nil
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
	questService := &fakeTUIQuestService{quests: map[string]*domain.Quest{
		"daily-1": {ID: "daily-1", Status: domain.StatusPending},
		"sub-2":   {ID: "sub-2", Status: domain.StatusPending},
	}}
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
	cycled, ok := msg.(QuestStatusCycledMsg)
	if !ok {
		t.Fatalf("expected QuestStatusCycledMsg, got %T", msg)
	}
	if cycled.QuestID != "sub-2" {
		t.Fatalf("expected sub quest cycle for sub-2, got %q", cycled.QuestID)
	}
	if cycled.NewStatus != domain.StatusInProgress {
		t.Fatalf("expected sub quest to cycle into in_progress, got %s", cycled.NewStatus)
	}
	if len(questService.updatedIDs) != 1 || questService.updatedIDs[0] != "sub-2" {
		t.Fatalf("expected sub quest cycle path to update sub-2, got %v", questService.updatedIDs)
	}
	if len(questService.updatedTo) != 1 || questService.updatedTo[0] != domain.StatusInProgress {
		t.Fatalf("expected sub quest to update status in_progress, got %v", questService.updatedTo)
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
	cycled, ok = msg.(QuestStatusCycledMsg)
	if !ok {
		t.Fatalf("expected QuestStatusCycledMsg, got %T", msg)
	}
	if cycled.QuestID != "daily-1" {
		t.Fatalf("expected root quest cycle for daily-1, got %q", cycled.QuestID)
	}
	if cycled.NewStatus != domain.StatusInProgress {
		t.Fatalf("expected root quest to cycle into in_progress, got %s", cycled.NewStatus)
	}
	if len(questService.updatedIDs) != 2 || questService.updatedIDs[1] != "daily-1" {
		t.Fatalf("expected root quest cycle path to update daily-1, got %v", questService.updatedIDs)
	}
}

func TestCycleQuestStatusCmd(t *testing.T) {
	questService := &fakeTUIQuestService{quests: map[string]*domain.Quest{
		"quest-1": {ID: "quest-1", Status: domain.StatusInProgress},
	}}
	cmd := CycleQuestStatusCmd(questService, "quest-1")
	if cmd == nil {
		t.Fatal("expected cycle quest status command")
	}

	msg := cmd()
	cycled, ok := msg.(QuestStatusCycledMsg)
	if !ok {
		t.Fatalf("expected QuestStatusCycledMsg, got %T", msg)
	}
	if cycled.QuestID != "quest-1" {
		t.Fatalf("expected quest-1 to be cycled, got %q", cycled.QuestID)
	}
	if cycled.NewStatus != domain.StatusCompleted {
		t.Fatalf("expected status to cycle from in_progress to completed, got %s", cycled.NewStatus)
	}
	if len(questService.updatedIDs) != 1 || questService.updatedIDs[0] != "quest-1" {
		t.Fatalf("expected UpdateQuestStatus to be called once, got %v", questService.updatedIDs)
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
