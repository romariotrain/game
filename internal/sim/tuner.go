package sim

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
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
		MinPower:    0.20,
		MaxPower:    2.20,
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
		opts.MinPower = 0.20
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

	enforceZonePowerConstraints(tuned, opts)
	ensureOneBossPerZone(tuned)
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
	baseHP := enemy.HP
	baseATK := enemy.Attack

	minHP := int(math.Round(float64(baseHP) * opts.MinPower))
	maxHP := int(math.Round(float64(baseHP) * opts.MaxPower))
	minATK := int(math.Round(float64(baseATK) * opts.MinPower))
	maxATK := int(math.Round(float64(baseATK) * opts.MaxPower))
	if minHP < 1 {
		minHP = 1
	}
	if maxHP < minHP {
		maxHP = minHP
	}
	if minATK < 1 {
		minATK = 1
	}
	if maxATK < minATK {
		maxATK = minATK
	}

	bestHP := baseHP
	bestATK := baseATK
	bestScore := math.Inf(1)
	bestWR := 0.0
	round := int64(0)

	for atk := minATK; atk <= maxATK; atk++ {
		hpLow := minHP
		hpHigh := maxHP
		localBestHP := hpLow
		localBestWR := 0.0
		localBestScore := math.Inf(1)

		for iter := 0; iter < opts.Iterations; iter++ {
			midHP := (hpLow + hpHigh) / 2
			winRate := evaluateEnemyCandidate(enemy, midHP, atk, stats, opts, round)
			round++

			score := math.Abs(winRate - targetMid)
			if score < localBestScore {
				localBestScore = score
				localBestHP = midHP
				localBestWR = winRate
			}

			// Higher HP makes enemy stronger and winrate lower.
			if winRate > targetMid {
				hpLow = midHP + 1
			} else {
				hpHigh = midHP - 1
			}
			if hpLow > hpHigh {
				break
			}
		}

		// Evaluate search boundaries for this ATK.
		for _, hp := range []int{hpLow, hpHigh} {
			if hp < minHP || hp > maxHP {
				continue
			}
			winRate := evaluateEnemyCandidate(enemy, hp, atk, stats, opts, round)
			round++
			score := math.Abs(winRate - targetMid)
			if score < localBestScore {
				localBestScore = score
				localBestHP = hp
				localBestWR = winRate
			}
		}

		if localBestScore < bestScore {
			bestScore = localBestScore
			bestHP = localBestHP
			bestATK = atk
			bestWR = localBestWR
		}
	}

	tuned := enemy
	tuned.HP = bestHP
	tuned.Attack = bestATK

	return tuned, EnemyTuneResult{
		EnemyName:    enemy.Name,
		TargetLabel:  band.Label,
		TargetMin:    band.Min,
		TargetMax:    band.Max,
		EvalLevel:    evalLevel,
		OldHP:        enemy.HP,
		OldAttack:    enemy.Attack,
		NewHP:        bestHP,
		NewAttack:    bestATK,
		FinalWinRate: bestWR,
	}
}

func evaluateEnemyCandidate(
	enemy EnemyDef,
	hp int,
	atk int,
	stats [4]int,
	opts AutoTuneOptions,
	round int64,
) float64 {
	candidate := enemy
	candidate.HP = hp
	candidate.Attack = atk

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

func enforceZonePowerConstraints(enemies []EnemyDef, opts AutoTuneOptions) {
	zoneToIDs := map[int][]int{}
	for i := range enemies {
		zoneToIDs[enemies[i].Zone] = append(zoneToIDs[enemies[i].Zone], i)
	}

	for _, ids := range zoneToIDs {
		sort.Slice(ids, func(i, j int) bool {
			return enemies[ids[i]].Floor < enemies[ids[j]].Floor
		})

		bossID := -1
		maxOther := 0.0
		for _, id := range ids {
			p := effectivePower(enemies[id])
			if enemies[id].IsBoss {
				bossID = id
				continue
			}
			if p > maxOther {
				maxOther = p
			}
		}
		if bossID >= 0 {
			bossPower := effectivePower(enemies[bossID])
			target := maxOther * 1.05
			if target > 0 && bossPower < target {
				scale := target / bossPower
				if scale > 1.12 {
					scale = 1.12
				}

				candidate := enemies[bossID]
				scaleEnemyPower(&candidate, scale)

				// Priority 1: do not break winrate band while enforcing boss dominance.
				band := EnemyWinRateBand(candidate)
				stats := StatsFromLevel(EnemyMidExpectedLevel(candidate))
				checkRuns := opts.RunsPerEval / 2
				if checkRuns < 120 {
					checkRuns = 120
				}
				check := MonteCarloAnalysis(
					stats[0], stats[1], stats[2], stats[3],
					candidate,
					checkRuns,
					rand.New(rand.NewSource(opts.Seed+int64(candidate.Index)*77+999)),
				)
				if check.WinRate >= band.Min-1.0 {
					enemies[bossID] = candidate
				}
			}
		}
	}
}

func effectivePower(enemy EnemyDef) float64 {
	return float64(enemy.HP)*0.65 + float64(enemy.Attack)*11.0
}

func scaleEnemyPower(enemy *EnemyDef, scale float64) {
	if enemy == nil || scale <= 0 {
		return
	}
	enemy.HP = int(math.Round(float64(enemy.HP) * scale))
	enemy.Attack = int(math.Round(float64(enemy.Attack) * scale))
	if enemy.HP < 1 {
		enemy.HP = 1
	}
	if enemy.Attack < 1 {
		enemy.Attack = 1
	}
}
