package tui

import (
	"strings"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

type QuestsLoadedMsg struct {
	Quests []*QuestNode
	Err    error
}

type PlayerLoadedMsg struct {
	Player *PlayerWithFlow
	Err    error
}

type SyncCompleteMsg struct {
	Result *service.SyncResult
	Err    error
}

type QuestToggledMsg struct {
	QuestID string
	Err     error
}

func LoadQuestsCmd(service interface {
	GetQuestTree() ([]*domain.Quest, error)
}) tea.Cmd {
	if service == nil {
		return nil
	}

	return func() tea.Msg {
		quests, err := service.GetQuestTree()
		if err != nil {
			return QuestsLoadedMsg{Err: err}
		}
		return QuestsLoadedMsg{Quests: buildQuestNodes(quests)}
	}
}

func LoadPlayerCmd(service interface {
	GetPlayerWithFlow() (*service.PlayerWithFlow, error)
}) tea.Cmd {
	if service == nil {
		return nil
	}

	return func() tea.Msg {
		player, err := service.GetPlayerWithFlow()
		return PlayerLoadedMsg{Player: player, Err: err}
	}
}

func PerformSyncCmd(service interface {
	EvaluateLazySync() (*service.SyncResult, error)
}) tea.Cmd {
	if service == nil {
		return nil
	}

	return func() tea.Msg {
		result, err := service.EvaluateLazySync()
		return SyncCompleteMsg{Result: result, Err: err}
	}
}

func ToggleQuestCmd(service interface {
	CompleteQuest(string) (*service.CompletionResult, error)
}, questID string) tea.Cmd {
	if service == nil || strings.TrimSpace(questID) == "" {
		return nil
	}

	return func() tea.Msg {
		_, err := service.CompleteQuest(questID)
		return QuestToggledMsg{QuestID: questID, Err: err}
	}
}

func buildQuestNodes(quests []*domain.Quest) []*QuestNode {
	if len(quests) == 0 {
		return nil
	}

	nodes := make(map[string]*QuestNode, len(quests))
	for _, quest := range quests {
		if quest == nil {
			continue
		}
		nodes[quest.ID] = &QuestNode{Quest: quest}
	}

	roots := make([]*QuestNode, 0, len(quests))
	for _, quest := range quests {
		if quest == nil {
			continue
		}

		node := nodes[quest.ID]
		if quest.ParentID != nil {
			if parent, ok := nodes[*quest.ParentID]; ok {
				parent.SubQuests = append(parent.SubQuests, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	for _, root := range roots {
		applyProgress(root)
	}

	return roots
}

func applyProgress(node *QuestNode) float64 {
	if node == nil || node.Quest == nil {
		return 0
	}
	if len(node.SubQuests) == 0 {
		if node.Quest.Status == domain.StatusCompleted {
			node.Progress = 1
		}
		return node.Progress
	}

	completed := 0
	for _, subQuest := range node.SubQuests {
		applyProgress(subQuest)
		if subQuest != nil && subQuest.Quest != nil && subQuest.Quest.Status == domain.StatusCompleted {
			completed++
		}
	}

	node.Progress = float64(completed) / float64(len(node.SubQuests))
	return node.Progress
}
