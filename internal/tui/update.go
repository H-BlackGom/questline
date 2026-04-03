package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.updateKey(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.clampCursors()
		return m, nil
	case QuestsLoadedMsg:
		if msg.Err != nil {
			m.Loading = false
			m.Error = msg.Err
			return m, nil
		}
		m.Quests = msg.Quests
		m.Error = nil
		m.Loading = false
		m.clampCursors()
		return m, nil
	case PlayerLoadedMsg:
		if msg.Err != nil {
			m.Loading = false
			m.Error = msg.Err
			return m, nil
		}
		m.Player = msg.Player
		m.Error = nil
		m.Loading = false
		return m, nil
	case SyncCompleteMsg:
		if msg.Err != nil {
			m.Loading = false
			m.Error = msg.Err
			return m, nil
		}
		m.Loading = true
		return m, tea.Batch(LoadQuestsCmd(m.questService), LoadPlayerCmd(m.playerService))
	case QuestStatusCycledMsg:
		if msg.Err != nil {
			m.Loading = false
			m.Error = msg.Err
			return m, nil
		}
		m.Loading = true
		m.Error = nil
		return m, tea.Batch(LoadQuestsCmd(m.questService), LoadPlayerCmd(m.playerService))
	case error:
		m.Loading = false
		m.Error = msg
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		m.moveCursor(-1)
		return m, nil
	case "down", "j":
		m.moveCursor(1)
		return m, nil
	case "right", "enter":
		if selected := m.SelectedQuest(); m.CurrentFocus == FocusMaster && selected != nil && len(selected.SubQuests) > 0 {
			m.CurrentFocus = FocusDetail
			m.DetailCursor = 0
			m.clampCursors()
		}
		return m, nil
	case "left", "esc":
		if m.CurrentFocus == FocusDetail {
			m.CurrentFocus = FocusMaster
			m.clampCursors()
		}
		return m, nil
	case " ":
		if questID := m.selectedQuestID(); questID != "" {
			m.Loading = true
			m.Error = nil
			return m, CycleQuestStatusCmd(m.questService, questID)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) moveCursor(delta int) {
	if m.CurrentFocus == FocusDetail {
		selected := m.SelectedQuest()
		if selected == nil || len(selected.SubQuests) == 0 {
			m.CurrentFocus = FocusMaster
			m.DetailCursor = 0
			return
		}
		m.DetailCursor = clamp(m.DetailCursor+delta, 0, len(selected.SubQuests)-1)
		return
	}

	roots := m.SortedRootQuests()
	if len(roots) == 0 {
		m.MasterCursor = 0
		m.DetailCursor = 0
		m.CurrentFocus = FocusMaster
		return
	}

	m.MasterCursor = clamp(m.MasterCursor+delta, 0, len(roots)-1)
	m.clampCursors()
}

func (m *Model) clampCursors() {
	roots := m.SortedRootQuests()
	if len(roots) == 0 {
		m.MasterCursor = 0
		m.DetailCursor = 0
		m.CurrentFocus = FocusMaster
		return
	}

	m.MasterCursor = clamp(m.MasterCursor, 0, len(roots)-1)
	selected := roots[m.MasterCursor]
	if selected == nil || len(selected.SubQuests) == 0 {
		m.DetailCursor = 0
		if m.CurrentFocus == FocusDetail {
			m.CurrentFocus = FocusMaster
		}
		return
	}

	m.DetailCursor = clamp(m.DetailCursor, 0, len(selected.SubQuests)-1)
}

func (m Model) selectedQuestID() string {
	if m.CurrentFocus == FocusDetail {
		if selected := m.SelectedSubQuest(); selected != nil && selected.Quest != nil {
			return selected.Quest.ID
		}
		return ""
	}

	if selected := m.SelectedQuest(); selected != nil && selected.Quest != nil {
		return selected.Quest.ID
	}
	return ""
}
