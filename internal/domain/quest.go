package domain

import "time"

// Status represents quest status
type Status string

const (
	StatusTODO    Status = "TODO"
	StatusDONE    Status = "DONE"
	StatusDROPPED Status = "DROPPED" // Reserved for MVP2
)

// Quest represents a task/quest
type Quest struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      Status     `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// IsOverdue checks if quest is past due date
func (q *Quest) IsOverdue() bool {
	if q.DueDate == nil || q.Status == StatusDONE {
		return false
	}
	return time.Now().After(*q.DueDate)
}
