// Package sim provides a headless balance simulator.
// Formulas here are tuned for simulation and balancing — no Fyne, no DB.
package sim

import "math"

// ──────────────────────────────────────────────
// Quest EXP
// ──────────────────────────────────────────────

// CalculateQuestEXP = round(minutes*0.6 + effort*4 + friction*3), min 1.
func CalculateQuestEXP(minutes, effort, friction int) int {
	if minutes < 0 {
		minutes = 0
	}
	if effort < 1 {
		effort = 1
	}
	if effort > 5 {
		effort = 5
	}
	if friction < 1 {
		friction = 1
	}
	if friction > 3 {
		friction = 3
	}
	base := float64(minutes) * 0.6
	exp := int(math.Round(base + float64(effort*4+friction*3)))
	if exp < 1 {
		return 1
	}
	return exp
}

// RankFromEXP maps quest EXP to rank string.
func RankFromEXP(exp int) string {
	switch {
	case exp <= 10:
		return "E"
	case exp <= 18:
		return "D"
	case exp <= 28:
		return "C"
	case exp <= 40:
		return "B"
	case exp <= 55:
		return "A"
	default:
		return "S"
	}
}

// BaseEXPForRank returns EXP awarded to stat on quest completion.
func BaseEXPForRank(rank string) int {
	switch rank {
	case "E":
		return 20
	case "D":
		return 40
	case "C":
		return 70
	case "B":
		return 120
	case "A":
		return 200
	case "S":
		return 350
	default:
		return 0
	}
}

// ExpForLevel returns EXP required to advance from `level` to `level+1`.
func ExpForLevel(level int) int {
	return 50 + (level-1)*30
}

// AttemptsForQuestEXP returns battle attempts awarded by quest EXP.
func AttemptsForQuestEXP(exp int) int {
	switch {
	case exp < 15:
		return 1
	case exp <= 30:
		return 2
	default:
		return 3
	}
}

const MaxAttempts = 8

// ──────────────────────────────────────────────
// Combat formulas (balance simulator)
// ──────────────────────────────────────────────

const (
	RegularGridSize = 6
	BossGridSize    = 8

	minCellsToShow = 4

	baseShowSeconds = 2.5
	minShowSeconds  = 2.0
	maxShowSeconds  = 4.0

	weaponBaseDamage = 20.0
	playerCritChance = 0.10

	minSimulatedAccuracy  = 0.61
	maxSimulatedAccuracy  = 0.86
	baseSimulatedAccuracy = 0.64
	accuracyINTScale      = 45.0
	accuracyRandomSpread  = 0.02

	playerDamageMinFactor = 0.90
	playerDamageMaxFactor = 1.10
	enemyDamageMinFactor  = 0.90
	enemyDamageMaxFactor  = 1.10

	enemyScaleMinFactor   = 0.90
	enemyScaleMaxFactor   = 1.10
	enemyScaleLevelBuffer = 3.0
)

func clampf(x, minVal, maxVal float64) float64 {
	if x < minVal {
		return minVal
	}
	if x > maxVal {
		return maxVal
	}
	return x
}

// GridSize returns grid dimension for enemy.
func GridSize(isBoss bool) int {
	if isBoss {
		return BossGridSize
	}
	return RegularGridSize
}

// BaseCellsByRank maps rank string to base highlighted cells.
func BaseCellsByRank(rank string) int {
	switch rank {
	case "E":
		return 6
	case "D":
		return 8
	case "C":
		return 10
	case "B":
		return 12
	case "A":
		return 14
	case "S":
		return 16
	default:
		return 8
	}
}

// CellsToShow calculates number of highlighted cells.
func CellsToShow(rank string, isBoss bool, intLevel int) int {
	baseCells := BaseCellsByRank(rank)
	if isBoss {
		baseCells += 2
	}

	if intLevel < 0 {
		intLevel = 0
	}
	reduction := float64(intLevel) / 3.0
	cells := int(math.Round(float64(baseCells) - reduction))
	if cells < minCellsToShow {
		cells = minCellsToShow
	}
	if isBoss {
		cells += 3
	}

	maxCells := GridSize(isBoss) * GridSize(isBoss)
	if cells > maxCells {
		cells = maxCells
	}
	return cells
}

// PlayerHP = 100 + STA*12
func PlayerHP(sta int) int {
	if sta < 0 {
		sta = 0
	}
	return 100 + sta*12
}

