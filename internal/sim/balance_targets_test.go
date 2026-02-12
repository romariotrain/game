package sim

import (
	"math/rand"
	"testing"
)

func TestBalanceTargets_AnchorBandsAndRedLines(t *testing.T) {
	enemies := GetPresetEnemies()
	const runs = 2000
	const tolerance = 2.0 // Monte Carlo noise margin

	for _, enemy := range enemies {
		level := EnemyMidExpectedLevel(enemy)
		stats := StatsFromLevel(level)

		mc := MonteCarloAnalysis(
			stats[0], stats[1], stats[2], stats[3],
			enemy,
			runs,
			rand.New(rand.NewSource(int64(10000+enemy.Index))),
		)
		wr := mc.WinRate
		band := EnemyWinRateBand(enemy)

		if wr < band.Min-tolerance || wr > band.Max+tolerance {
			t.Fatalf("%s (idx=%d) winrate %.1f%% is outside target %.1f-%.1f%%",
				enemy.Name, enemy.Index, wr, band.Min, band.Max)
		}

		if (band.Label == "Лёгкий" || band.Label == "Нормальный" || band.Label == "Сложный" || band.Label == "Переход") && wr < 15.0-tolerance {
			t.Fatalf("%s (idx=%d) violates red line: ordinary enemy below 15%% (got %.1f%%)",
				enemy.Name, enemy.Index, wr)
		}
		if band.Label == "Лёгкий" && wr > 70.0+tolerance {
			t.Fatalf("%s (idx=%d) violates red line: easy enemy above 70%% (got %.1f%%)",
				enemy.Name, enemy.Index, wr)
		}
		if enemy.IsBoss && wr < 4.5-tolerance {
			t.Fatalf("%s (idx=%d) violates red line: boss below 4.5%% (got %.1f%%)",
				enemy.Name, enemy.Index, wr)
		}
	}
}

func TestBalanceTargets_ZoneTransitionEnemiesAreScaryButReal(t *testing.T) {
	enemies := GetPresetEnemies()
	const runs = 2000
	const minWR = 15.0
	const maxWR = 25.0
	const tolerance = 2.0

	for _, enemy := range enemies {
		if !IsZoneTransitionEnemy(enemy) {
			continue
		}

		stats := StatsFromLevel(TransitionEntryLevel(enemy))
		mc := MonteCarloAnalysis(
			stats[0], stats[1], stats[2], stats[3],
			enemy,
			runs,
			rand.New(rand.NewSource(int64(20000+enemy.Index))),
		)
		wr := mc.WinRate

		if wr < minWR-tolerance || wr > maxWR+tolerance {
			t.Fatalf("%s (idx=%d) transition winrate %.1f%% should be in %.1f-%.1f%%",
				enemy.Name, enemy.Index, wr, minWR, maxWR)
		}
		if wr < 12.0-tolerance {
			t.Fatalf("%s (idx=%d) transition enemy should not be below 12%% (got %.1f%%)",
				enemy.Name, enemy.Index, wr)
		}
	}
}
