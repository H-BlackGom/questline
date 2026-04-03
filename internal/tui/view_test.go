package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	"github.com/H-BlackGom/questline/internal/tui/views"
)

func TestInitialModel(t *testing.T) {
	model := NewModel(nil, nil)

	if model.CurrentFocus != FocusMaster {
		t.Fatalf("expected initial focus to be master, got %v", model.CurrentFocus)
	}
	if model.MasterCursor != 0 || model.DetailCursor != 0 {
		t.Fatalf("expected cursors to start at zero, got master=%d detail=%d", model.MasterCursor, model.DetailCursor)
	}
	if model.Width != 80 || model.Height != 24 {
		t.Fatalf("expected default viewport 80x24, got %dx%d", model.Width, model.Height)
	}

	view := model.View()
	if !strings.Contains(view, "Quest Log") || !strings.Contains(view, "Quest Detail") {
		t.Fatalf("expected initial render to show two-pane scaffolding, got %q", view)
	}
	if strings.Contains(view, "터미널이 너무 작습니다") {
		t.Fatalf("expected initial render to avoid small terminal fallback, got %q", view)
	}
}

func TestRenderMasterDetail(t *testing.T) {
	now := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)
	due := now.Add(48 * time.Hour)

	model := NewModel([]*QuestNode{
		newRootQuest("guild-1", "길드 원정", domain.QuestTypeGuild, now.Add(4*time.Hour)),
		newRootQuest("epic-1", "장기 프로젝트", domain.QuestTypeEpic, now.Add(3*time.Hour),
			newSubQuest("sub-1", "설계 확정", now.Add(3*time.Hour)),
			newSubQuest("sub-2", "리뷰 반영", now.Add(3*time.Hour).Add(time.Minute)),
		),
		newRootQuest("weekly-1", "주간 회고", domain.QuestTypeWeekly, now.Add(2*time.Hour)),
		newRootQuest("daily-1", "아침 루틴", domain.QuestTypeDaily, now.Add(time.Hour)),
	}, &PlayerWithFlow{
		Player:         &domain.Player{Level: 7, CurrentXP: 90},
		CurrentFlow:    domain.FlowStatusBurning,
		FlowMultiplier: 1.5,
		NextEvaluation: due,
	})
	model.Width = 100
	model.Height = 30
	model.MasterCursor = 2

	sorted := model.SortedRootQuests()
	if len(sorted) != 4 {
		t.Fatalf("expected 4 root quests, got %d", len(sorted))
	}
	for index, expected := range []string{"아침 루틴", "주간 회고", "장기 프로젝트", "길드 원정"} {
		if sorted[index].Quest.Title != expected {
			t.Fatalf("expected sorted root quest %d to be %q, got %q", index, expected, sorted[index].Quest.Title)
		}
	}

	view := model.View()

	for _, expected := range []string{"Quest Log", "Quest Detail", "장기 프로젝트", "설계 확정", "리뷰 반영", "Lv.7", "Intern", "20%", "BURNING"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected render to include %q, got %q", expected, view)
		}
	}
}

func TestRenderDetailShowsSingularityProfileStatus(t *testing.T) {
	now := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)

	model := NewModel([]*QuestNode{
		newRootQuest("epic-1", "장기 프로젝트", domain.QuestTypeEpic, now),
	}, &PlayerWithFlow{
		Player:      &domain.Player{Level: 30, CurrentXP: 960},
		CurrentFlow: domain.FlowStatusSingularity,
	})
	model.Width = 100
	model.Height = 30

	view := model.View()

	for _, expected := range []string{"Lv.30", "60%", "SINGULARITY"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("expected singularity summary to include %q, got %q", expected, view)
		}
	}
}

func TestMasterPanelShowsPlayerSummaryAboveQuestList(t *testing.T) {
	now := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)
	model := NewModel([]*QuestNode{
		newRootQuest("epic-1", "장기 프로젝트", domain.QuestTypeEpic, now),
	}, &PlayerWithFlow{
		Player:      &domain.Player{Level: 30, CurrentXP: 960},
		CurrentFlow: domain.FlowStatusSingularity,
	})

	master := views.RenderMaster(masterPanelForModel(model, 48, 12), theme.DefaultStyles())
	for _, expected := range []string{"Lv.30", "60%", "SINGULARITY", "Quest Log", "장기 프로젝트"} {
		if !strings.Contains(master, expected) {
			t.Fatalf("expected master panel to include %q, got %q", expected, master)
		}
	}

	detail := views.RenderDetail(detailPanelForModel(model, 48, 12), theme.DefaultStyles())
	for _, unexpected := range []string{"Lv.30", "60%", "SINGULARITY"} {
		if strings.Contains(detail, unexpected) {
			t.Fatalf("expected detail panel to omit %q after summary move, got %q", unexpected, detail)
		}
	}
}

func TestSmallTerminalFallback(t *testing.T) {
	model := NewModel(nil, nil)
	model.Width = 60
	model.Height = 20

	view := model.View()
	if !strings.Contains(view, "최소 80x24") {
		t.Fatalf("expected explicit fallback message, got %q", view)
	}
	if strings.Contains(view, "Quest Detail") {
		t.Fatalf("expected fallback to skip two-pane render, got %q", view)
	}
}

func newRootQuest(id, title string, questType domain.QuestType, createdAt time.Time, subQuests ...*QuestNode) *QuestNode {
	return &QuestNode{
		Quest: &domain.Quest{
			ID:        id,
			Title:     title,
			Type:      questType,
			Status:    domain.StatusPending,
			CreatedAt: createdAt,
		},
		SubQuests: subQuests,
		Progress:  0.5,
	}

}

func newSubQuest(id, title string, createdAt time.Time) *QuestNode {
	parentID := "parent"
	return &QuestNode{
		Quest: &domain.Quest{
			ID:        id,
			Title:     title,
			Type:      domain.QuestTypeSub,
			Status:    domain.StatusPending,
			ParentID:  &parentID,
			CreatedAt: createdAt,
		},
	}
}
