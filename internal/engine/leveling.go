package engine

// LevelingResult contains the outcome of XP addition
type LevelingResult struct {
	LeveledUp    bool
	LevelsGained int
	OldLevel     int
	NewLevel     int
	OldXP        int
	NewXP        int
	RequiredXP   int
}

// AddXP adds XP to player and calculates level ups
// Fixed 50 XP per quest, formula: next_level_xp = 100 + (level * 50)
func AddXP(currentLevel, currentXP, xpAmount int) LevelingResult {
	result := LevelingResult{
		OldLevel: currentLevel,
		OldXP:    currentXP,
		NewLevel: currentLevel,
		NewXP:    currentXP,
	}

	if xpAmount <= 0 {
		return result
	}

	result.NewXP = currentXP + xpAmount
	result.NewLevel = currentLevel
	levelsGained := 0

	// Check for level ups
	for {
		requiredXP := 100 + (result.NewLevel * 50)
		if result.NewXP >= requiredXP {
			result.NewXP -= requiredXP
			result.NewLevel++
			levelsGained++
		} else {
			result.RequiredXP = requiredXP
			break
		}
	}

	result.LevelsGained = levelsGained
	result.LeveledUp = levelsGained > 0

	return result
}

// GetTitle returns title based on level
func GetTitle(level int) string {
	switch {
	case level >= 99:
		return "Guru"
	case level >= 50:
		return "Principal"
	case level >= 30:
		return "Lead"
	case level >= 20:
		return "Senior"
	case level >= 10:
		return "Junior"
	default:
		return "Intern"
	}
}

// GetRequiredXPForNextLevel calculates XP needed for next level
func GetRequiredXPForNextLevel(level int) int {
	return 100 + (level * 50)
}
