package tui

import (
	"sort"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/service"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
)

type FocusType int

const (
	FocusMaster FocusType = iota
	FocusDetail
)

type QuestNode struct {
	Quest     *domain.Quest
	SubQuests []*QuestNode
	Progress  float64
}

type PlayerWithFlow = service.PlayerWithFlow

type questCommandService interface {
	GetQuestTree() ([]*domain.Quest, error)
	GetQuest(string) (*domain.Quest, error)
	UpdateQuestStatus(string, domain.QuestStatus) error
}

type playerCommandService interface {
	GetPlayerWithFlow() (*service.PlayerWithFlow, error)
}

type syncCommandService interface {
	EvaluateLazySync() (*service.SyncResult, error)
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

	questService  questCommandService
	playerService playerCommandService
	syncService   syncCommandService
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
	if m.questService == nil && m.playerService == nil {
		return nil
	}

	cmds := make([]tea.Cmd, 0, 2)
	if cmd := LoadQuestsCmd(m.questService); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := LoadPlayerCmd(m.playerService); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) WithServices(services *service.Services) Model {
	if services == nil {
		return m
	}

	m.questService = services.Quest
	if playerService, ok := any(services.Player).(playerCommandService); ok {
		m.playerService = playerService
	}
	m.syncService = services.Sync
	return m
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
	if maxValue < minValue {
		return minValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
