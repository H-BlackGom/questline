package tui

import (
	"sort"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
)

type FocusType int

const (
	FocusMaster FocusType = iota
	FocusDetail
)

type QuestNode struct {
	Quest      *domain.Quest
	SubQuests  []*QuestNode
	IsExpanded bool
	Progress   float64
}

type PlayerWithFlow struct {
	*domain.Player
	CurrentFlow    domain.FlowStatus
	FlowMultiplier float64
	NextEvaluation time.Time
}

type Model struct {
	CurrentFocus FocusType
	MasterCursor int
	DetailCursor int

	Quests []*QuestNode
	Player *PlayerWithFlow

	Width   int
	Height  int
	Loading bool
	Error   error
}

func NewModel(quests []*QuestNode, player *PlayerWithFlow) Model {
	layout := theme.DefaultStyles().Tokens.Layout

	return Model{
		CurrentFocus: FocusMaster,
		MasterCursor: 0,
		DetailCursor: 0,
		Quests:       quests,
		Player:       player,
		Width:        layout.MinWidth,
		Height:       layout.MinHeight,
	}

}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case error:
		m.Error = msg
	}

	return m, nil
}

func (m Model) SortedRootQuests() []*QuestNode {
	roots := make([]*QuestNode, 0, len(m.Quests))
	for _, quest := range m.Quests {
		if quest == nil || quest.Quest == nil {
			continue
		}
		if quest.Quest.ParentID != nil || quest.Quest.Type == domain.QuestTypeSub {
			continue
		}
		roots = append(roots, quest)
	}

	sort.SliceStable(roots, func(i, j int) bool {
		left := roots[i].Quest
		right := roots[j].Quest

		leftRank := questTypeRank(left.Type)
		rightRank := questTypeRank(right.Type)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}
		return left.ID < right.ID
	})

	return roots
}

func (m Model) SelectedQuest() *QuestNode {
	quests := m.SortedRootQuests()
	if len(quests) == 0 {
		return nil
	}

	index := clamp(m.MasterCursor, 0, len(quests)-1)
	return quests[index]
}

func (m Model) SelectedSubQuest() *QuestNode {
	selected := m.SelectedQuest()
	if selected == nil || len(selected.SubQuests) == 0 {
		return nil
	}

	index := clamp(m.DetailCursor, 0, len(selected.SubQuests)-1)
	return selected.SubQuests[index]
}

func questTypeRank(questType domain.QuestType) int {
	switch questType {
	case domain.QuestTypeDaily:
		return 1
	case domain.QuestTypeWeekly:
		return 2
	case domain.QuestTypeEpic:
		return 3
	case domain.QuestTypeGuild:
		return 4
	default:
		return 5
	}
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
