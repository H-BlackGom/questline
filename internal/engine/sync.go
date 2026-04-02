package engine

import (
	"fmt"
	"time"
)

type SyncResult struct {
	LogicalDate      time.Time
	EvaluationDate   *time.Time
	WeeklyRolloverTo time.Time
}

func EvaluateLazySync(lastSynced time.Time, now time.Time, clock Clock) (*SyncResult, error) {
	if clock == nil {
		clock = SystemClock{}
	}
	if now.IsZero() {
		now = clock.Now()
	}
	if now.IsZero() {
		return nil, fmt.Errorf("now is required for lazy sync evaluation")
	}

	if lastSynced.IsZero() {
		lastSynced = now
	}

	logicalNow := LogicalDate(now)
	logicalLastSynced := LogicalDate(lastSynced.In(now.Location()))

	result := &SyncResult{
		LogicalDate:      logicalNow,
		WeeklyRolloverTo: nextSunday(logicalNow),
	}

	if logicalLastSynced.Before(logicalNow) {
		evaluationDate := logicalLastSynced
		result.EvaluationDate = &evaluationDate
	}

	return result, nil
}

func nextSunday(from time.Time) time.Time {
	base := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	daysUntilSunday := (7 - int(base.Weekday())) % 7
	if daysUntilSunday == 0 {
		daysUntilSunday = 7
	}
	return base.AddDate(0, 0, daysUntilSunday)
}
