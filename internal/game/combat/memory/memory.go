package memory

import (
	"errors"
	"math"

	"solo-leveling/internal/models"
)

const (
	RegularGridSize = 6
	BossGridSize    = 8

	minCellsToShow = 4

	baseShowSeconds = 2.5
	minShowSeconds  = 2.0
	maxShowSeconds  = 4.0

	critMultiplier = 1.5

	enemyDamageMinFactor = 0.8
	enemyDamageMaxFactor = 1.2

	lowAccuracyPenaltyThreshold  = 0.5
	lowAccuracyPenaltyMultiplier = 1.2

	// Keep this as a constant so tuning does not leak into UI code.
	enableLowAccuracyPenalty = true
)

// Stats represents combat-relevant player stats.
type Stats struct {
	STR int
	AGI int
	INT int
	STA int
}

// RNG is a narrow random source interface for deterministic tests.
type RNG interface {
	Float64() float64
	Perm(n int) []int
}

// GridSize selects memory field size by enemy type.
func GridSize(enemy models.Enemy) int {
	if enemy.Type == models.EnemyBoss {
		return BossGridSize
	}
	return RegularGridSize
}

// BaseCellsByRank maps enemy rank to base highlighted cells.
func BaseCellsByRank(rank models.QuestRank) int {
	switch rank {
	case models.RankE:
		return 6
	case models.RankD:
		return 8
	case models.RankC:
		return 10
	case models.RankB:
		return 12
	case models.RankA:
		return 14
	case models.RankS:
		return 16
	default:
		return 8
	}
}

// CellsToShow calculates the number of cells shown for the current round.
func CellsToShow(enemy models.Enemy, stats Stats) int {
	baseCells := BaseCellsByRank(enemy.Rank)
	if enemy.Type == models.EnemyBoss {
		baseCells += 2
	}

	reduction := stats.INT / 3
	cells := baseCells - reduction
	if cells < minCellsToShow {
		cells = minCellsToShow
	}
	if enemy.Type == models.EnemyBoss {
		cells += 3
	}

	maxCells := GridSize(enemy) * GridSize(enemy)
	if cells > maxCells {
		cells = maxCells
	}
	return cells
}

// TimeToShow returns highlight phase duration in milliseconds.
func TimeToShow(stats Stats) int {
	seconds := baseShowSeconds + float64(stats.INT)*0.05
	if seconds < minShowSeconds {
		seconds = minShowSeconds
	}
	if seconds > maxShowSeconds {
		seconds = maxShowSeconds
	}
	return int(math.Round(seconds * 1000))
}

// GenerateShownCells generates unique cell indices [0, grid*grid).
func GenerateShownCells(grid, count int, rng RNG) ([]int, error) {
	if grid <= 0 {
		return nil, errors.New("grid must be positive")
	}
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}
	if rng == nil {
		return nil, errors.New("rng is nil")
	}

	total := grid * grid
	if count > total {
		return nil, errors.New("count exceeds cell count")
	}

	perm := rng.Perm(total)
	out := make([]int, count)
	copy(out, perm[:count])
	return out, nil
}

// CorrectClicks counts unique intersections between shown and selected.
func CorrectClicks(shown, selected []int) int {
	if len(shown) == 0 || len(selected) == 0 {
		return 0
	}

	shownSet := make(map[int]struct{}, len(shown))
	for _, idx := range shown {
		shownSet[idx] = struct{}{}
	}

	selectedSet := make(map[int]struct{}, len(selected))
	for _, idx := range selected {
		selectedSet[idx] = struct{}{}
	}

	correct := 0
	for idx := range selectedSet {
		if _, ok := shownSet[idx]; ok {
			correct++
		}
	}
	return correct
}

// ComputeAccuracy returns ratio 0..1 based on selected vs shown intersection.
func ComputeAccuracy(shown, selected []int) float64 {
	if len(shown) == 0 {
		return 0
	}
	acc := float64(CorrectClicks(shown, selected)) / float64(len(shown))
	if acc < 0 {
		return 0
	}
	if acc > 1 {
		return 1
	}
	return acc
}

// PlayerHP returns base HP: 100 + STA*12.
func PlayerHP(sta int) int {
	if sta < 0 {
		sta = 0
	}
	return 100 + sta*12
}

// CritChance returns crit probability from AGI (1.5% per level).
func CritChance(agi int) float64 {
	if agi < 0 {
		return 0
	}
	chance := float64(agi) * 0.015
	if chance > 1 {
		return 1
	}
	return chance
}

// BasePlayerDamage returns damage before accuracy multiplier.
func BasePlayerDamage(str int) int {
	return 10 + str*2
}

// ComputePlayerDamage applies accuracy and crit to outgoing damage.
func ComputePlayerDamage(stats Stats, accuracy float64, rng RNG) (int, bool) {
	if accuracy < 0 {
		accuracy = 0
	}
	if accuracy > 1 {
		accuracy = 1
	}

	raw := float64(BasePlayerDamage(stats.STR)) * accuracy
	isCrit := false
	if raw > 0 && rng != nil && rng.Float64() < CritChance(stats.AGI) {
		raw *= critMultiplier
		isCrit = true
	}

	damage := int(math.Round(raw))
	if damage < 0 {
		damage = 0
	}
	return damage, isCrit
}

// ComputeEnemyDamage calculates incoming damage with variance and STA mitigation.
func ComputeEnemyDamage(enemy models.Enemy, stats Stats, accuracy float64, rng RNG) int {
	factor := 1.0
	if rng != nil {
		factor = enemyDamageMinFactor + rng.Float64()*(enemyDamageMaxFactor-enemyDamageMinFactor)
	}

	raw := float64(enemy.Attack) * factor
	raw -= float64(stats.STA) * 0.25

	if enableLowAccuracyPenalty && accuracy < lowAccuracyPenaltyThreshold {
		raw *= lowAccuracyPenaltyMultiplier
	}

	damage := int(math.Round(raw))
	if damage < 1 {
		damage = 1
	}
	return damage
}
