package boss

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"solo-leveling/internal/game/combat/memory"
	"solo-leveling/internal/models"
)

type Phase string

const (
	PhaseMemory Phase = "memory"
	PhaseWin    Phase = "win"
	PhaseLose   Phase = "lose"
)

type MemoryRound struct {
	GridSize    int
	CellsToShow int
	ShowTimeMs  int
	ShownCells  []int
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

	DamageDealt int
	DamageTaken int
	TotalHits   int
	TotalMisses int
	TotalCrits  int

	// Per-round info for UI
	LastRoundDamage   int
	LastRoundEnemyDmg int
	LastRoundHits     int
	LastRoundTotal    int
	LastRoundAccuracy float64
	LastRoundCrit     bool
	RoundLog          []string
}

func NewState(enemy models.Enemy, stats memory.Stats, playerHP int) (*State, error) {
	gridSize := memory.GridSize(enemy)
	cellsToShow := memory.CellsToShow(enemy, stats)
	showTimeMs := memory.TimeToShow(stats)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	shown, err := memory.GenerateShownCells(gridSize, cellsToShow, rng)
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
			GridSize:    gridSize,
			CellsToShow: cellsToShow,
			ShowTimeMs:  showTimeMs,
			ShownCells:  shown,
		},
	}
	return st, nil
}

func ApplyMemoryInput(state *State, choices []int, stats memory.Stats, bonusAttack int) error {
	if state.Phase != PhaseMemory {
		return errors.New("not in memory phase")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	accuracy := memory.ComputeAccuracy(state.Memory.ShownCells, choices)
	hits := memory.CorrectClicks(state.Memory.ShownCells, choices)
	total := state.Memory.CellsToShow
	misses := total - hits
	if misses < 0 {
		misses = 0
	}

	damage, isCrit := memory.ComputePlayerDamage(stats, accuracy, rng)
	damage += bonusAttack
	enemyDamage := memory.ComputeEnemyDamage(state.Enemy, stats, accuracy, rng)

	state.EnemyHP -= damage
	state.PlayerHP -= enemyDamage
	state.DamageDealt += damage
	state.DamageTaken += enemyDamage
	state.TotalHits += hits
	state.TotalMisses += misses
	if isCrit {
		state.TotalCrits++
	}

	state.LastRoundDamage = damage
	state.LastRoundEnemyDmg = enemyDamage
	state.LastRoundHits = hits
	state.LastRoundTotal = total
	state.LastRoundAccuracy = accuracy
	state.LastRoundCrit = isCrit

	logLine := fmt.Sprintf("Раунд %d: %.0f%% (%d/%d) → %d урона", state.Round, accuracy*100, hits, total, damage)
	state.RoundLog = append(state.RoundLog, logLine)
	if isCrit {
		state.RoundLog = append(state.RoundLog, "⚡ Крит! x1.5")
	}
	if enemyDamage > 0 {
		state.RoundLog = append(state.RoundLog, fmt.Sprintf("Враг атакует: -%d HP", enemyDamage))
	}
	if len(state.RoundLog) > 6 {
		state.RoundLog = state.RoundLog[len(state.RoundLog)-6:]
	}

	if state.EnemyHP <= 0 {
		state.EnemyHP = 0
		state.Phase = PhaseWin
		return nil
	}
	if state.PlayerHP <= 0 {
		state.PlayerHP = 0
		state.Phase = PhaseLose
		return nil
	}

	state.Round++
	state.Memory.CellsToShow = memory.CellsToShow(state.Enemy, stats)
	state.Memory.ShowTimeMs = memory.TimeToShow(stats)
	shown, err := memory.GenerateShownCells(state.Memory.GridSize, state.Memory.CellsToShow, rng)
	if err != nil {
		return err
	}
	state.Memory.ShownCells = shown

	return nil
}

func CalcAccuracy(totalHits, totalMisses int) float64 {
	total := totalHits + totalMisses
	if total == 0 {
		return 0
	}
	return math.Round((float64(totalHits)/float64(total))*1000) / 10
}
