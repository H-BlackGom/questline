package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
)

type DetailSubItem struct {
	Title  string
	Type   domain.QuestType
	Status domain.QuestStatus
}

type DetailPanel struct {
	Title       string
	Type        domain.QuestType
	Status      domain.QuestStatus
	DueDate     *time.Time
	Progress    float64
	SubQuests   []DetailSubItem
	SelectedSub int
	Focused     bool
	Width       int
	Height      int
}

func RenderDetail(panel DetailPanel, styles theme.Styles) string {
	content := []string{styles.PanelTitle.Render("Quest Detail")}

	if strings.TrimSpace(panel.Title) == "" {
		content = append(content, styles.MutedText.Render("선택된 퀘스트가 없습니다."))
	} else {
		typeStyle, ok := styles.QuestType[panel.Type]
		if !ok {
			typeStyle = styles.MutedText
		}

		content = append(content,
			styles.BodyText.Render(panel.Title),
			fmt.Sprintf("타입: %s", typeStyle.Render(strings.ToUpper(string(panel.Type)))),
			fmt.Sprintf("상태: %s", strings.ToUpper(string(panel.Status))),
		)

		if panel.DueDate != nil {
			content = append(content, fmt.Sprintf("마감일: %s", panel.DueDate.Format("2006-01-02")))
		}
		if panel.Progress > 0 {
			content = append(content, fmt.Sprintf("진행률: %d%%", int(panel.Progress*100)))
		}

		content = append(content, "", styles.PanelTitle.Render("Sub Quests"))
		if len(panel.SubQuests) == 0 {
			content = append(content, styles.MutedText.Render("하위 퀘스트가 없습니다."))
		} else {
			for index, subQuest := range panel.SubQuests {
				typeBadge, ok := styles.QuestType[subQuest.Type]
				if !ok {
					typeBadge = styles.MutedText
				}

				line := fmt.Sprintf("%s %s %s", selectionMarker(index == panel.SelectedSub), typeBadge.Render(strings.ToUpper(string(subQuest.Type))), subQuest.Title)
				if index == panel.SelectedSub && panel.Focused {
					line = styles.SelectedRow.Render(line)
				}
				content = append(content, line)
			}
		}
	}

	panelStyle := styles.Panel
	if panel.Focused {
		panelStyle = styles.PanelFocused
	}

	return panelStyle.Width(panel.Width).Height(panel.Height).Render(strings.Join(content, "\n"))
}
