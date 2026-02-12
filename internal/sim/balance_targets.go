package sim

import "strings"

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
		if enemy.Level > 0 {
			return enemy.Level
		}
		return 1
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

// EnemyWinRateBand returns target range from enemy metadata.
func EnemyWinRateBand(enemy EnemyDef) WinRateBand {
	minV := enemy.TargetWinRateMin
	maxV := enemy.TargetWinRateMax
	if minV <= 0 || maxV <= 0 || minV >= maxV {
		minV, maxV = 30, 45
	}
	return WinRateBand{
		Label: RoleLabel(enemy.Role),
		Min:   minV,
		Max:   maxV,
	}
}

func RoleLabel(role string) string {
	switch strings.ToUpper(role) {
	case "TRANSITION":
		return "Переход"
	case "TRANSITION_ELITE":
		return "Переход/Элитка"
	case "NORMAL":
		return "Нормальный"
	case "HARD":
		return "Сложный"
	case "EASY":
		return "Лёгкий"
	case "ELITE":
		return "Элитка"
	case "MINIBOSS":
		return "Мини-босс"
	case "BOSS":
		return "Босс зоны"
	default:
		return "Нормальный"
	}
}

func RolePriority(role string) int {
	switch strings.ToUpper(role) {
	case "EASY":
		return 0
	case "NORMAL":
		return 1
	case "HARD":
		return 2
	case "TRANSITION":
		return 3
	case "TRANSITION_ELITE":
		return 4
	case "ELITE":
		return 5
	case "MINIBOSS":
		return 6
	case "BOSS":
		return 7
	default:
		return 1
	}
}

// IsZoneTransitionEnemy marks zone-opening enemies.
func IsZoneTransitionEnemy(enemy EnemyDef) bool {
	return enemy.IsTransition
}

// TransitionEntryLevel returns "just entered zone" level for transition checks.
func TransitionEntryLevel(enemy EnemyDef) int {
	if enemy.Zone <= 1 {
		return EnemyMidExpectedLevel(enemy)
	}
	if enemy.ExpectedMinLevel > 0 {
		lvl := enemy.ExpectedMinLevel - 1
		if lvl < 1 {
			return 1
		}
		return lvl
	}
	if enemy.Level > 1 {
		return enemy.Level - 1
	}
	return 1
}

func maxZone(enemies []EnemyDef) int {
	mz := 1
	for _, e := range enemies {
		if e.Zone > mz {
			mz = e.Zone
		}
	}
	return mz
}
