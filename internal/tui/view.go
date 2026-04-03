package tui

import (
	"fmt"
	"strings"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	"github.com/H-BlackGom/questline/internal/tui/views"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	styles := theme.DefaultStyles()
	width := resolvedDimension(m.Width, styles.Tokens.Layout.MinWidth)
	height := resolvedDimension(m.Height, styles.Tokens.Layout.MinHeight)

	if width < styles.Tokens.Layout.MinWidth || height < styles.Tokens.Layout.MinHeight {
		return renderFallback(width, height, styles)
	}

	header := views.RenderHeader(headerDataForModel(m, width), styles)
	footer := views.RenderFooter(views.FooterData{FocusLabel: focusLabel(m.CurrentFocus), Width: width}, styles)
	bodyHeight := height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	masterWidth := width * styles.Tokens.Layout.MasterWidthPercent / 100
	detailWidth := width - masterWidth - styles.Tokens.Layout.Gap
	if detailWidth < 1 {
		detailWidth = 1
	}

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		views.RenderMaster(masterPanelForModel(m, masterWidth, bodyHeight), styles),
		lipgloss.NewStyle().Width(styles.Tokens.Layout.Gap).Render(strings.Repeat(" ", styles.Tokens.Layout.Gap)),
		views.RenderDetail(detailPanelForModel(m, detailWidth, bodyHeight), styles),
	)

	return styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func renderFallback(width, height int, styles theme.Styles) string {
	message := fmt.Sprintf(
		"터미널이 너무 작습니다. 최소 %dx%d가 필요합니다. 현재 %dx%d",
		styles.Tokens.Layout.MinWidth,
		styles.Tokens.Layout.MinHeight,
		width,
		height,
	)

	return styles.Fallback.Width(resolvedDimension(width, styles.Tokens.Layout.MinWidth)).Render(message)
}

func masterPanelForModel(m Model, width, height int) views.MasterPanel {
	items := make([]views.MasterItem, 0, len(m.Quests))
	for _, quest := range m.SortedRootQuests() {
		items = append(items, views.MasterItem{
			Title:    quest.Quest.Title,
			Type:     quest.Quest.Type,
			Status:   quest.Quest.Status,
			Progress: quest.Progress,
		})
	}

	return views.MasterPanel{
		Items:    items,
		Selected: clamp(m.MasterCursor, 0, max(len(items)-1, 0)),
		Focused:  m.CurrentFocus == FocusMaster,
		Width:    width,
		Height:   height,
	}
}

func detailPanelForModel(m Model, width, height int) views.DetailPanel {
	selected := m.SelectedQuest()
	if selected == nil || selected.Quest == nil {
		return views.DetailPanel{Focused: m.CurrentFocus == FocusDetail, Width: width, Height: height}
	}

	var profile *views.DetailProfile
	if m.Player != nil && m.Player.Player != nil {
		flow := m.Player.CurrentFlow
		if !flow.IsValid() {
			flow = domain.FlowStatusSmooth
		}

		profile = &views.DetailProfile{
			Level:      m.Player.Level,
			Title:      m.Player.Player.GetTitle(),
			CurrentXP:  m.Player.CurrentXP,
			RequiredXP: m.Player.Player.GetRequiredXPForNextLevel(),
			Flow:       flow,
		}
	}

	subQuests := make([]views.DetailSubItem, 0, len(selected.SubQuests))
	for _, subQuest := range selected.SubQuests {
		if subQuest == nil || subQuest.Quest == nil {
			continue
		}
		subQuests = append(subQuests, views.DetailSubItem{
			Title:  subQuest.Quest.Title,
			Type:   subQuest.Quest.Type,
			Status: subQuest.Quest.Status,
		})
	}

	return views.DetailPanel{
		Profile:     profile,
		Title:       selected.Quest.Title,
		Type:        selected.Quest.Type,
		Status:      selected.Quest.Status,
		DueDate:     selected.Quest.DueDate,
		Progress:    selected.Progress,
		SubQuests:   subQuests,
		SelectedSub: clamp(m.DetailCursor, 0, max(len(subQuests)-1, 0)),
		Focused:     m.CurrentFocus == FocusDetail,
		Width:       width,
		Height:      height,
	}
}

func headerDataForModel(m Model, width int) views.HeaderData {
	data := views.HeaderData{Flow: domain.FlowStatusSmooth, FlowMultiplier: 1.0, Width: width}
	if m.Player != nil {
		data.Level = m.Player.Level
		data.CurrentXP = m.Player.CurrentXP
		data.RequiredXP = m.Player.GetRequiredXPForNextLevel()
		data.Flow = m.Player.CurrentFlow
		if !data.Flow.IsValid() {
			data.Flow = domain.FlowStatusSmooth
		}
		data.FlowMultiplier = m.Player.FlowMultiplier
		if data.FlowMultiplier == 0 {
			data.FlowMultiplier = 1.0
		}
		data.NextEvaluation = m.Player.NextEvaluation
	}
	data.Loading = m.Loading
	if m.Error != nil {
		data.Error = m.Error.Error()
	}
	return data
}

func focusLabel(focus FocusType) string {
	if focus == FocusDetail {
		return "DETAIL"
	}
	return "MASTER"
}

func resolvedDimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
