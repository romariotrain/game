package sim

import (
	"math"
	"testing"
)

func nearlyEqual(got, want, eps float64) bool {
	return math.Abs(got-want) <= eps
}

func seq(values ...float64) func() float64 {
	i := 0
	return func() float64 {
		v := values[i%len(values)]
		i++
		return v
	}
}

func TestSimulatedAccuracy_AlwaysAboveSixtyPercent(t *testing.T) {
	for _, intStat := range []int{0, 5, 20, 50} {
		lowRoll := SimulatedAccuracy(intStat, func() float64 { return 0.0 })
		highRoll := SimulatedAccuracy(intStat, func() float64 { return 1.0 })

		if lowRoll <= 0.60 {
			t.Fatalf("accuracy should stay above 60%%, INT=%d lowRoll=%f", intStat, lowRoll)
		}
		if highRoll <= 0.60 {
			t.Fatalf("accuracy should stay above 60%%, INT=%d highRoll=%f", intStat, highRoll)
		}
		if highRoll > 0.90 {
			t.Fatalf("accuracy should respect 0.90 cap, INT=%d highRoll=%f", intStat, highRoll)
		}
	}
}

func TestSimulatedAccuracy_IncreasesWithINT(t *testing.T) {
	low := SimulatedAccuracy(2, func() float64 { return 0.5 })
	high := SimulatedAccuracy(30, func() float64 { return 0.5 })
	if high <= low {
		t.Fatalf("higher INT should increase accuracy, got low=%f high=%f", low, high)
	}
}

func TestComputePlayerDamage_ScalesWithAccuracy(t *testing.T) {
	low, _ := ComputePlayerDamage(20, 10, 0.61, seq(0.5, 0.99))
	high, _ := ComputePlayerDamage(20, 10, 0.90, seq(0.5, 0.99))
	if high <= low {
		t.Fatalf("higher accuracy must increase damage, got low=%d high=%d", low, high)
	}
}

func TestCritChanceAndCritMultiplier(t *testing.T) {
	if !nearlyEqual(CritChance(), 0.10, 1e-9) {
		t.Fatalf("crit chance should be fixed at 0.10, got %f", CritChance())
	}

	low := CritDamageMultiplier(0)
	high := CritDamageMultiplier(30)
	if high <= low {
		t.Fatalf("AGI should increase crit damage multiplier, got low=%f high=%f", low, high)
	}
	if !nearlyEqual(CritDamageMultiplier(1000), 2.0, 1e-9) {
		t.Fatalf("crit damage multiplier should cap at 2.0, got %f", CritDamageMultiplier(1000))
	}
}

func TestDamageMitigation_ExpectedBounds(t *testing.T) {
	if !nearlyEqual(DamageMitigation(0), 0.95, 1e-9) {
		t.Fatalf("mitigation at STA=0 should clamp to 0.95, got %f", DamageMitigation(0))
	}
	if !nearlyEqual(DamageMitigation(40), 0.5, 1e-9) {
		t.Fatalf("mitigation at STA=40 should be 0.5, got %f", DamageMitigation(40))
	}
	if !nearlyEqual(DamageMitigation(999), 0.35, 1e-9) {
		t.Fatalf("mitigation lower bound should clamp to 0.35, got %f", DamageMitigation(999))
	}
}

func TestCellsToShow_NoIntegerDivisionStepAtTwoINT(t *testing.T) {
	got := CellsToShow("A", false, 2)
	if got != 13 {
		t.Fatalf("expected smooth reduction at INT=2 (13 cells), got %d", got)
	}
}
