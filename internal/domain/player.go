package domain

import "time"

// Player represents user progression
type Player struct {
	ID              int        `json:"id"`
	Level           int        `json:"level"`
	CurrentXP       int        `json:"current_xp"`
	TotalXPEarned   int        `json:"total_xp_earned"`
	QuestsCompleted int        `json:"quests_completed"`
	FlowStatus      FlowStatus `json:"flow_status"`
	LastSyncedAt    *time.Time `json:"last_synced_at,omitempty"`
	LastEvaluated   *time.Time `json:"last_evaluated,omitempty"`
	StreakDays      int        `json:"streak_days"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// GetRequiredXPForNextLevel calculates XP needed for next level
// Formula: 100 + (current_level * 50)
func (p *Player) GetRequiredXPForNextLevel() int {
	return 100 + (p.Level * 50)
}

// GetTitle returns title based on level
func (p *Player) GetTitle() string {
	switch {
	case p.Level >= 99:
		return "Guru"
	case p.Level >= 50:
		return "Principal"
	case p.Level >= 30:
		return "Lead"
	case p.Level >= 20:
		return "Senior"
	case p.Level >= 10:
		return "Junior"
	default:
		return "Intern"
	}
}
