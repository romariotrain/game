package sim

import "math/rand"

// SimulateDay runs one day of gameplay: quests + battles.
func SimulateDay(player *PlayerState, arch Archetype, enemies []EnemyDef, rng *rand.Rand) DaySnapshot {
	player.DayNumber++
	questsToday := 0
	expToday := 0
	battlesToday := 0
	winsToday := 0

	// Complete quests
	for q := 0; q < arch.QuestsPerDay; q++ {
		questEXP, _ := SimulateQuest(player, arch, rng)
		questsToday++
		expToday += questEXP
	}

	// Update streak
	if questsToday > 0 {
		player.CurrentStreak++
	} else {
		player.CurrentStreak = 0
	}

	// Fight battles if archetype wants to and attempts are available
	if arch.FightsWhenPossible {
		for player.Attempts > 0 {
			enemy := pickNextEnemy(player, enemies)
			if enemy == nil {
				break // all enemies defeated
			}

			player.Attempts--
			player.TotalBattles++
			battlesToday++

			outcome := SimulateBattle(
				player.STR, player.AGI, player.INT, player.STA,
				*enemy,
				rng,
			)

			if outcome.Win {
				player.TotalBattleWins++
				winsToday++
				player.DefeatedEnemyIDs[enemy.Index] = true
				// Advance zone if boss defeated
				updateZone(player, enemies)
			} else {
				player.TotalBattleLosses++
			}
		}
	}

	totalWinRate := 0.0
	if player.TotalBattles > 0 {
		totalWinRate = float64(player.TotalBattleWins) / float64(player.TotalBattles) * 100
	}

	return DaySnapshot{
		Day:          player.DayNumber,
		Level:        player.OverallLevel(),
		STR:          player.STR,
		AGI:          player.AGI,
		INT:          player.INT,
		STA:          player.STA,
		Zone:         player.CurrentZone,
		QuestsToday:  questsToday,
		EXPToday:     expToday,
		BattlesToday: battlesToday,
		WinsToday:    winsToday,
		TotalWinRate: totalWinRate,
		Attempts:     player.Attempts,
	}
}

// pickNextEnemy returns the next undefeated enemy in the current zone.
// Transition enemy first, then easiest remaining regulars, then boss.
func pickNextEnemy(player *PlayerState, enemies []EnemyDef) *EnemyDef {
	var undefeatedRegularIDs []int
	bossID := -1

	for i := range enemies {
		e := enemies[i]
		if e.Zone != player.CurrentZone || player.DefeatedEnemyIDs[e.Index] {
			continue
		}
		if e.IsBoss {
			bossID = i
			continue
		}
		undefeatedRegularIDs = append(undefeatedRegularIDs, i)
	}

	if len(undefeatedRegularIDs) == 0 {
		if bossID >= 0 {
			return &enemies[bossID]
		}
		return nil
	}

	// Keep zone opening scary-but-real: first fight transition enemy.
	transitionID := -1
	for _, id := range undefeatedRegularIDs {
		if !IsZoneTransitionEnemy(enemies[id]) {
			continue
		}
		if transitionID == -1 || enemies[id].ExpectedMinLevel < enemies[transitionID].ExpectedMinLevel {
			transitionID = id
		}
	}
	if transitionID >= 0 {
		return &enemies[transitionID]
	}

	// After transition, route to easier remaining regular enemies first.
	bestID := undefeatedRegularIDs[0]
	for _, id := range undefeatedRegularIDs[1:] {
		if isEasierRegular(enemies[id], enemies[bestID]) {
			bestID = id
		}
	}
	return &enemies[bestID]
}

func regularDifficultyPriority(enemy EnemyDef) int {
	switch EnemyWinRateBand(enemy).Label {
	case "Лёгкий":
		return 0
	case "Нормальный":
		return 1
	case "Сложный":
		return 2
	case "Элитка":
		return 3
	case "Переход":
		return 4
	default:
		return 2
	}
}

func isEasierRegular(a, b EnemyDef) bool {
	pa := regularDifficultyPriority(a)
	pb := regularDifficultyPriority(b)
	if pa != pb {
		return pa < pb
	}
	if a.ExpectedMinLevel != b.ExpectedMinLevel {
		return a.ExpectedMinLevel < b.ExpectedMinLevel
	}
	return a.Index < b.Index
}

