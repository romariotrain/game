package sim

import (
	"fmt"
	"math/rand"
	"strings"
)

// FullReport generates the complete simulation report as a string.
func FullReport(seed int64) string {
	return FullReportWithEnemies(seed, GetPresetEnemies())
}

// FullReportWithEnemies generates the complete simulation report for a specific enemy set.
func FullReportWithEnemies(seed int64, enemies []EnemyDef) string {
	var sb strings.Builder
	rng := rand.New(rand.NewSource(seed))
	if len(enemies) == 0 {
		enemies = GetPresetEnemies()
	}

	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           SOLO LEVELING — BALANCE SIMULATOR                 ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n\n")

	// ── 1. EXP Economy Analysis ──────────────────────────────
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  1. EXP ECONOMY ANALYSIS (1000 random quests per archetype)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, arch := range DefaultArchetypes() {
		econ := EXPEconomyAnalysis(arch, 1000, rng)
		sb.WriteString(fmt.Sprintf("  ▸ %s\n", arch.Name))
		sb.WriteString(fmt.Sprintf("    Avg EXP: %.1f | Min: %d | Max: %d\n", econ.AvgEXP, econ.MinEXP, econ.MaxEXP))
		sb.WriteString(fmt.Sprintf("    Avg Attempts/quest: %.2f | Total Attempts: %d\n", econ.AvgAttempts, econ.TotalAttempts))
		sb.WriteString("    Rank distribution: ")
		for _, r := range []string{"E", "D", "C", "B", "A", "S"} {
			count := econ.RankDistrib[r]
			pct := float64(count) / float64(econ.TotalQuests) * 100
			sb.WriteString(fmt.Sprintf("%s=%d(%.0f%%) ", r, count, pct))
		}
		sb.WriteString("\n\n")
	}

	// ── 2. Monte Carlo Battle Analysis ──────────────────────
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  2. MONTE CARLO BATTLE ANALYSIS (1000 runs each)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Test at "matching level" stats for each enemy
	matchingStats := []struct {
		STR, AGI, INT, STA int
		Label              string
	}{
		{3, 3, 3, 3, "Early (Lv3)"},
		{7, 7, 7, 7, "Mid-early (Lv7)"},
		{12, 12, 12, 12, "Mid (Lv12)"},
		{18, 18, 18, 18, "Mid-late (Lv18)"},
		{25, 25, 25, 25, "Late (Lv25)"},
	}

	for _, stats := range matchingStats {
		sb.WriteString(fmt.Sprintf("  ▸ Player Stats: STR=%d AGI=%d INT=%d STA=%d (%s)\n",
			stats.STR, stats.AGI, stats.INT, stats.STA, stats.Label))
		sb.WriteString(fmt.Sprintf("    %-30s %8s %8s %8s %8s %8s\n",
			"Enemy", "WinRate", "AvgDmg", "StdDev", "AvgRnds", "AvgAcc"))
		sb.WriteString(fmt.Sprintf("    %s\n", strings.Repeat("-", 80)))

		for _, enemy := range enemies {
			mc := MonteCarloAnalysis(
				stats.STR, stats.AGI, stats.INT, stats.STA,
				enemy, 1000, rng,
			)
			sb.WriteString(fmt.Sprintf("    %-30s %7.1f%% %8.0f %8.0f %8.1f %7.1f%%\n",
				mc.EnemyName, mc.WinRate, mc.AvgDamage, mc.StdDevDamage, mc.AvgRounds, mc.AvgAccuracy))
		}
		sb.WriteString("\n")
	}

	// ── 3. Stat Sweep Analysis ───────────────────────────────
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  3. STAT SWEEP ANALYSIS (vary one stat 0-50, others fixed at 10)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Pick a mid-tier enemy for sweep testing
	sweepEnemy := enemies[7] // Демон-маг (B rank, HP 280, ATK 24)
	sb.WriteString(fmt.Sprintf("  Target enemy: %s (Rank %s, HP %d, ATK %d)\n\n",
		sweepEnemy.Name, sweepEnemy.Rank, sweepEnemy.HP, sweepEnemy.Attack))

	for _, stat := range []string{"STR", "AGI", "INT", "STA"} {
		sweep := StatSweep(stat, 50, 10, 10, 10, 10, sweepEnemy, 500, rng)
		sb.WriteString(fmt.Sprintf("  ▸ Sweep: %s (others fixed at 10)\n", stat))
		sb.WriteString(fmt.Sprintf("    %5s %8s %8s %8s\n", "Value", "WinRate", "AvgDmg", "AvgRnds"))
		sb.WriteString(fmt.Sprintf("    %s\n", strings.Repeat("-", 35)))

		// Show every 5th step + endpoints
		for _, r := range sweep {
			if r.StatValue%5 == 0 || r.StatValue == 50 {
				sb.WriteString(fmt.Sprintf("    %5d %7.1f%% %8.0f %8.1f\n",
					r.StatValue, r.WinRate, r.AvgDamage, r.AvgRounds))
			}
		}
		sb.WriteString("\n")
	}

	// ── 4. Progression Runs ──────────────────────────────────
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  4. PROGRESSION RUNS (30/90/180 days, averaged over 10 runs)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, arch := range DefaultArchetypes() {
		sb.WriteString(fmt.Sprintf("  ═══ %s ═══\n\n", arch.Name))

		for _, days := range []int{30, 90, 180} {
			cfg := SimConfig{
				Days:      days,
				Seed:      seed,
				Archetype: arch,
				Enemies:   enemies,
			}

			avgSnapshots := RunProgressionMultiple(cfg, 10)

			sb.WriteString(fmt.Sprintf("  ▸ %d-day run:\n", days))
			sb.WriteString(fmt.Sprintf("    %5s %5s %5s %5s %5s %5s %5s %8s\n",
				"Day", "Lvl", "STR", "AGI", "INT", "STA", "Zone", "WinRate"))
			sb.WriteString(fmt.Sprintf("    %s\n", strings.Repeat("-", 55)))

			// Show key milestones
			milestones := getMilestones(days)
			for _, mi := range milestones {
				if mi < len(avgSnapshots) {
					s := avgSnapshots[mi]
					sb.WriteString(fmt.Sprintf("    %5d %5d %5d %5d %5d %5d %5d %7.1f%%\n",
						s.Day, s.Level, s.STR, s.AGI, s.INT, s.STA, s.Zone, s.TotalWinRate))
				}
			}
			sb.WriteString("\n")

			// Progression timeline check
			checks := CheckProgressionTimeline(avgSnapshots)
			sb.WriteString("    Timeline targets:\n")
			for _, c := range checks {
				status := "✓ MET"
				if !c.Met {
					status = "✗ MISSED"
				}
				if c.ActualDays == -1 {
					status = "✗ NOT REACHED"
				}
				sb.WriteString(fmt.Sprintf("      Zone %d by day %d: actual day %d %s\n",
					c.TargetZone, c.TargetDays, c.ActualDays, status))
			}
			sb.WriteString("\n")
		}
	}

	// ── 5. Balance Summary ───────────────────────────────────
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  5. BALANCE SUMMARY\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	sb.WriteString("  Balance Criteria:\n")
	sb.WriteString("    • Лёгкий враг: 45-60%\n")
	sb.WriteString("    • Нормальный враг: 30-45%\n")
	sb.WriteString("    • Сложный враг: 20-30%\n")
	sb.WriteString("    • Элитка / мини-босс: 12-20%\n")
	sb.WriteString("    • Босс зоны: 5-12%\n")
	sb.WriteString("    • Переход зоны (первые 1-2 врага): 15-25%\n\n")

	for _, zone := range []int{1, 2, 3} {
		sb.WriteString(fmt.Sprintf("  ▸ Zone %d (enemy expected-level interpolation)\n", zone))
		for _, enemy := range enemies {
			if enemy.Zone != zone {
				continue
			}

			evalLevel := EnemyMidExpectedLevel(enemy)
			stats := StatsFromLevel(evalLevel)
			mc := MonteCarloAnalysis(
				stats[0], stats[1], stats[2], stats[3],
				enemy, 1000, rng,
			)

			band := EnemyWinRateBand(enemy)
			targetLow, targetHigh := band.Min, band.Max

			status := "✓ OK"
			if mc.WinRate < targetLow {
				status = "⚠ TOO HARD"
			} else if mc.WinRate > targetHigh {
				status = "⚠ TOO EASY"
			}

			extraTag := ""
			if IsZoneTransitionEnemy(enemy) {
				extraTag = " [TRANSITION]"
			}

			sb.WriteString(fmt.Sprintf("    %-30s L%-2d  %7.1f%% (target %.0f-%.0f%%, %s) %s%s\n",
				enemy.Name, evalLevel, mc.WinRate, targetLow, targetHigh, band.Label, status, extraTag))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// CompactTable generates a concise Day|Level|Zone|Stats|Winrate table.
func CompactTable(cfg SimConfig, runs int) string {
	var sb strings.Builder

	avgSnapshots := RunProgressionMultiple(cfg, runs)

	sb.WriteString(fmt.Sprintf("Archetype: %s | %d days | %d runs averaged\n\n",
		cfg.Archetype.Name, cfg.Days, runs))
	sb.WriteString(fmt.Sprintf("%5s │ %5s │ %5s │ %5s │ %5s │ %5s │ %5s │ %8s\n",
		"Day", "Level", "Zone", "STR", "AGI", "INT", "STA", "WinRate"))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("─", 65)))

	milestones := getMilestones(cfg.Days)
	for _, mi := range milestones {
		if mi < len(avgSnapshots) {
			s := avgSnapshots[mi]
			sb.WriteString(fmt.Sprintf("%5d │ %5d │ %5d │ %5d │ %5d │ %5d │ %5d │ %7.1f%%\n",
				s.Day, s.Level, s.Zone, s.STR, s.AGI, s.INT, s.STA, s.TotalWinRate))
		}
	}

	return sb.String()
}

func getMilestones(days int) []int {
	if days <= 30 {
		milestones := make([]int, 0)
		for d := 0; d < days; d += 5 {
			milestones = append(milestones, d)
		}
		if days-1 > 0 {
			milestones = append(milestones, days-1)
		}
		return milestones
	}

	milestones := []int{0, 6, 13, 20, 29} // week 1,2,3,4
	if days > 30 {
		for d := 59; d < days; d += 30 {
			milestones = append(milestones, d)
		}
	}
	if days-1 > 29 {
		milestones = append(milestones, days-1)
	}
	return milestones
}
