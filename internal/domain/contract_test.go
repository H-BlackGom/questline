package domain

import "testing"

func TestQuestTypeValidation(t *testing.T) {
	valid := []QuestType{QuestTypeDaily, QuestTypeWeekly, QuestTypeEpic, QuestTypeGuild, QuestTypeSub}
	for _, questType := range valid {
		if !questType.IsValid() {
			t.Fatalf("expected %q to be valid", questType)
		}
	}

	invalid := QuestType("boss")
	if invalid.IsValid() {
		t.Fatalf("expected %q to be invalid", invalid)
	}
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name string
		from QuestStatus
		to   QuestStatus
		ok   bool
	}{
		{name: "pending to in progress", from: StatusPending, to: StatusInProgress, ok: true},
		{name: "pending to completed", from: StatusPending, to: StatusCompleted, ok: true},
		{name: "in progress to pending completion", from: StatusInProgress, to: StatusPendingCompletion, ok: true},
		{name: "in progress to completed", from: StatusInProgress, to: StatusCompleted, ok: true},
		{name: "pending completion to completed", from: StatusPendingCompletion, to: StatusCompleted, ok: true},
		{name: "completed to archived", from: StatusCompleted, to: StatusArchived, ok: true},
		{name: "pending cannot jump to archived", from: StatusPending, to: StatusArchived, ok: false},
		{name: "archived cannot transition", from: StatusArchived, to: StatusCompleted, ok: false},
		{name: "same state transition invalid", from: StatusCompleted, to: StatusCompleted, ok: false},
		{name: "invalid state rejected", from: QuestStatus("legacy"), to: StatusCompleted, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.ok {
				t.Fatalf("transition %q -> %q = %v, want %v", tt.from, tt.to, got, tt.ok)
			}
		})
	}
}
