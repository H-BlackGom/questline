package views

import (
	"fmt"
	"strings"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
)

type MasterItem struct {
	Title    string
	Type     domain.QuestType
	Status   domain.QuestStatus
	Progress float64
}

type MasterPanel struct {
	Summary  *PlayerSummary
	Items    []MasterItem
	Selected int
	Focused  bool
	Width    int
	Height   int
}

func RenderMaster(panel MasterPanel, styles theme.Styles) string {
	content := make([]string, 0, len(panel.Items)+4)
	if panel.Summary != nil {
		content = append(content,
			renderPlayerSummary(*panel.Summary, panel.Width, styles),
			styles.MutedText.Render(strings.Repeat("─", panelContentWidth(panel.Width, styles))),
		)
	}
	content = append(content, styles.PanelTitle.Render("Quest Log"))

	if len(panel.Items) == 0 {
		content = append(content, styles.MutedText.Render("표시할 루트 퀘스트가 없습니다."))
	} else {
		for index, item := range panel.Items {
			badgeStyle, ok := styles.QuestType[item.Type]
			if !ok {
				badgeStyle = styles.MutedText
			}

			line := fmt.Sprintf("%s %s %s", selectionMarker(index == panel.Selected), badgeStyle.Render("["+strings.ToUpper(string(item.Type))+"]"), item.Title)
			if item.Progress > 0 {
				line = fmt.Sprintf("%s · %d%%", line, int(item.Progress*100))
			}

			styledLine := styles.BodyText.Render(line)
			if index == panel.Selected {
				styledLine = styles.SelectedRow.Render(line)
			}

			content = append(content, styledLine)
		}
	}

	panelStyle := styles.Panel
	if panel.Focused {
		panelStyle = styles.PanelFocused
	}

	return panelStyle.Width(panel.Width).Height(panel.Height).Render(strings.Join(content, "\n"))
}

func selectionMarker(selected bool) string {
	if selected {
		return "›"
	}
	return "·"
}
