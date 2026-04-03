package theme

import (
	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

type LayoutTokens struct {
	MinWidth           int
	MinHeight          int
	MasterWidthPercent int
	Gap                int
	PaddingX           int
	PaddingY           int
}

type ColorTokens struct {
	Primary   string
	Secondary string
	Success   string
	Warning   string
	Danger    string
	Info      string
	Muted     string
	Surface   string
	SurfaceHi string
	Text      string

	DailyColor  string
	WeeklyColor string
	EpicColor   string
	GuildColor  string
	SubColor    string

	BurningColor     string
	SingularityColor string
	SmoothColor      string
	HazyColor        string
}

type Tokens struct {
	Layout LayoutTokens
	Colors ColorTokens
}

type Styles struct {
	Tokens Tokens

	App          lipgloss.Style
	Header       lipgloss.Style
	Footer       lipgloss.Style
	Panel        lipgloss.Style
	PanelFocused lipgloss.Style
	PanelTitle   lipgloss.Style
	BodyText     lipgloss.Style
	MutedText    lipgloss.Style
	SelectedRow  lipgloss.Style
	Error        lipgloss.Style
	Loading      lipgloss.Style
	Fallback     lipgloss.Style
	KeyHint      lipgloss.Style

	QuestType map[domain.QuestType]lipgloss.Style
	FlowState map[domain.FlowStatus]lipgloss.Style
}

func DefaultStyles() Styles {
	tokens := Tokens{
		Layout: LayoutTokens{
			MinWidth:           80,
			MinHeight:          24,
			MasterWidthPercent: 50,
			Gap:                2,
			PaddingX:           1,
			PaddingY:           0,
		},
		Colors: ColorTokens{
			Primary:          "#7D56C4",
			Secondary:        "#5A3D8F",
			Success:          "#04B575",
			Warning:          "#F4D03F",
			Danger:           "#E74C3C",
			Info:             "#3498DB",
			Muted:            "#95A5A6",
			Surface:          "#1F2430",
			SurfaceHi:        "#2B3245",
			Text:             "#F5F7FA",
			DailyColor:       "#3498DB",
			WeeklyColor:      "#9B59B6",
			EpicColor:        "#E67E22",
			GuildColor:       "#27AE60",
			SubColor:         "#7F8C8D",
			BurningColor:     "#E74C3C",
			SingularityColor: "#C678DD",
			SmoothColor:      "#3498DB",
			HazyColor:        "#F4D03F",
		},
	}

	panelBase := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(tokens.Colors.Secondary)).
		Foreground(lipgloss.Color(tokens.Colors.Text)).
		Padding(tokens.Layout.PaddingY, tokens.Layout.PaddingX)

	questType := map[domain.QuestType]lipgloss.Style{
		domain.QuestTypeDaily:  lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.DailyColor)).Bold(true),
		domain.QuestTypeWeekly: lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.WeeklyColor)).Bold(true),
		domain.QuestTypeEpic:   lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.EpicColor)).Bold(true),
		domain.QuestTypeGuild:  lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.GuildColor)).Bold(true),
		domain.QuestTypeSub:    lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.SubColor)).Bold(true),
	}

	flowState := map[domain.FlowStatus]lipgloss.Style{
		domain.FlowStatusSingularity: lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.SingularityColor)).Bold(true),
		domain.FlowStatusBurning:     lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.BurningColor)).Bold(true),
		domain.FlowStatusSmooth:      lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.SmoothColor)).Bold(true),
		domain.FlowStatusHazy:        lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.HazyColor)).Bold(true),
	}

	panelFocused := panelBase.BorderForeground(lipgloss.Color(tokens.Colors.Primary))

	return Styles{
		Tokens:       tokens,
		App:          lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)),
		Header:       lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)).Background(lipgloss.Color(tokens.Colors.Primary)).Padding(0, 1).Bold(true),
		Footer:       lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)).Background(lipgloss.Color(tokens.Colors.Secondary)).Padding(0, 1),
		Panel:        panelBase,
		PanelFocused: panelFocused,
		PanelTitle:   lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Primary)).Bold(true),
		BodyText:     lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)),
		MutedText:    lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Muted)),
		SelectedRow:  lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)).Background(lipgloss.Color(tokens.Colors.Secondary)).Bold(true),
		Error:        lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Danger)).Bold(true),
		Loading:      lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Info)).Bold(true),
		Fallback:     lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Warning)).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(tokens.Colors.Warning)).Padding(tokens.Layout.PaddingY+1, tokens.Layout.PaddingX+1),
		KeyHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(tokens.Colors.Text)).Background(lipgloss.Color(tokens.Colors.Secondary)).Bold(true),
		QuestType:    questType,
		FlowState:    flowState,
	}
}
