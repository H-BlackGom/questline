package domain

import "time"

// Quest represents a task/quest
type Quest struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Status      QuestStatus `json:"status"`
	Type        QuestType   `json:"type"`
	ParentID    *string     `json:"parent_id,omitempty"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// IsOverdue checks if quest is past due date
func (q *Quest) IsOverdue() bool {
	if q.DueDate == nil || q.Status == StatusCompleted || q.Status == StatusArchived {
		return false
	}
	return time.Now().After(*q.DueDate)
}
