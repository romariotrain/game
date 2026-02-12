package sim

// WinRateBand describes an expected winrate range for one enemy archetype.
type WinRateBand struct {
	Label string
	Min   float64
	Max   float64
}

// StatsFromLevel builds equalized player stats for one synthetic level.
func StatsFromLevel(level int) [4]int {
	if level < 1 {
		level = 1
	}
	return [4]int{level, level, level, level}
}

// EnemyMidExpectedLevel returns midpoint of the enemy expected-level window.
func EnemyMidExpectedLevel(enemy EnemyDef) int {
	minLvl := enemy.ExpectedMinLevel
	maxLvl := enemy.ExpectedMaxLevel
	if minLvl <= 0 && maxLvl <= 0 {
		switch enemy.Zone {
		case 1:
			return 3
		case 2:
			return 10
		default:
			return 20
		}
	}
	if minLvl <= 0 {
		minLvl = maxLvl
	}
	if maxLvl <= 0 {
		maxLvl = minLvl
	}
	if maxLvl < minLvl {
		maxLvl = minLvl
	}
	return int((float64(minLvl) + float64(maxLvl)) / 2.0)
}

// EnemyWinRateBand maps each preset enemy to its target band.
func EnemyWinRateBand(enemy EnemyDef) WinRateBand {
	switch enemy.Index {
	case 0, 7, 13:
		return WinRateBand{Label: "Лёгкий", Min: 45, Max: 60}
	case 1, 6, 11:
		return WinRateBand{Label: "Нормальный", Min: 30, Max: 45}
	case 2, 12:
		return WinRateBand{Label: "Сложный", Min: 20, Max: 30}
	case 4, 9:
		return WinRateBand{Label: "Переход", Min: 15, Max: 25}
	case 5, 10:
		return WinRateBand{Label: "Элитка", Min: 12, Max: 20}
	case 3, 8, 14:
		return WinRateBand{Label: "Босс зоны", Min: 5, Max: 12}
	default:
		if enemy.IsBoss {
			return WinRateBand{Label: "Босс зоны", Min: 5, Max: 12}
		}
		return WinRateBand{Label: "Нормальный", Min: 30, Max: 45}
	}
}

// IsZoneTransitionEnemy marks the first enemy in zones 2 and 3.
func IsZoneTransitionEnemy(enemy EnemyDef) bool {
	return enemy.Index == 4 || enemy.Index == 9
}

// TransitionEntryLevel returns "just entered zone" level for transition checks.
func TransitionEntryLevel(enemy EnemyDef) int {
	lvl := enemy.ExpectedMinLevel - 1
	if lvl < 1 {
		lvl = 1
	}
	return lvl
}
