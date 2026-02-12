package sim

import (
	"math/rand"
	"testing"
)

func TestAutoTune_HitsBandsAndRedLines(t *testing.T) {
	base := GetPresetEnemies()
	opts := DefaultAutoTuneOptions(42)
	opts.RunsPerEval = 200
	opts.Iterations = 9
	tuned, _ := AutoTuneEnemies(base, opts)

	const runs = 1200
	const tolerance = 3.0
	const bossFloor = 4.5

	for _, enemy := range tuned {
		level := EnemyMidExpectedLevel(enemy)
		if IsZoneTransitionEnemy(enemy) {
			level = TransitionEntryLevel(enemy)
		}
		stats := StatsFromLevel(level)
		mc := MonteCarloAnalysis(
			stats[0], stats[1], stats[2], stats[3],
			enemy,
			runs,
			rand.New(rand.NewSource(int64(10000+enemy.Index))),
		)

		band := EnemyWinRateBand(enemy)
		if mc.WinRate < band.Min-tolerance || mc.WinRate > band.Max+tolerance {
			t.Fatalf("%s (idx=%d role=%s) winrate %.1f%% outside target %.1f-%.1f%%",
				enemy.Name, enemy.Index, enemy.Role, mc.WinRate, band.Min, band.Max)
		}

		switch enemy.Role {
		case "TRANSITION", "TRANSITION_ELITE", "NORMAL", "HARD", "EASY":
			if mc.WinRate < 15.0-tolerance {
				t.Fatalf("%s (idx=%d) ordinary role below 15%% (%.1f%%)",
					enemy.Name, enemy.Index, mc.WinRate)
			}
		}
		if enemy.Role == "EASY" && mc.WinRate > 70.0+tolerance {
			t.Fatalf("%s (idx=%d) easy role above 70%% (%.1f%%)",
				enemy.Name, enemy.Index, mc.WinRate)
		}
		if enemy.IsBoss && mc.WinRate < bossFloor-tolerance {
			t.Fatalf("%s (idx=%d) boss below %.1f%% (%.1f%%)",
				enemy.Name, enemy.Index, bossFloor, mc.WinRate)
		}
	}
}

func TestEnemyCatalog_5Zones10EnemiesAndBosses(t *testing.T) {
	enemies := GetPresetEnemies()
	if len(enemies) != 50 {
		t.Fatalf("expected 50 enemies, got %d", len(enemies))
	}

	zoneCount := map[int]int{}
	zoneBosses := map[int]int{}
	for _, enemy := range enemies {
		zoneCount[enemy.Zone]++
		if enemy.IsBoss {
			zoneBosses[enemy.Zone]++
		}
	}

	for zone := 1; zone <= 5; zone++ {
		if zoneCount[zone] != 10 {
			t.Fatalf("zone %d expected 10 enemies, got %d", zone, zoneCount[zone])
		}
		if zoneBosses[zone] != 1 {
			t.Fatalf("zone %d expected 1 boss, got %d", zone, zoneBosses[zone])
		}
	}
}
