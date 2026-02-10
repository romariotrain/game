package memory

import (
	"math/rand"
	"testing"

	"solo-leveling/internal/models"
)

func TestGeneratePattern_ValidRoute(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	grid := 5
	length := 9
	pattern, err := GeneratePattern(grid, length, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pattern) != length {
		t.Fatalf("expected length %d, got %d", length, len(pattern))
	}
	for i := 0; i < len(pattern); i++ {
		if pattern[i] < 0 || pattern[i] >= grid*grid {
			t.Fatalf("index out of bounds at %d: %d", i, pattern[i])
		}
		if i > 0 {
			if pattern[i] == pattern[i-1] {
				t.Fatalf("consecutive repeat at %d", i)
			}
			if !isAdjacent(pattern[i-1], pattern[i], grid) {
				t.Fatalf("non-adjacent step at %d: %d -> %d", i, pattern[i-1], pattern[i])
			}
		}
	}
}

func TestDifficulty_MonotonicByRank(t *testing.T) {
	stats := Stats{}
	ranks := []models.QuestRank{models.RankE, models.RankD, models.RankC, models.RankB, models.RankA, models.RankS}

	prev := DifficultyFor(ranks[0], stats)
	for i := 1; i < len(ranks); i++ {
		next := DifficultyFor(ranks[i], stats)
		if next.GridSize < prev.GridSize {
			t.Fatalf("grid size should not decrease: %v -> %v", prev.GridSize, next.GridSize)
		}
		if next.PatternLength < prev.PatternLength {
			t.Fatalf("pattern length should not decrease: %v -> %v", prev.PatternLength, next.PatternLength)
		}
		if next.ShowTimeMs > prev.ShowTimeMs {
			t.Fatalf("show time should not increase: %v -> %v", prev.ShowTimeMs, next.ShowTimeMs)
		}
		prev = next
	}
}

func TestDifficulty_ShowTimeBounds(t *testing.T) {
	stats := Stats{INT: 999}
	d := DifficultyFor(models.RankE, stats)
	if d.ShowTimeMs < minShowTimeMs || d.ShowTimeMs > maxShowTimeMs {
		t.Fatalf("show time out of bounds: %d", d.ShowTimeMs)
	}
}

func isAdjacent(a, b, grid int) bool {
	ax, ay := a%grid, a/grid
	bx, by := b%grid, b/grid
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	return (dx == 1 && dy == 0) || (dx == 0 && dy == 1)
}
