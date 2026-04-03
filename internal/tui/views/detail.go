package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

type DetailProfile struct {
	Level      int
	Title      string
	CurrentXP  int
	RequiredXP int
	Flow       domain.FlowStatus
}

type DetailSubItem struct {
	Title  string
	Type   domain.QuestType
	Status domain.QuestStatus
}

type DetailPanel struct {
	Profile     *DetailProfile
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
	if panel.Profile != nil {
		content = append(content, renderProfileSummary(*panel.Profile, panel.Width, styles), "")
	}

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

func renderProfileSummary(profile DetailProfile, panelWidth int, styles theme.Styles) string {
	flowStyle, ok := styles.FlowState[profile.Flow]
	if !ok {
		flowStyle = styles.MutedText
	}

	left := styles.BodyText.Bold(true).Render(fmt.Sprintf("🧙 [Lv.%d %s]", profile.Level, profile.Title))
	center := fmt.Sprintf("[%s] %d%%", renderProgressBar(profile.CurrentXP, profile.RequiredXP, styles.Tokens.Layout.SummaryProgressWidth, styles), progressPercent(profile.CurrentXP, profile.RequiredXP))
	right := flowStyle.Render(flowStatusLabel(profile.Flow))
	availableWidth := detailContentWidth(panelWidth, styles)

	if line, ok := layoutSummaryLine([]string{left, center, right}, availableWidth); ok {
		return line
	}
	if line, ok := layoutSummaryLine([]string{center, right}, availableWidth); ok {
		return strings.Join([]string{left, line}, "\n")
	}
	return strings.Join([]string{left, center, right}, "\n")
}

func renderProgressBar(currentXP, requiredXP, width int, styles theme.Styles) string {
	filled := max(0, min(progressFilled(currentXP, requiredXP, width), width))
	empty := width - filled

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.PanelTitle.Render(strings.Repeat("▓", filled)),
		styles.MutedText.Render(strings.Repeat("░", empty)),
	)
}

func progressFilled(currentXP, requiredXP, width int) int {
	if requiredXP <= 0 {
		return width
	}
	return (currentXP * width) / requiredXP
}

func progressPercent(currentXP, requiredXP int) int {
	if requiredXP <= 0 {
		return 100
	}
	percent := (currentXP * 100) / requiredXP
	if percent > 100 {
		return 100
	}
	if percent < 0 {
		return 0
	}
	return percent
}

func flowStatusLabel(status domain.FlowStatus) string {
	switch status {
	case domain.FlowStatusSingularity:
		return "✨ SINGULARITY"
	case domain.FlowStatusBurning:
		return "🔥 BURNING"
	case domain.FlowStatusHazy:
		return "🌫️ HAZY"
	default:
		return "🌊 SMOOTH"
	}
}

func detailContentWidth(panelWidth int, styles theme.Styles) int {
	const panelBorderColumns = 2
	available := panelWidth - (styles.Tokens.Layout.PaddingX * 2) - panelBorderColumns
	if available < 1 {
		return 1
	}
	return available
}

func layoutSummaryLine(segments []string, width int) (string, bool) {
	if len(segments) == 0 {
		return "", true
	}
	if len(segments) == 1 {
		return segments[0], lipgloss.Width(segments[0]) <= width
	}

	const minGap = 2
	segmentWidth := 0
	for _, segment := range segments {
		segmentWidth += lipgloss.Width(segment)
	}

	minimumWidth := segmentWidth + ((len(segments) - 1) * minGap)
	if minimumWidth > width {
		return "", false
	}

	extra := width - minimumWidth
	gaps := make([]int, len(segments)-1)
	for index := range gaps {
		gaps[index] = minGap
	}
	for extra > 0 {
		for index := range gaps {
			if extra == 0 {
				break
			}
			gaps[index]++
			extra--
		}
	}

	var builder strings.Builder
	for index, segment := range segments {
		if index > 0 {
			builder.WriteString(strings.Repeat(" ", gaps[index-1]))
		}
		builder.WriteString(segment)
	}

	return builder.String(), true
}