// SimulatedAccuracy returns per-round correct-cells ratio in the simulator.
// INT is the only stat that affects this value.
func SimulatedAccuracy(intStat int, rngFloat func() float64) float64 {
	if intStat < 0 {
		intStat = 0
	}
	jitter := 0.0
	if rngFloat != nil {
		jitter = (rngFloat()*2.0 - 1.0) * accuracyRandomSpread
	}
	progress := 1.0 - math.Exp(-float64(intStat)/accuracyINTScale)
	acc := baseSimulatedAccuracy + (maxSimulatedAccuracy-baseSimulatedAccuracy)*progress + jitter
	return clampf(acc, minSimulatedAccuracy, maxSimulatedAccuracy)
}

// CritChance is fixed in simulation; AGI affects only crit damage multiplier.
func CritChance() float64 {
	return playerCritChance
}

// CritDamageMultiplier scales only from AGI and applies only to player crits.
func CritDamageMultiplier(agi int) float64 {
	if agi < 0 {
		agi = 0
	}
	return clampf(1.20+0.010*float64(agi), 1.20, 2.00)
}

// BasePlayerDamage applies STR scaling before accuracy/RNG/crit.
func BasePlayerDamage(str int) float64 {
	return weaponBaseDamage * (1 + 0.035*float64(str))
}

// ComputePlayerDamage applies accuracy scaling, variance and crit.
// Returns (damage, isCrit).
func ComputePlayerDamage(str, agi int, accuracy float64, rngFloat func() float64) (int, bool) {
	accuracy = clampf(accuracy, 0, 1)
	variance := playerDamageMinFactor + rngFloat()*(playerDamageMaxFactor-playerDamageMinFactor)
	raw := BasePlayerDamage(str) * accuracy * variance

	isCrit := false
	if raw > 0 && rngFloat() < CritChance() {
		raw *= CritDamageMultiplier(agi)
		isCrit = true
	}

	damage := int(math.Round(raw))
	if damage < 1 {
		damage = 1
	}
	return damage, isCrit
}

// DamageMitigation uses STA as a soft diminishing-returns damage reducer.
func DamageMitigation(sta int) float64 {
	if sta < 0 {
		sta = 0
	}
	mit := 1.0 - (float64(sta) / (float64(sta) + 40.0))
	return clampf(mit, 0.35, 0.95)
}

// ComputeEnemyDamage calculates incoming damage with variance and STA mitigation.
func ComputeEnemyDamage(enemyAttack, sta int, rngFloat func() float64) int {
	factor := enemyDamageMinFactor + rngFloat()*(enemyDamageMaxFactor-enemyDamageMinFactor)
	raw := float64(enemyAttack) * DamageMitigation(sta) * factor

	damage := int(math.Round(raw))
	if damage < 1 {
		damage = 1
	}
	return damage
}

// PlayerCombatLevel uses STR/STA progression backbone for enemy interpolation.
// AGI/INT stay combat modifiers and do not drive enemy scaling directly.
func PlayerCombatLevel(str, sta int) float64 {
	if str < 0 {
		str = 0
	}
	if sta < 0 {
		sta = 0
	}
	level := (float64(str) + float64(sta)) / 2.0
	if level < 1.0 {
		return 1.0
	}
	return level
}

// EnemyScaleFactor interpolates enemy power inside its expected level window.
func EnemyScaleFactor(enemy EnemyDef, playerLevel float64) float64 {
	minLvl := enemy.ExpectedMinLevel
	maxLvl := enemy.ExpectedMaxLevel
	if minLvl <= 0 || maxLvl <= 0 {
		return 1.0
	}
	if maxLvl < minLvl {
		maxLvl = minLvl
	}

	low := float64(minLvl) - enemyScaleLevelBuffer
	high := float64(maxLvl) + enemyScaleLevelBuffer
	if high <= low {
		return 1.0
	}

	pos := clampf((float64(playerLevel)-low)/(high-low), 0.0, 1.0)
	return enemyScaleMinFactor + pos*(enemyScaleMaxFactor-enemyScaleMinFactor)
}

// ──────────────────────────────────────────────
// Zone logic
// ──────────────────────────────────────────────

// ZoneForRank maps rank to zone.
func ZoneForRank(rank string) int {
	switch rank {
	case "E", "D":
		return 1
	case "C", "B":
		return 2
	case "A", "S":
		return 3
	default:
		return 1
	}
}
