package engine

import "time"

const LogicalDayCutoffHour = 4

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type FixedClock struct {
	now time.Time
}

func NewFixedClock(now time.Time) FixedClock {
	return FixedClock{now: now}
}

func (c FixedClock) Now() time.Time {
	return c.now
}

type LogicalClock struct {
	clock Clock
}

func NewLogicalClock(clock Clock) LogicalClock {
	if clock == nil {
		clock = SystemClock{}
	}
	return LogicalClock{clock: clock}
}

func (c LogicalClock) Now() time.Time {
	return c.clock.Now()
}

func (c LogicalClock) LogicalDate() time.Time {
	return LogicalDate(c.clock.Now())
}

func LogicalDate(now time.Time) time.Time {
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := date.Add(time.Duration(LogicalDayCutoffHour) * time.Hour)
	if now.Before(cutoff) {
		return date.AddDate(0, 0, -1)
	}
	return date
}
