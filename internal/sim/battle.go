package sim

import (
	"math"
	"math/rand"
)

// BattleOutcome represents a single simulated battle result.
type BattleOutcome struct {
	Win         bool
	Rounds      int
	DamageDealt int
	DamageTaken int
	Crits       int
	Accuracy    float64
}

// SimulateBattle runs one complete battle between player stats and an enemy.
func SimulateBattle(
	str, agi, intStat, sta int,
	enemy EnemyDef,
	rng *rand.Rand,
) BattleOutcome {
	playerLevel := PlayerCombatLevel(str, sta)
	enemyScale := EnemyScaleFactor(enemy, playerLevel)

	playerHP := PlayerHP(sta)
	enemyHP := int(math.Round(float64(enemy.HP) * enemyScale))
	if enemyHP < 1 {
		enemyHP = 1
	}
	enemyAttack := int(math.Round(float64(enemy.Attack) * enemyScale))
	if enemyAttack < 1 {
		enemyAttack = 1
	}

	totalDamageDealt := 0
	totalDamageTaken := 0
	totalCrits := 0
	totalHits := 0
	totalShown := 0
	round := 0

	rngFloat := func() float64 { return rng.Float64() }

	for playerHP > 0 && enemyHP > 0 {
		round++
		accuracy := SimulatedAccuracy(intStat, rngFloat)
		totalShown += 100
		totalHits += int(math.Round(accuracy * 100.0))

		damage, isCrit := ComputePlayerDamage(str, agi, accuracy, rngFloat)
		if isCrit {
			totalCrits++
		}
		enemyHP -= damage
		totalDamageDealt += damage

		if enemyHP <= 0 {
			break
		}

		eDmg := ComputeEnemyDamage(enemyAttack, sta, rngFloat)
		playerHP -= eDmg
		totalDamageTaken += eDmg
	}

	overallAcc := 0.0
	if totalShown > 0 {
		overallAcc = float64(totalHits) / float64(totalShown)
	}

	return BattleOutcome{
		Win:         enemyHP <= 0,
		Rounds:      round,
		DamageDealt: totalDamageDealt,
		DamageTaken: totalDamageTaken,
		Crits:       totalCrits,
		Accuracy:    overallAcc,
	}
}

// MonteCarloAnalysis runs N battles for a given player+enemy config.
func MonteCarloAnalysis(
	str, agi, intStat, sta int,
	enemy EnemyDef,
	runs int,
	rng *rand.Rand,
) BattleMCResult {
	wins := 0
	var totalDmg float64
	var totalRounds float64
	var totalAcc float64
	damages := make([]float64, runs)

	for i := 0; i < runs; i++ {
		outcome := SimulateBattle(str, agi, intStat, sta, enemy, rng)
		if outcome.Win {
			wins++
		}
		dmg := float64(outcome.DamageDealt)
		damages[i] = dmg
		totalDmg += dmg
		totalRounds += float64(outcome.Rounds)
		totalAcc += outcome.Accuracy
	}

	avgDmg := totalDmg / float64(runs)

	// Compute std dev of damage
	var varianceSum float64
	for _, d := range damages {
		diff := d - avgDmg
		varianceSum += diff * diff
	}
	stdDevDmg := math.Sqrt(varianceSum / float64(runs))

	return BattleMCResult{
		EnemyName:    enemy.Name,
		Runs:         runs,
		Wins:         wins,
		Losses:       runs - wins,
		WinRate:      float64(wins) / float64(runs) * 100,
		AvgDamage:    avgDmg,
		StdDevDamage: stdDevDmg,
		AvgRounds:    totalRounds / float64(runs),
		AvgAccuracy:  totalAcc / float64(runs) * 100,
	}
}

// StatSweep varies one stat from 0 to maxVal, keeping others fixed.
// Returns winrate and avg damage for each stat value.
func StatSweep(
	statName string,
	maxVal int,
	fixedSTR, fixedAGI, fixedINT, fixedSTA int,
	enemy EnemyDef,
	runsPerPoint int,
	rng *rand.Rand,
) []StatSweepResult {
	results := make([]StatSweepResult, 0, maxVal+1)
	baselineLevel := PlayerCombatLevel(fixedSTR, fixedSTA)
	baselineScale := EnemyScaleFactor(enemy, baselineLevel)
	sweepEnemy := enemy
	// Sweep is meant to isolate stat impact, so keep enemy baseline fixed
	// at the reference player level.
	sweepEnemy.HP = int(math.Round(float64(sweepEnemy.HP) * baselineScale))
	if sweepEnemy.HP < 1 {
		sweepEnemy.HP = 1
	}
	sweepEnemy.Attack = int(math.Round(float64(sweepEnemy.Attack) * baselineScale))
	if sweepEnemy.Attack < 1 {
		sweepEnemy.Attack = 1
	}
	sweepEnemy.ExpectedMinLevel = 0
	sweepEnemy.ExpectedMaxLevel = 0

	for v := 0; v <= maxVal; v++ {
		str, agi, intS, sta := fixedSTR, fixedAGI, fixedINT, fixedSTA
		switch statName {
		case "STR":
			str = v
		case "AGI":
			agi = v
		case "INT":
			intS = v
		case "STA":
			sta = v
		}

		mc := MonteCarloAnalysis(str, agi, intS, sta, sweepEnemy, runsPerPoint, rng)

		results = append(results, StatSweepResult{
			StatName:  statName,
			StatValue: v,
			WinRate:   mc.WinRate,
			AvgDamage: mc.AvgDamage,
			AvgRounds: mc.AvgRounds,
		})
	}

	return results
}
