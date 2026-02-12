package sim

import (
	"math/rand"
	"testing"
)

func TestMonteCarloAnalysis_HigherINTImprovesWinRate(t *testing.T) {
	enemy := EnemyDef{
		Name:   "Test Enemy",
		Rank:   "B",
		HP:     280,
		Attack: 24,
		AGI:    12,
		INT:    14,
	}

	const runs = 2000

	lowINT := MonteCarloAnalysis(
		12, 12, 2, 12,
		enemy,
		runs,
		rand.New(rand.NewSource(42)),
	)
	highINT := MonteCarloAnalysis(
		12, 12, 28, 12,
		enemy,
		runs,
		rand.New(rand.NewSource(42)),
	)

	if lowINT.AvgAccuracy <= 60.0 {
		t.Fatalf("simulated accuracy should stay above 60%%, got %.2f%%", lowINT.AvgAccuracy)
	}
	if highINT.AvgAccuracy <= lowINT.AvgAccuracy {
		t.Fatalf("higher INT should increase average accuracy, got low=%.2f high=%.2f", lowINT.AvgAccuracy, highINT.AvgAccuracy)
	}

	if highINT.WinRate <= lowINT.WinRate {
		t.Fatalf("expected higher INT to increase winrate, got low=%.2f high=%.2f", lowINT.WinRate, highINT.WinRate)
	}
}