// updateZone checks if zone boss is defeated and advances zone.
func updateZone(player *PlayerState, enemies []EnemyDef) {
	for _, e := range enemies {
		if e.Zone == player.CurrentZone && e.IsBoss && player.DefeatedEnemyIDs[e.Index] {
			// Boss defeated — advance zone
			nextZone := player.CurrentZone + 1
			// Check if next zone has enemies
			hasEnemies := false
			for _, ne := range enemies {
				if ne.Zone == nextZone {
					hasEnemies = true
					break
				}
			}
			if hasEnemies {
				player.CurrentZone = nextZone
			}
			return
		}
	}
}

// RunProgression simulates the full progression for N days.
// Returns snapshots at key intervals.
func RunProgression(cfg SimConfig) ([]DaySnapshot, *PlayerState) {
	rng := rand.New(rand.NewSource(cfg.Seed))
	player := NewPlayerState()
	enemies := cfg.Enemies
	if len(enemies) == 0 {
		enemies = GetPresetEnemies()
	}

	snapshots := make([]DaySnapshot, 0, cfg.Days)

	for day := 1; day <= cfg.Days; day++ {
		snap := SimulateDay(player, cfg.Archetype, enemies, rng)
		snapshots = append(snapshots, snap)
	}

	return snapshots, player
}

// RunProgressionMultiple runs multiple independent simulations and averages results.
func RunProgressionMultiple(cfg SimConfig, runs int) []DaySnapshot {
	// Aggregate: sum all snapshots per day, then average
	type dayAccum struct {
		Level        float64
		STR          float64
		AGI          float64
		INT          float64
		STA          float64
		Zone         float64
		QuestsToday  float64
		EXPToday     float64
		BattlesToday float64
		WinsToday    float64
		TotalWinRate float64
	}

	accum := make([]dayAccum, cfg.Days)

	for r := 0; r < runs; r++ {
		runCfg := cfg
		runCfg.Seed = cfg.Seed + int64(r)
		snapshots, _ := RunProgression(runCfg)

		for i, s := range snapshots {
			accum[i].Level += float64(s.Level)
			accum[i].STR += float64(s.STR)
			accum[i].AGI += float64(s.AGI)
			accum[i].INT += float64(s.INT)
			accum[i].STA += float64(s.STA)
			accum[i].Zone += float64(s.Zone)
			accum[i].QuestsToday += float64(s.QuestsToday)
			accum[i].EXPToday += float64(s.EXPToday)
			accum[i].BattlesToday += float64(s.BattlesToday)
			accum[i].WinsToday += float64(s.WinsToday)
			accum[i].TotalWinRate += s.TotalWinRate
		}
	}

	n := float64(runs)
	result := make([]DaySnapshot, cfg.Days)
	for i, a := range accum {
		result[i] = DaySnapshot{
			Day:          i + 1,
			Level:        int(a.Level / n),
			STR:          int(a.STR / n),
			AGI:          int(a.AGI / n),
			INT:          int(a.INT / n),
			STA:          int(a.STA / n),
			Zone:         int(a.Zone / n),
			QuestsToday:  int(a.QuestsToday / n),
			EXPToday:     int(a.EXPToday / n),
			BattlesToday: int(a.BattlesToday / n),
			WinsToday:    int(a.WinsToday / n),
			TotalWinRate: a.TotalWinRate / n,
		}
	}
	return result
}

// CheckProgressionTimeline verifies zone arrival against target days.
func CheckProgressionTimeline(snapshots []DaySnapshot) []ProgressionCheck {
	targets := []struct {
		Zone int
		Days int
	}{
		{1, 0},  // Start zone (always met)
		{2, 7},  // Zone 2 by day 7
		{3, 21}, // Zone 3 by day 21
	}

	checks := make([]ProgressionCheck, len(targets))
	for i, t := range targets {
		checks[i] = ProgressionCheck{
			TargetZone: t.Zone,
			TargetDays: t.Days,
			ActualDays: -1,
			Met:        false,
		}

		for _, snap := range snapshots {
			if snap.Zone >= t.Zone {
				checks[i].ActualDays = snap.Day
				checks[i].Met = snap.Day <= t.Days || t.Days == 0
				break
			}
		}
	}

	return checks
}
