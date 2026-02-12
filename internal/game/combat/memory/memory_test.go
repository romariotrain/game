package memory

import (
	"math/rand"
	"testing"

	"solo-leveling/internal/models"
)

type fixedRNG struct {
	floatValues []float64
	index       int
}

func (r *fixedRNG) Float64() float64 {
	if len(r.floatValues) == 0 {
		return 0
	}
	v := r.floatValues[r.index%len(r.floatValues)]
	r.index++
	return v
}

func (r *fixedRNG) Perm(n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = i
	}
	return out
}

func TestGenerateShownCells_UniqueAndRange(t *testing.T) {
	grid := 6
	count := 10
	rng := rand.New(rand.NewSource(42))

	shown, err := GenerateShownCells(grid, count, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shown) != count {
		t.Fatalf("expected %d cells, got %d", count, len(shown))
	}

	seen := make(map[int]bool, len(shown))
	for _, idx := range shown {
		if idx < 0 || idx >= grid*grid {
			t.Fatalf("index out of bounds: %d", idx)
		}
		if seen[idx] {
			t.Fatalf("duplicate index: %d", idx)
		}
		seen[idx] = true
	}
}

func TestCellsToShow_IntReductionAndMin(t *testing.T) {
	enemy := models.Enemy{Rank: models.RankA, Type: models.EnemyRegular}
	stats := Stats{INT: 9}
	got := CellsToShow(enemy, stats)
	// base 14 - (9/3=3) = 11
	if got != 11 {
		t.Fatalf("expected 11 cells, got %d", got)
	}

	easy := models.Enemy{Rank: models.RankE, Type: models.EnemyRegular}
	minGot := CellsToShow(easy, Stats{INT: 60})
	if minGot != 4 {
		t.Fatalf("expected min 4 cells, got %d", minGot)
	}
}

func TestCellsToShow_BossExtraDifficulty(t *testing.T) {
	enemy := models.Enemy{Rank: models.RankA, Type: models.EnemyBoss}
	stats := Stats{INT: 9}
	got := CellsToShow(enemy, stats)
	// base 14 + 2 - 3 + 3 = 16
	if got != 16 {
		t.Fatalf("expected 16 cells for boss, got %d", got)
	}
}

func TestComputeAccuracy(t *testing.T) {
	shown := []int{1, 2, 3, 4}
	selected := []int{2, 2, 5, 4}
	got := ComputeAccuracy(shown, selected)
	if got != 0.5 {
		t.Fatalf("expected accuracy 0.5, got %f", got)
	}
}

func TestComputePlayerDamage_WithAndWithoutCrit(t *testing.T) {
	stats := Stats{STR: 10, AGI: 20}
	accuracy := 0.5

	nonCritRng := &fixedRNG{floatValues: []float64{0.99}}
	damage, crit := ComputePlayerDamage(stats, accuracy, nonCritRng)
	if crit {
		t.Fatal("expected non-crit")
	}
	// base = 10 + 20 = 30, damage = 30 * 0.5 = 15
	if damage != 15 {
		t.Fatalf("expected damage 15, got %d", damage)
	}

	critRng := &fixedRNG{floatValues: []float64{0.01}}
	damage, crit = ComputePlayerDamage(stats, accuracy, critRng)
	if !crit {
		t.Fatal("expected crit")
	}
	// 15 * 1.5 = 22.5 -> 23
	if damage != 23 {
		t.Fatalf("expected crit damage 23, got %d", damage)
	}
}

func TestComputeEnemyDamage_DeterministicAndMinOne(t *testing.T) {
	enemy := models.Enemy{Attack: 20}
	stats := Stats{STA: 8}
	rng := &fixedRNG{floatValues: []float64{0.5}} // factor = 1.0

	damage := ComputeEnemyDamage(enemy, stats, 0.4, rng)
	// 20*1.0 - 8*0.25 = 18, low accuracy penalty -> 21.6 -> 22
	if damage != 22 {
		t.Fatalf("expected enemy damage 22, got %d", damage)
	}

	veryTanky := Stats{STA: 999}
	damage = ComputeEnemyDamage(models.Enemy{Attack: 1}, veryTanky, 1.0, rng)
	if damage != 1 {
		t.Fatalf("damage should be clamped to min 1, got %d", damage)
	}
}
