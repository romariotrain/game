package boss

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"solo-leveling/internal/game/combat/memory"
	"solo-leveling/internal/models"
)

type Phase string

const (
	PhaseMemory   Phase = "memory"
	PhasePressure Phase = "pressure"
	PhaseWin      Phase = "win"
	PhaseLose     Phase = "lose"
)

type MemoryRound struct {
	GridSize      int
	PatternLength int
	ShowTimeMs    int
	AllowedErrors int
	Pattern       []int
}

type PressurePuzzle struct {
	Steps         int
	TimeLimitMs   int
	AllowedErrors int
	AttemptsLeft  int
	NextExpected  int
	Errors        int
	StartedAt     time.Time
}

type State struct {
	Enemy       models.Enemy
	Phase       Phase
	Round       int
	PlayerHP    int
	PlayerMaxHP int
	EnemyHP     int
	EnemyMaxHP  int
	Memory      MemoryRound
	Puzzle      PressurePuzzle

	DamageDealt int
	DamageTaken int
	TotalHits   int
	TotalMisses int
}

func NewState(enemy models.Enemy, stats memory.Stats, playerHP int) (*State, error) {
	memDiff := memory.DifficultyFor(enemy.Rank, stats)
	memDiff = tougher(memDiff)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	pattern, err := memory.GeneratePattern(memDiff.GridSize, memDiff.PatternLength, rng)
	if err != nil {
		return nil, err
	}

	st := &State{
		Enemy:       enemy,
		Phase:       PhaseMemory,
		Round:       1,
		PlayerHP:    playerHP,
		PlayerMaxHP: playerHP,
		EnemyHP:     enemy.HP,
		EnemyMaxHP:  enemy.HP,
		Memory: MemoryRound{
			GridSize:      memDiff.GridSize,
			PatternLength: memDiff.PatternLength,
			ShowTimeMs:    memDiff.ShowTimeMs,
			AllowedErrors: memDiff.AllowedErrors,
			Pattern:       pattern,
		},
	}
	return st, nil
}

func tougher(d memory.Difficulty) memory.Difficulty {
	// Boss phase 1 is tougher than regular Tactical Memory
	return memory.Difficulty{
		GridSize:      clamp(d.GridSize+1, 3, 6),
		PatternLength: clamp(d.PatternLength+1, 4, 9),
		ShowTimeMs:    clamp(d.ShowTimeMs-200, 700, 3500),
		AllowedErrors: clamp(d.AllowedErrors-1, 0, 3),
	}
}

func ApplyMemoryInput(state *State, guesses []int, stats memory.Stats, bonusAttack int) error {
	if state.Phase != PhaseMemory {
		return errors.New("not in memory phase")
	}

	total := len(state.Memory.Pattern)
	hits := 0
	for i := 0; i < total; i++ {
		if i < len(guesses) && guesses[i] == state.Memory.Pattern[i] {
			hits++
		}
	}
	errors := total - hits

	accuracy := 0.0
	if total > 0 {
		accuracy = float64(hits) / float64(total)
	}

	requiredHits := total - (stats.STR / 5)
	if requiredHits < 1 {
		requiredHits = 1
	}

	baseDamage := 10 + stats.STR*2 + bonusAttack
	damage := 0
	if hits >= requiredHits {
		damage = int(float64(baseDamage) * accuracy)
	}

	enemyDamage := state.Enemy.Attack
	enemyDamage -= stats.STA / 5
	if enemyDamage < 0 {
		enemyDamage = 0
	}
	if errors > state.Memory.AllowedErrors {
		enemyDamage += (errors - state.Memory.AllowedErrors) * 6
	}

	state.EnemyHP -= damage
	state.PlayerHP -= enemyDamage
	state.DamageDealt += damage
	state.DamageTaken += enemyDamage
	state.TotalHits += hits
	state.TotalMisses += errors

	if state.EnemyHP <= 0 {
		state.EnemyHP = 0
		state.Phase = PhasePressure
		state.Puzzle = NewPressurePuzzle(state.Enemy.Rank, stats)
		return nil
	}
	if state.PlayerHP <= 0 {
		state.PlayerHP = 0
		state.Phase = PhaseLose
		return nil
	}

	state.Round++
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	pattern, err := memory.GeneratePattern(state.Memory.GridSize, state.Memory.PatternLength, rng)
	if err != nil {
		return err
	}
	state.Memory.Pattern = pattern

	return nil
}

func NewPressurePuzzle(rank models.QuestRank, stats memory.Stats) PressurePuzzle {
	steps := baseSteps(rank)
	steps -= stats.STR / 5
	if steps < 3 {
		steps = 3
	}

	timeLimit := baseTimeMs(rank) + stats.INT*120
	if timeLimit < 1200 {
		timeLimit = 1200
	}
	if timeLimit > 8000 {
		timeLimit = 8000
	}

	allowedErrors := 0
	if stats.AGI >= 10 {
		allowedErrors = 1
	}

	attempts := 1 + stats.STA/10
	if attempts > 5 {
		attempts = 5
	}

	return PressurePuzzle{
		Steps:         steps,
		TimeLimitMs:   timeLimit,
		AllowedErrors: allowedErrors,
		AttemptsLeft:  attempts,
		NextExpected:  1,
		StartedAt:     time.Now(),
	}
}

func ApplyPuzzleInput(state *State, value int) error {
	if state.Phase != PhasePressure {
		return errors.New("not in pressure phase")
	}
	if state.Puzzle.AttemptsLeft <= 0 {
		state.Phase = PhaseLose
		return nil
	}

	if value == state.Puzzle.NextExpected {
		state.Puzzle.NextExpected++
		if state.Puzzle.NextExpected > state.Puzzle.Steps {
			state.Phase = PhaseWin
		}
		return nil
	}

	state.Puzzle.Errors++
	if state.Puzzle.Errors > state.Puzzle.AllowedErrors {
		state.Puzzle.AttemptsLeft--
		state.Puzzle.Errors = 0
		state.Puzzle.NextExpected = 1
		if state.Puzzle.AttemptsLeft <= 0 {
			state.Phase = PhaseLose
		}
	}
	return nil
}

func PressureTimedOut(state *State) {
	if state.Phase == PhasePressure {
		state.Phase = PhaseLose
	}
}

func baseSteps(rank models.QuestRank) int {
	switch rank {
	case models.RankE:
		return 4
	case models.RankD:
		return 5
	case models.RankC:
		return 6
	case models.RankB:
		return 7
	case models.RankA:
		return 8
	case models.RankS:
		return 9
	default:
		return 6
	}
}

func baseTimeMs(rank models.QuestRank) int {
	switch rank {
	case models.RankE:
		return 4000
	case models.RankD:
		return 3800
	case models.RankC:
		return 3400
	case models.RankB:
		return 3000
	case models.RankA:
		return 2600
	case models.RankS:
		return 2300
	default:
		return 3200
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func CalcAccuracy(totalHits, totalMisses int) float64 {
	total := totalHits + totalMisses
	if total == 0 {
		return 0
	}
	return math.Round((float64(totalHits)/float64(total))*1000) / 10
}
