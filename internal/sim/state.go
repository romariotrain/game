package sim

import "math/rand"

// PlayerState holds a simulated player's full state.
type PlayerState struct {
	// Stat levels (STR, AGI, INT, STA)
	STR int
	AGI int
	INT int
	STA int

	// EXP per stat (accumulated toward next level)
	STREXP int
	AGIEXP int
	INTEXP int
	STAEXP int

	// Battle attempts available
	Attempts int

	// Zone progression
	CurrentZone       int
	DefeatedEnemyIDs  map[int]bool // index into preset enemies
	CurrentEnemyIndex int          // index into zone-filtered enemy list

	// Tracking
	TotalQuestsCompleted int
	TotalQuestsFailed    int
	TotalEXPEarned       int
	TotalBattles         int
	TotalBattleWins      int
	TotalBattleLosses    int
	CurrentStreak        int // consecutive days with ≥1 quest completed
	DayNumber            int
}

// NewPlayerState creates a fresh player at day 0.
func NewPlayerState() *PlayerState {
	return &PlayerState{
		STR:              1,
		AGI:              1,
		INT:              1,
		STA:              1,
		Attempts:         0,
		CurrentZone:      1,
		DefeatedEnemyIDs: make(map[int]bool),
	}
}

// OverallLevel = average of all 4 stat levels (integer division).
func (p *PlayerState) OverallLevel() int {
	return (p.STR + p.AGI + p.INT + p.STA) / 4
}

// StatLevel returns the level for a given stat name.
func (p *PlayerState) StatLevel(stat string) int {
	switch stat {
	case "strength":
		return p.STR
	case "agility":
		return p.AGI
	case "intellect":
		return p.INT
	case "endurance":
		return p.STA
	default:
		return 0
	}
}

// EnemyDef is a simplified enemy definition for simulation (no DB dependency).
type EnemyDef struct {
	Index  int
	Name   string
	Rank   string // "E","D","C","B","A","S"
	HP     int
	Attack int
	AGI    int
	INT    int
	// ExpectedMinLevel/ExpectedMaxLevel define the intended player-level window.
	ExpectedMinLevel int
	ExpectedMaxLevel int
	Floor            int
	Zone             int
	IsBoss           bool
}

// Archetype describes a player behavior pattern for simulation.
type Archetype struct {
	Name string

	// Quest generation parameters
	QuestsPerDay   int     // how many quests completed per day
	AvgMinutes     int     // average minutes per quest
	AvgEffort      int     // average effort [1-5]
	AvgFriction    int     // average friction [1-3]
	MinutesStdDev  float64 // std deviation for minutes
	EffortStdDev   float64 // std deviation for effort
	FrictionStdDev float64 // std deviation for friction

	// Stat distribution preference (weights, must sum to 1.0)
	STRWeight float64
	AGIWeight float64
	INTWeight float64
	STAWeight float64

	// Whether the player fights whenever attempts are available
	FightsWhenPossible bool
}

// SimConfig controls simulation parameters.
type SimConfig struct {
	Days           int
	Seed           int64
	Archetype      Archetype
	Enemies        []EnemyDef
	Verbose        bool
	MonteCarloRuns int // for battle MC analysis; 0 = skip
}

// SimRNG wraps *rand.Rand to satisfy the memory.RNG interface.
type SimRNG struct {
	R *rand.Rand
}

func (s *SimRNG) Float64() float64 { return s.R.Float64() }
func (s *SimRNG) Perm(n int) []int { return s.R.Perm(n) }

// DaySnapshot records the state at the end of a simulated day.
type DaySnapshot struct {
	Day          int
	Level        int
	STR          int
	AGI          int
	INT          int
	STA          int
	Zone         int
	QuestsToday  int
	EXPToday     int
	BattlesToday int
	WinsToday    int
	TotalWinRate float64
	Attempts     int
}

// BattleMCResult holds Monte Carlo analysis for a single battle configuration.
type BattleMCResult struct {
	EnemyName    string
	Runs         int
	Wins         int
	Losses       int
	WinRate      float64
	AvgDamage    float64
	StdDevDamage float64
	AvgRounds    float64
	AvgAccuracy  float64
}

// StatSweepResult holds one data point from a stat sweep.
type StatSweepResult struct {
	StatName  string
	StatValue int
	WinRate   float64
	AvgDamage float64
	AvgRounds float64
}

// EXPEconomyResult holds analysis of EXP distribution.
type EXPEconomyResult struct {
	TotalQuests   int
	AvgEXP        float64
	MinEXP        int
	MaxEXP        int
	RankDistrib   map[string]int // rank -> count
	AvgAttempts   float64
	TotalAttempts int
}

// ProgressionCheck verifies timeline targets.
type ProgressionCheck struct {
	TargetZone int
	TargetDays int
	ActualDays int // -1 if not reached
	Met        bool
}
