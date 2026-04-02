//go:build mvp2

// Package contracts defines the service interfaces for Questline MVP 2.
// These interfaces establish clear boundaries between layers and enable mocking for testing.
package contracts

import (
	"time"
)

type QuestType string

type QuestStatus string

type FlowStatus string

type Quest struct{}

type Player struct{}

// QuestService defines the interface for quest management operations.
// Implementations handle CRUD operations, quest hierarchy, and completion workflows.
type QuestService interface {
	// CreateQuest creates a new quest with the specified type and optional parent.
	// Parameters:
	//   - title: The quest title
	//   - questType: One of daily, weekly, epic, guild, sub
	//   - parentID: Optional parent quest ID (required for sub quests)
	//   - dueDate: Optional due date for the quest
	// Returns the created quest or an error if validation fails.
	CreateQuest(title string, questType QuestType, parentID *string, dueDate *time.Time) (*Quest, error)

	// CompleteQuest marks a quest as completed and triggers XP reward.
	// For sub quests, this may transition the parent to pending_completion state.
	// For quests in pending_completion state, this confirms completion and grants XP.
	// Parameters:
	//   - questID: The unique quest identifier
	// Returns the completion result including XP earned and level changes.
	CompleteQuest(questID string) (*CompletionResult, error)

	// GetQuest retrieves a quest by its ID, including its sub-quests if any.
	GetQuest(questID string) (*QuestNode, error)

	// ListQuests returns quests filtered by the given criteria.
	ListQuests(filter QuestFilter) ([]*QuestNode, error)

	// GetQuestTree returns the full quest hierarchy for TUI display.
	// Quests are sorted by type: Daily → Weekly → Epic → Guild.
	GetQuestTree() ([]*QuestNode, error)

	// ArchiveQuest soft-deletes a quest by setting deleted_at timestamp.
	ArchiveQuest(questID string) error

	// RestoreQuest restores a soft-deleted quest.
	RestoreQuest(questID string) error
}

// QuestNode represents a quest in the hierarchy tree for TUI display.
type QuestNode struct {
	Quest      *Quest
	SubQuests  []*QuestNode
	IsExpanded bool    // Whether epic/guild is expanded in TUI
	Progress   float64 // Completion progress (0.0 to 1.0)
}

// QuestFilter defines filtering criteria for quest queries.
type QuestFilter struct {
	Types          []QuestType   // Filter by quest types (empty = all)
	Statuses       []QuestStatus // Filter by statuses (empty = all)
	ParentID       *string       // Filter by parent (nil = root quests, "null" = sub quests)
	IncludeDeleted bool          // Include soft-deleted quests
}

// CompletionResult contains the outcome of completing a quest.
type CompletionResult struct {
	Quest            *Quest
	XPBefore         int
	XPAfter          int
	XPBonus          float64 // Flow multiplier applied (e.g., 1.5 for BURNING)
	LevelBefore      int
	LevelAfter       int
	LevelUpOccurred  bool
	TitleBefore      string
	TitleAfter       string
	ParentTransition *ParentTransition // If parent entered pending_completion
}

// ParentTransition indicates a parent quest state change.
type ParentTransition struct {
	ParentID         string
	OldStatus        QuestStatus
	NewStatus        QuestStatus // Will be pending_completion
	AllSubquestsDone bool
}

// SyncService handles lazy evaluation and time-based quest management.
// This service is triggered on app startup to synchronize quest states
// based on elapsed time since last sync.
type SyncService interface {
	// EvaluateLazySync performs lazy evaluation from last_synced_at to now.
	// This should be called on every app startup (including TUI).
	// It handles:
	//   - Daily quest archival for missed days
	//   - Flow grade calculation for completed days
	//   - Streak updates
	// Returns the evaluation results for all processed days.
	EvaluateLazySync() (*SyncResult, error)

	// CalculateFlowGrade determines the Flow status for a specific day's completion rate.
	// Parameters:
	//   - date: The date to evaluate (4 AM based)
	// Returns the calculated flow grade.
	CalculateFlowGrade(date time.Time) (*FlowGrade, error)

	// GetLastSync returns the timestamp of the last successful sync.
	GetLastSync() (time.Time, error)

	// ForceSync forces a sync for the given date range (for testing/admin).
	ForceSync(startDate, endDate time.Time) (*SyncResult, error)
}

// SyncResult contains the outcomes of a lazy evaluation.
type SyncResult struct {
	EvaluatedFrom    time.Time         // Start of evaluation period
	EvaluatedTo      time.Time         // End of evaluation period (now)
	ProcessedDays    int               // Number of days processed
	EvaluatedDates   []string          // List of dates processed (YYYY-MM-DD)
	FlowAdjustments  []*FlowAdjustment // Flow grade changes
	ArchivedQuests   int               // Number of quests archived
	NewQuestsCreated int               // Number of new daily/weekly quests created
	StreakUpdated    bool              // Whether streak was updated
	NewStreakCount   int               // Current streak after evaluation
}

// FlowAdjustment represents a flow grade change for a specific date.
type FlowAdjustment struct {
	Date              string     // YYYY-MM-DD
	PreviousGrade     FlowStatus // Grade before evaluation
	NewGrade          FlowStatus // Grade after evaluation
	CompletionRate    float64    // 0.0 to 1.0
	TotalRoutines     int
	CompletedRoutines int
}

// FlowGrade contains the detailed flow calculation for a date.
type FlowGrade struct {
	Date              time.Time
	CompletionRate    float64    // 0.0 to 1.0
	Grade             FlowStatus // BURNING, SMOOTH, or HAZY
	Multiplier        float64    // 1.5, 1.0, or 0.5
	TotalRoutines     int
	CompletedRoutines int
}

// PlayerService manages player stats and progression.
type PlayerService interface {
	// GetPlayer returns the current player state.
	GetPlayer() (*Player, error)

	// GetPlayerWithFlow returns player with current flow multiplier.
	GetPlayerWithFlow() (*PlayerWithFlow, error)

	// AwardXP grants XP to the player, applying flow multiplier.
	// Returns level-up information if a level up occurred.
	AwardXP(baseXP int, questID string) (*AwardResult, error)

	// UpdateFlowStatus manually updates the flow status (for testing).
	UpdateFlowStatus(status FlowStatus) error

	// GetFlowHistory returns the flow grade history for the last N days.
	GetFlowHistory(days int) ([]*FlowHistoryEntry, error)
}

// PlayerWithFlow combines player stats with flow information.
type PlayerWithFlow struct {
	*Player
	CurrentFlow    FlowStatus
	FlowMultiplier float64
	NextEvaluation time.Time
}

// AwardResult contains the outcome of an XP award.
type AwardResult struct {
	XPBefore        int
	XPAfter         int
	BaseXP          int
	Multiplier      float64
	BonusXP         int
	LevelBefore     int
	LevelAfter      int
	LevelUpOccurred bool
}

// FlowHistoryEntry represents a single day's flow status.
type FlowHistoryEntry struct {
	Date              string // YYYY-MM-DD
	FlowStatus        FlowStatus
	CompletionRate    float64
	RoutinesTotal     int
	RoutinesCompleted int
}
