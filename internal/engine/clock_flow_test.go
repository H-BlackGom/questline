package engine

import (
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

func TestLogicalDate(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before cutoff uses previous day",
			now:  time.Date(2026, 4, 3, 3, 59, 59, 0, loc),
			want: time.Date(2026, 4, 2, 0, 0, 0, 0, loc),
		},
		{
			name: "at cutoff uses current day",
			now:  time.Date(2026, 4, 3, 4, 0, 0, 0, loc),
			want: time.Date(2026, 4, 3, 0, 0, 0, 0, loc),
		},
		{
			name: "after cutoff uses current day",
			now:  time.Date(2026, 4, 3, 9, 30, 0, 0, loc),
			want: time.Date(2026, 4, 3, 0, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LogicalDate(tt.now); !got.Equal(tt.want) {
				t.Fatalf("LogicalDate(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestFourAMBoundary(t *testing.T) {
	loc := time.FixedZone("KST", 9*60*60)

	before := time.Date(2026, 4, 3, 3, 59, 59, 0, loc)
	at := time.Date(2026, 4, 3, 4, 0, 0, 0, loc)

	if got := LogicalDate(before); got.Day() != 2 {
		t.Fatalf("expected previous day before 04:00, got %v", got)
	}
	if got := LogicalDate(at); got.Day() != 3 {
		t.Fatalf("expected current day at 04:00:00, got %v", got)
	}
}

func TestFlowGrade(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		total     int
		want      domain.FlowStatus
	}{
		{name: "100 percent singularity", completed: 10, total: 10, want: domain.FlowStatusSingularity},
		{name: "80 percent burning", completed: 4, total: 5, want: domain.FlowStatusBurning},
		{name: "79 percent smooth", completed: 79, total: 100, want: domain.FlowStatusSmooth},
		{name: "50 percent smooth", completed: 1, total: 2, want: domain.FlowStatusSmooth},
		{name: "49 percent hazy", completed: 49, total: 100, want: domain.FlowStatusHazy},
		{name: "0 percent hazy when dailies exist", completed: 0, total: 3, want: domain.FlowStatusHazy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EvaluateFlowStatus(tt.total, tt.completed); got != tt.want {
				t.Fatalf("EvaluateFlowStatus(%d, %d) = %q, want %q", tt.total, tt.completed, got, tt.want)
			}
		})
	}
}

func TestNoDailyDefaultsToSmooth(t *testing.T) {
	if got := EvaluateFlowStatus(0, 0); got != domain.FlowStatusSmooth {
		t.Fatalf("EvaluateFlowStatus(0, 0) = %q, want %q", got, domain.FlowStatusSmooth)
	}
}
