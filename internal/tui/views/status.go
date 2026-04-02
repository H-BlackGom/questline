package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/tui/theme"
)

type HeaderData struct {
	Level          int
	CurrentXP      int
	RequiredXP     int
	Flow           domain.FlowStatus
	FlowMultiplier float64
	NextEvaluation time.Time
	Loading        bool
	Error          string
	Width          int
}

type FooterData struct {
	FocusLabel string
	Width      int
}

func RenderHeader(data HeaderData, styles theme.Styles) string {
	flowStyle, ok := styles.FlowState[data.Flow]
	if !ok {
		flowStyle = styles.MutedText
	}

	parts := []string{
		fmt.Sprintf("Lv.%d", data.Level),
		fmt.Sprintf("XP %d/%d", data.CurrentXP, data.RequiredXP),
		flowStyle.Render(strings.ToUpper(string(data.Flow))),
		fmt.Sprintf("x%.1f", data.FlowMultiplier),
	}

	if !data.NextEvaluation.IsZero() {
		parts = append(parts, fmt.Sprintf("다음 평가 %s", data.NextEvaluation.Format("01-02 15:04")))
	}
	if data.Loading {
		parts = append(parts, styles.Loading.Render("LOADING"))
	}
	if strings.TrimSpace(data.Error) != "" {
		parts = append(parts, styles.Error.Render(data.Error))
	}

	return styles.Header.Width(data.Width).Render(strings.Join(parts, "  •  "))
}

func RenderFooter(data FooterData, styles theme.Styles) string {
	hints := []string{
		fmt.Sprintf("포커스: %s", data.FocusLabel),
		styles.KeyHint.Render("↑/↓ 이동"),
		styles.KeyHint.Render("enter/→ 진입"),
		styles.KeyHint.Render("esc/← 복귀"),
		styles.KeyHint.Render("space 완료"),
		styles.KeyHint.Render("q 종료"),
	}

	return styles.Footer.Width(data.Width).Render(strings.Join(hints, "  •  "))
}
