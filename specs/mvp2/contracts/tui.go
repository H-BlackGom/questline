//go:build mvp2

// Package contracts defines TUI-specific interfaces and messages for Questline MVP 2.
// These contracts establish the communication protocol between TUI components.
package contracts

type Msg = any

type Cmd func() Msg

type Model interface {
	Init() Cmd
	Update(msg Msg) (Model, Cmd)
	View() string
}

// TUIModel defines the interface for TUI state management.
// Implementations must follow the Bubble Tea Elm architecture (Model/Update/View).
type TUIModel interface {
	// Init returns the initial command to run.
	Init() Cmd

	// Update handles incoming messages and returns the updated model and command.
	Update(msg Msg) (Model, Cmd)

	// View renders the current state as a string.
	View() string
}

// FocusType indicates which panel is currently focused in the TUI.
type FocusType int

const (
	// FocusMaster indicates the left panel (quest list) is focused.
	FocusMaster FocusType = iota
	// FocusDetail indicates the right panel (sub-quests/details) is focused.
	FocusFocusDetail FocusType = iota
)

// TUI Messages
// These messages are used for communication between components in the Update loop.

// ToggleQuestMsg is sent when the user toggles a quest's completion status.
type ToggleQuestMsg struct {
	QuestID string
	IsSub   bool // If true, this is a sub-quest toggle
}

// NavigateMsg is sent for navigation commands.
type NavigateMsg struct {
	Direction Direction
}

// Direction represents navigation directions.
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// FocusChangeMsg is sent when the focus changes between panels.
type FocusChangeMsg struct {
	NewFocus FocusType
}

// ExpandQuestMsg is sent to expand/collapse an epic or guild quest.
type ExpandQuestMsg struct {
	QuestID string
}

// SyncCompleteMsg is sent when lazy evaluation completes.
type SyncCompleteMsg struct {
	Result any
	Error  error
}

// QuestsLoadedMsg is sent when quests are loaded from the database.
type QuestsLoadedMsg struct {
	Quests []any
	Error  error
}

// PlayerLoadedMsg is sent when player data is loaded.
type PlayerLoadedMsg struct {
	Player any
	Error  error
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Error error
}

type WindowSizeMsg struct {
	Width  int
	Height int
}

// LoadingMsg indicates loading state changes.
type LoadingMsg struct {
	IsLoading bool
	Message   string
}

// QuitMsg is sent to signal the TUI should quit.
type QuitMsg struct{}

// Commands
// LoadQuestsCmd creates a command to load quests from the database.
func LoadQuestsCmd(service interface{ GetQuestTree() ([]any, error) }) Cmd {
	return func() Msg {
		quests, err := service.GetQuestTree()
		return QuestsLoadedMsg{Quests: quests, Error: err}
	}
}

// LoadPlayerCmd creates a command to load player data.
func LoadPlayerCmd(service interface{ GetPlayerWithFlow() (any, error) }) Cmd {
	return func() Msg {
		player, err := service.GetPlayerWithFlow()
		return PlayerLoadedMsg{Player: player, Error: err}
	}
}

// PerformSyncCmd creates a command to perform lazy evaluation.
func PerformSyncCmd(service interface{ EvaluateLazySync() (any, error) }) Cmd {
	return func() Msg {
		result, err := service.EvaluateLazySync()
		return SyncCompleteMsg{Result: result, Error: err}
	}
}

// ToggleQuestCmd creates a command to toggle a quest's completion status.
func ToggleQuestCmd(service interface {
	CompleteQuest(string) (any, error)
	GetQuestTree() ([]any, error)
}, questID string) Cmd {
	return func() Msg {
		_, err := service.CompleteQuest(questID)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		// Reload quests after toggle
		return LoadQuestsCmd(service)()
	}
}

// TUIConfiguration contains styling and layout configuration.
type TUIConfiguration struct {
	// Layout
	MasterWidthPercent int // Width of master panel (0-100, default 50)
	MinWidth           int // Minimum terminal width (default 80)
	MinHeight          int // Minimum terminal height (default 24)

	// Colors (Lip Gloss styles will use these)
	Colors ThemeColors
}

// ThemeColors defines the color scheme for the TUI.
type ThemeColors struct {
	Primary   string // Header, borders
	Secondary string // Secondary text
	Success   string // Completed quests
	Warning   string // Warnings, HAZY status
	Danger    string // Errors, BURNING status (actually positive)
	Info      string // Info, SMOOTH status
	Muted     string // Muted text

	// Quest Type Colors
	DailyColor  string
	WeeklyColor string
	EpicColor   string
	GuildColor  string
	SubColor    string
}

// DefaultTheme returns the default color theme.
func DefaultTheme() ThemeColors {
	return ThemeColors{
		Primary:     "#7D56C4",
		Secondary:   "#5A3D8F",
		Success:     "#04B575",
		Warning:     "#F4D03F",
		Danger:      "#E74C3C",
		Info:        "#3498DB",
		Muted:       "#95A5A6",
		DailyColor:  "#3498DB", // Blue
		WeeklyColor: "#9B59B6", // Purple
		EpicColor:   "#E67E22", // Orange
		GuildColor:  "#27AE60", // Green
		SubColor:    "#7F8C8D", // Gray
	}
}
