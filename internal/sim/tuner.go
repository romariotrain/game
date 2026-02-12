package sim

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// AutoTuneOptions controls binary-search based enemy stat tuning.
type AutoTuneOptions struct {
	Seed        int64
	RunsPerEval int
	Iterations  int
	MinPower    float64
	MaxPower    float64
}

// EnemyTuneResult describes one tuned enemy outcome.
type EnemyTuneResult struct {
	EnemyName    string
	TargetLabel  string
	TargetMin    float64
	TargetMax    float64
	EvalLevel    int
	OldHP        int
	OldAttack    int
	NewHP        int
	NewAttack    int
	FinalWinRate float64
}

// DefaultAutoTuneOptions returns balanced defaults for fast iterative tuning.
func DefaultAutoTuneOptions(seed int64) AutoTuneOptions {
	return AutoTuneOptions{
		Seed:        seed,
		RunsPerEval: 240,
		Iterations:  10,
		MinPower:    0.55,
		MaxPower:    1.85,
	}
}

// AutoTuneEnemies adjusts HP/ATK for each enemy to match its target winrate band.
// It uses binary search over one "power" multiplier and returns tuned copies.
func AutoTuneEnemies(base []EnemyDef, opts AutoTuneOptions) ([]EnemyDef, []EnemyTuneResult) {
	if len(base) == 0 {
		base = GetPresetEnemies()
	}
	if opts.RunsPerEval <= 0 {
		opts.RunsPerEval = 240
	}
	if opts.Iterations <= 0 {
		opts.Iterations = 10
	}
	if opts.MinPower <= 0 {
		opts.MinPower = 0.55
	}
	if opts.MaxPower <= opts.MinPower {
		opts.MaxPower = opts.MinPower + 1.0
	}

	tuned := make([]EnemyDef, len(base))
	copy(tuned, base)

	results := make([]EnemyTuneResult, 0, len(tuned))
	for i := range tuned {
		tunedEnemy, result := tuneOneEnemy(tuned[i], opts)
		tuned[i] = tunedEnemy
		results = append(results, result)
	}

	assignBosses(tuned)
	return tuned, results
}

func tuneOneEnemy(enemy EnemyDef, opts AutoTuneOptions) (EnemyDef, EnemyTuneResult) {
	band := EnemyWinRateBand(enemy)
	evalLevel := EnemyMidExpectedLevel(enemy)
	if IsZoneTransitionEnemy(enemy) {
		// For zone-opening enemies, optimize at entry level first.
		evalLevel = TransitionEntryLevel(enemy)
	}
	stats := StatsFromLevel(evalLevel)
	targetMid := (band.Min + band.Max) / 2.0

	low := opts.MinPower
	high := opts.MaxPower
	bestPower := 1.0
	bestScore := math.Inf(1)
	bestWR := 0.0

	for iter := 0; iter < opts.Iterations; iter++ {
		mid := (low + high) / 2.0
		winRate := evaluateEnemyPower(enemy, mid, stats, opts, int64(iter))
		score := math.Abs(winRate - targetMid)
		if score < bestScore {
			bestScore = score
			bestPower = mid
			bestWR = winRate
		}

		// Higher power makes enemy stronger and winrate lower.
		if winRate > targetMid {
			low = mid
		} else {
			high = mid
		}
	}

	// Final check on interval edges in case MC-noise shifted best point.
	for _, power := range []float64{low, high} {
		winRate := evaluateEnemyPower(enemy, power, stats, opts, int64(opts.Iterations)+1)
		score := math.Abs(winRate - targetMid)
		if score < bestScore {
			bestScore = score
			bestPower = power
			bestWR = winRate
		}
	}

	tuned := enemy
	tuned.HP = int(math.Round(float64(enemy.HP) * bestPower))
	if tuned.HP < 1 {
		tuned.HP = 1
	}
	tuned.Attack = int(math.Round(float64(enemy.Attack) * bestPower))
	if tuned.Attack < 1 {
		tuned.Attack = 1
	}

	return tuned, EnemyTuneResult{
		EnemyName:    enemy.Name,
		TargetLabel:  band.Label,
		TargetMin:    band.Min,
		TargetMax:    band.Max,
		EvalLevel:    evalLevel,
		OldHP:        enemy.HP,
		OldAttack:    enemy.Attack,
		NewHP:        tuned.HP,
		NewAttack:    tuned.Attack,
		FinalWinRate: bestWR,
	}
}

func evaluateEnemyPower(
	enemy EnemyDef,
	power float64,
	stats [4]int,
	opts AutoTuneOptions,
	round int64,
) float64 {
	candidate := enemy
	candidate.HP = int(math.Round(float64(enemy.HP) * power))
	if candidate.HP < 1 {
		candidate.HP = 1
	}
	candidate.Attack = int(math.Round(float64(enemy.Attack) * power))
	if candidate.Attack < 1 {
		candidate.Attack = 1
	}

	seed := opts.Seed + int64(enemy.Index)*1000 + round*17
	mc := MonteCarloAnalysis(
		stats[0], stats[1], stats[2], stats[3],
		candidate,
		opts.RunsPerEval,
		rand.New(rand.NewSource(seed)),
	)
	return mc.WinRate
}

// AutoTuneSummary renders a concise table for tuned enemies.
func AutoTuneSummary(results []EnemyTuneResult) string {
	var sb strings.Builder
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  AUTO-TUNE SUMMARY\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString(fmt.Sprintf("  %-30s %4s %20s %14s %9s\n",
		"Enemy", "Lvl", "HP/ATK old→new", "Target", "FinalWR"))
	sb.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("-", 86)))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("  %-30s %4d %9d/%-4d→%5d/%-4d %6.0f-%.0f%% %-8s %7.1f%%\n",
			r.EnemyName,
			r.EvalLevel,
			r.OldHP, r.OldAttack,
			r.NewHP, r.NewAttack,
			r.TargetMin, r.TargetMax, "("+r.TargetLabel+")",
			r.FinalWinRate,
		))
	}
	sb.WriteString("\n")
	return sb.String()
}
