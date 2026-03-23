package engine

import (
	"strconv"
	"testing"
)

func TestAddXP(t *testing.T) {
	tests := []struct {
		name           string
		currentLevel   int
		currentXP      int
		xpAmount       int
		expectLeveled  bool
		expectNewLevel int
		expectNewXP    int
		expectLevels   int
	}{
		{
			name:           "No level up - partial XP",
			currentLevel:   1,
			currentXP:      0,
			xpAmount:       50,
			expectLeveled:  false,
			expectNewLevel: 1,
			expectNewXP:    50,
			expectLevels:   0,
		},
		{
			name:           "Single level up",
			currentLevel:   1,
			currentXP:      100,
			xpAmount:       50,
			expectLeveled:  true,
			expectNewLevel: 2,
			expectNewXP:    0,
			expectLevels:   1,
		},
		{
			name:           "Multiple level ups",
			currentLevel:   1,
			currentXP:      0,
			xpAmount:       400,
			expectLeveled:  true,
			expectNewLevel: 3,
			expectNewXP:    50, // 400 - 150 - 200 = 50
			expectLevels:   2,
		},
		{
			name:           "Zero XP - no change",
			currentLevel:   5,
			currentXP:      50,
			xpAmount:       0,
			expectLeveled:  false,
			expectNewLevel: 5,
			expectNewXP:    50,
			expectLevels:   0,
		},
		{
			name:           "Negative XP - rejected",
			currentLevel:   5,
			currentXP:      50,
			xpAmount:       -10,
			expectLeveled:  false,
			expectNewLevel: 5,
			expectNewXP:    50,
			expectLevels:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddXP(tt.currentLevel, tt.currentXP, tt.xpAmount)

			if result.LeveledUp != tt.expectLeveled {
				t.Errorf("LeveledUp = %v, want %v", result.LeveledUp, tt.expectLeveled)
			}
			if result.NewLevel != tt.expectNewLevel {
				t.Errorf("NewLevel = %d, want %d", result.NewLevel, tt.expectNewLevel)
			}
			if result.NewXP != tt.expectNewXP {
				t.Errorf("NewXP = %d, want %d", result.NewXP, tt.expectNewXP)
			}
			if result.LevelsGained != tt.expectLevels {
				t.Errorf("LevelsGained = %d, want %d", result.LevelsGained, tt.expectLevels)
			}
		})
	}
}

func TestGetTitle(t *testing.T) {
	tests := []struct {
		level int
		want  string
	}{
		{1, "Intern"},
		{5, "Intern"},
		{9, "Intern"},
		{10, "Junior"},
		{15, "Junior"},
		{19, "Junior"},
		{20, "Senior"},
		{29, "Senior"},
		{30, "Lead"},
		{49, "Lead"},
		{50, "Principal"},
		{98, "Principal"},
		{99, "Guru"},
		{100, "Guru"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetTitle(tt.level)
			if got != tt.want {
				t.Errorf("GetTitle(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestGetRequiredXPForNextLevel(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{1, 150},
		{2, 200},
		{3, 250},
		{5, 350},
		{10, 600},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.want), func(t *testing.T) {
			got := GetRequiredXPForNextLevel(tt.level)
			if got != tt.want {
				t.Errorf("GetRequiredXPForNextLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}
