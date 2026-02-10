package memory

import (
	"errors"
	"math/rand"

	"solo-leveling/internal/models"
)

// Stats represents player stat levels for difficulty scaling.
type Stats struct {
	STR int
	AGI int
	INT int
	STA int
}

// Difficulty defines tactical memory parameters for a battle.
type Difficulty struct {
	GridSize      int
	PatternLength int
	ShowTimeMs    int
	AllowedErrors int
}

const (
	minGridSize      = 3
	maxGridSize      = 6
	minPatternLength = 4
	maxPatternLength = 9
	minShowTimeMs    = 800
	maxShowTimeMs    = 3500
	minErrors        = 0
	maxErrors        = 3
)

// DifficultyFor returns difficulty parameters based on enemy rank and player stats.
func DifficultyFor(rank models.QuestRank, stats Stats) Difficulty {
	base := baseDifficulty(rank)

	// INT: more show time, slightly shorter pattern
	intBonusTime := stats.INT * 30
	intPatternReduction := stats.INT / 15

	base.ShowTimeMs += intBonusTime
	base.PatternLength -= intPatternReduction

	// AGI: +1 allowed error per 20 levels
	base.AllowedErrors += stats.AGI / 20

	// Clamp to bounds
	base.GridSize = clamp(base.GridSize, minGridSize, maxGridSize)
	base.PatternLength = clamp(base.PatternLength, minPatternLength, maxPatternLength)
	base.ShowTimeMs = clamp(base.ShowTimeMs, minShowTimeMs, maxShowTimeMs)
	base.AllowedErrors = clamp(base.AllowedErrors, minErrors, maxErrors)

	return base
}

func baseDifficulty(rank models.QuestRank) Difficulty {
	switch rank {
	case models.RankE:
		return Difficulty{GridSize: 3, PatternLength: 4, ShowTimeMs: 2600, AllowedErrors: 2}
	case models.RankD:
		return Difficulty{GridSize: 3, PatternLength: 5, ShowTimeMs: 2400, AllowedErrors: 2}
	case models.RankC:
		return Difficulty{GridSize: 4, PatternLength: 6, ShowTimeMs: 2200, AllowedErrors: 1}
	case models.RankB:
		return Difficulty{GridSize: 4, PatternLength: 7, ShowTimeMs: 2000, AllowedErrors: 1}
	case models.RankA:
		return Difficulty{GridSize: 5, PatternLength: 8, ShowTimeMs: 1800, AllowedErrors: 1}
	case models.RankS:
		return Difficulty{GridSize: 6, PatternLength: 9, ShowTimeMs: 1600, AllowedErrors: 0}
	default:
		return Difficulty{GridSize: 3, PatternLength: 4, ShowTimeMs: 2600, AllowedErrors: 2}
	}
}

// GeneratePattern creates a valid route (adjacent steps) without consecutive repeats.
// The pattern is a sequence of cell indices for a gridSize x gridSize grid.
func GeneratePattern(gridSize, length int, rng *rand.Rand) ([]int, error) {
	if gridSize < minGridSize || gridSize > maxGridSize {
		return nil, errors.New("grid size out of bounds")
	}
	if length < 1 {
		return nil, errors.New("pattern length must be positive")
	}
	if rng == nil {
		return nil, errors.New("rng is nil")
	}

	cells := gridSize * gridSize
	start := rng.Intn(cells)
	pattern := make([]int, 0, length)
	pattern = append(pattern, start)

	current := start
	for len(pattern) < length {
		n := neighbors(current, gridSize)
		if len(n) == 0 {
			return nil, errors.New("no neighbors for current cell")
		}
		next := n[rng.Intn(len(n))]
		// Prevent immediate repeats just in case
		if next == current {
			continue
		}
		pattern = append(pattern, next)
		current = next
	}

	return pattern, nil
}

func neighbors(idx, gridSize int) []int {
	x := idx % gridSize
	y := idx / gridSize

	var out []int
	if x > 0 {
		out = append(out, coordToIdx(x-1, y, gridSize))
	}
	if x < gridSize-1 {
		out = append(out, coordToIdx(x+1, y, gridSize))
	}
	if y > 0 {
		out = append(out, coordToIdx(x, y-1, gridSize))
	}
	if y < gridSize-1 {
		out = append(out, coordToIdx(x, y+1, gridSize))
	}
	return out
}

func coordToIdx(x, y, gridSize int) int {
	return y*gridSize + x
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
