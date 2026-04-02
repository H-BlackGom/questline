package domain

type QuestStatus string

const (
	StatusPending           QuestStatus = "pending"
	StatusInProgress        QuestStatus = "in_progress"
	StatusPendingCompletion QuestStatus = "pending_completion"
	StatusCompleted         QuestStatus = "completed"
	StatusArchived          QuestStatus = "archived"
)

var allowedStatusTransitions = map[QuestStatus]map[QuestStatus]struct{}{
	StatusPending: {
		StatusInProgress: {},
		StatusCompleted:  {},
	},
	StatusInProgress: {
		StatusPendingCompletion: {},
		StatusCompleted:         {},
	},
	StatusPendingCompletion: {
		StatusCompleted: {},
	},
	StatusCompleted: {
		StatusArchived: {},
	},
	StatusArchived: {},
}

func (s QuestStatus) IsValid() bool {
	_, ok := allowedStatusTransitions[s]
	return ok
}

func (s QuestStatus) CanTransitionTo(next QuestStatus) bool {
	if !s.IsValid() || !next.IsValid() {
		return false
	}
	if s == next {
		return false
	}
	_, ok := allowedStatusTransitions[s][next]
	return ok
}
