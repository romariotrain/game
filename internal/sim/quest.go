package sim

import (
	"math"
	"math/rand"
)

// SimulateQuest generates a quest with random parameters based on archetype,
// calculates EXP, awards it to a stat, and handles level-ups.
// Returns the quest EXP and stat awarded to.
func SimulateQuest(player *PlayerState, arch Archetype, rng *rand.Rand) (int, string) {
	// Generate quest parameters with normal distribution
	minutes := int(math.Round(float64(arch.AvgMinutes) + rng.NormFloat64()*arch.MinutesStdDev))
	if minutes < 5 {
		minutes = 5
	}
	if minutes > 120 {
		minutes = 120
	}

	effort := int(math.Round(float64(arch.AvgEffort) + rng.NormFloat64()*arch.EffortStdDev))
	if effort < 1 {
		effort = 1
	}
	if effort > 5 {
		effort = 5
	}

	friction := int(math.Round(float64(arch.AvgFriction) + rng.NormFloat64()*arch.FrictionStdDev))
	if friction < 1 {
		friction = 1
	}
	if friction > 3 {
		friction = 3
	}

	questEXP := CalculateQuestEXP(minutes, effort, friction)
	rank := RankFromEXP(questEXP)
	statEXP := BaseEXPForRank(rank)

	// Pick target stat based on archetype weights
	stat := pickStat(arch, rng)

	// Award stat EXP and handle level-ups
	addStatEXP(player, stat, statEXP)

	// Award battle attempts
	attempts := AttemptsForQuestEXP(questEXP)
	player.Attempts += attempts
	if player.Attempts > MaxAttempts {
		player.Attempts = MaxAttempts
	}

	// Tracking
	player.TotalQuestsCompleted++
	player.TotalEXPEarned += statEXP

	return questEXP, stat
}

// pickStat selects a stat based on archetype weight distribution.
func pickStat(arch Archetype, rng *rand.Rand) string {
	roll := rng.Float64()
	cumulative := 0.0

	cumulative += arch.STRWeight
	if roll < cumulative {
		return "strength"
	}
	cumulative += arch.AGIWeight
	if roll < cumulative {
		return "agility"
	}
	cumulative += arch.INTWeight
	if roll < cumulative {
		return "intellect"
	}
	return "endurance"
}

// addStatEXP adds EXP to a stat and processes level-ups.
func addStatEXP(player *PlayerState, stat string, exp int) {
	var level *int
	var currentEXP *int

	switch stat {
	case "strength":
		level = &player.STR
		currentEXP = &player.STREXP
	case "agility":
		level = &player.AGI
		currentEXP = &player.AGIEXP
	case "intellect":
		level = &player.INT
		currentEXP = &player.INTEXP
	case "endurance":
		level = &player.STA
		currentEXP = &player.STAEXP
	default:
		return
	}

	*currentEXP += exp
	for {
		required := ExpForLevel(*level)
		if *currentEXP >= required {
			*currentEXP -= required
			*level++
		} else {
			break
		}
	}
}

// EXPEconomyAnalysis generates N random quests and analyzes EXP distribution.
func EXPEconomyAnalysis(arch Archetype, n int, rng *rand.Rand) EXPEconomyResult {
	rankDist := make(map[string]int)
	totalEXP := 0
	minEXP := math.MaxInt64
	maxEXP := 0
	totalAttempts := 0

	for i := 0; i < n; i++ {
		minutes := int(math.Round(float64(arch.AvgMinutes) + rng.NormFloat64()*arch.MinutesStdDev))
		if minutes < 5 {
			minutes = 5
		}
		if minutes > 120 {
			minutes = 120
		}

		effort := int(math.Round(float64(arch.AvgEffort) + rng.NormFloat64()*arch.EffortStdDev))
		if effort < 1 {
			effort = 1
		}
		if effort > 5 {
			effort = 5
		}

		friction := int(math.Round(float64(arch.AvgFriction) + rng.NormFloat64()*arch.FrictionStdDev))
		if friction < 1 {
			friction = 1
		}
		if friction > 3 {
			friction = 3
		}

		exp := CalculateQuestEXP(minutes, effort, friction)
		rank := RankFromEXP(exp)
		rankDist[rank]++

		totalEXP += exp
		if exp < minEXP {
			minEXP = exp
		}
		if exp > maxEXP {
			maxEXP = exp
		}

		attempts := AttemptsForQuestEXP(exp)
		totalAttempts += attempts
	}

	return EXPEconomyResult{
		TotalQuests:   n,
		AvgEXP:        float64(totalEXP) / float64(n),
		MinEXP:        minEXP,
		MaxEXP:        maxEXP,
		RankDistrib:   rankDist,
		AvgAttempts:   float64(totalAttempts) / float64(n),
		TotalAttempts: totalAttempts,
	}
}
