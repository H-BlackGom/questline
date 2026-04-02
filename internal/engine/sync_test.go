package engine

import (
	"testing"
	"time"
)

func TestEvaluateLazySync(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)
	now := time.Date(2026, 4, 5, 5, 0, 0, 0, loc)
	lastSynced := time.Date(2026, 4, 3, 10, 0, 0, 0, loc)

	result, err := EvaluateLazySync(lastSynced, now, NewFixedClock(now))
	if err != nil {
		t.Fatalf("EvaluateLazySync returned error: %v", err)
	}
	if result.EvaluationDate == nil {
		t.Fatalf("expected evaluation date for stale sync")
	}

	wantEvaluationDate := time.Date(2026, 4, 3, 0, 0, 0, 0, loc)
	if !result.EvaluationDate.Equal(wantEvaluationDate) {
		t.Fatalf("expected evaluation date %v, got %v", wantEvaluationDate, *result.EvaluationDate)
	}
}

func TestWeeklyRollover(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)
	base := time.Date(2026, 4, 3, 9, 0, 0, 0, loc)

	result, err := EvaluateLazySync(base, base, NewFixedClock(base))
	if err != nil {
		t.Fatalf("EvaluateLazySync returned error: %v", err)
	}

	want := time.Date(2026, 4, 5, 0, 0, 0, 0, loc)
	if !result.WeeklyRolloverTo.Equal(want) {
		t.Fatalf("expected weekly rollover %v, got %v", want, result.WeeklyRolloverTo)
	}

	sunday := time.Date(2026, 4, 5, 9, 0, 0, 0, loc)
	next, err := EvaluateLazySync(sunday, sunday, NewFixedClock(sunday))
	if err != nil {
		t.Fatalf("EvaluateLazySync returned error on sunday: %v", err)
	}
	wantNextWeek := time.Date(2026, 4, 12, 0, 0, 0, 0, loc)
	if !next.WeeklyRolloverTo.Equal(wantNextWeek) {
		t.Fatalf("expected next-week Sunday rollover %v, got %v", wantNextWeek, next.WeeklyRolloverTo)
	}
}
