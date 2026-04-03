package views

import (
	"fmt"
	"strings"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

type PlayerSummary struct {
	Level      int
	Title      string
	CurrentXP  int
	RequiredXP int
	Flow       domain.FlowStatus
}

func renderPlayerSummary(summary PlayerSummary, panelWidth int, styles theme.Styles) string {
	flowStyle, ok := styles.FlowState[summary.Flow]
	if !ok {
		flowStyle = styles.MutedText
	}

	left := styles.BodyText.Bold(true).Render(fmt.Sprintf("🧙 Lv.%d %s", summary.Level, summary.Title))
	bar := fmt.Sprintf("[%s]", renderProgressBar(summary.CurrentXP, summary.RequiredXP, styles.Tokens.Layout.SummaryProgressWidth, styles))
	percent := styles.MutedText.Render(fmt.Sprintf("%d%%", progressPercent(summary.CurrentXP, summary.RequiredXP)))
	flow := flowStyle.Render(flowStatusLabel(summary.Flow))
	availableWidth := panelContentWidth(panelWidth, styles)

	firstLine, firstFits := layoutSummaryLine([]string{left, bar}, availableWidth)
	if !firstFits {
		firstLine = strings.Join([]string{left, bar}, "\n")
	}

	secondLine, secondFits := layoutSummaryLine([]string{percent, flow}, availableWidth)
	if !secondFits {
		secondLine = strings.Join([]string{percent, flow}, "\n")
	}

	return strings.Join([]string{firstLine, secondLine}, "\n")
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

func panelContentWidth(panelWidth int, styles theme.Styles) int {
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
