package models

import (
	"math"
	"time"
)

type StatType string

const (
	StatStrength  StatType = "strength"
	StatAgility   StatType = "agility"
	StatIntellect StatType = "intellect"
	StatEndurance StatType = "endurance"
)

var AllStats = []StatType{StatStrength, StatAgility, StatIntellect, StatEndurance}

func (s StatType) DisplayName() string {
	switch s {
	case StatStrength:
		return "Сила"
	case StatAgility:
		return "Ловкость"
	case StatIntellect:
		return "Интеллект"
	case StatEndurance:
		return "Выносливость"
	default:
		return string(s)
	}
}

func (s StatType) Icon() string {
	switch s {
	case StatStrength:
		return "⚔️"
	case StatAgility:
		return "🏃"
	case StatIntellect:
		return "📖"
	case StatEndurance:
		return "🛡️"
	default:
		return "?"
	}
}

type QuestRank string

const (
	RankE QuestRank = "E"
	RankD QuestRank = "D"
	RankC QuestRank = "C"
	RankB QuestRank = "B"
	RankA QuestRank = "A"
	RankS QuestRank = "S"
)

var AllRanks = []QuestRank{RankE, RankD, RankC, RankB, RankA, RankS}

// RankFromEXP maps quest EXP to display rank.
func RankFromEXP(exp int) QuestRank {
	switch {
	case exp <= 10:
		return RankE
	case exp <= 18:
		return RankD
	case exp <= 28:
		return RankC
	case exp <= 40:
		return RankB
	case exp <= 55:
		return RankA
	default:
		return RankS
	}
}

// CalculateQuestEXP calculates deterministic quest EXP from workload signals.
func CalculateQuestEXP(minutes, effort, friction int) int {
	if minutes < 0 {
		minutes = 0
	}
	if effort < 1 {
		effort = 1
	}
	if effort > 5 {
		effort = 5
	}
	if friction < 1 {
		friction = 1
	}
	if friction > 3 {
		friction = 3
	}
	base := float64(minutes) * 0.6
	exp := int(math.Round(base + float64(effort*4+friction*3)))
	if exp < 1 {
		return 1
	}
	return exp
}

func (r QuestRank) BaseEXP() int {
	switch r {
	case RankE:
		return 20
	case RankD:
		return 40
	case RankC:
		return 70
	case RankB:
		return 120
	case RankA:
		return 200
	case RankS:
		return 350
	default:
		return 0
	}
}

func (r QuestRank) Color() string {
	switch r {
	case RankE:
		return "#8a8a8a"
	case RankD:
		return "#4a9e4a"
	case RankC:
		return "#4a7fbf"
	case RankB:
		return "#9b59b6"
	case RankA:
		return "#e67e22"
	case RankS:
		return "#e74c3c"
	default:
		return "#ffffff"
	}
}

type QuestStatus string

const (
	QuestActive    QuestStatus = "active"
	QuestCompleted QuestStatus = "completed"
	QuestFailed    QuestStatus = "failed"
)

type Character struct {
	ID          int64
	Name        string
	Attempts    int
	ActiveTitle string
}

type Achievement struct {
	ID          int64
	Key         string
	Title       string
	Description string
	Category    string
	ObtainedAt  *time.Time
	IsUnlocked  bool
}

type HunterProfile struct {
	CharID                   int64
	About                    string
	Goals                    string
	Priorities               string
	TimeBudget               string
	PhysicalConstraints      string
	PsychologicalConstraints string
	DayRoutine               string
	PrimaryPlaces            string
	Dislikes                 string
	SupportStyle             string
	ExtraDetails             string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type QuestSuggestion struct {
	Title            string
	Description      string
	Rank             QuestRank
	TargetStat       StatType
	EstimatedMinutes int
}

type PlayerStats struct {
	STR int
	AGI int
	INT int
	STA int
}

type AISuggestion struct {
	Title    string   `json:"title"`
	Desc     string   `json:"desc"`
	Minutes  int      `json:"minutes"`
	Effort   int      `json:"effort"`
	Friction int      `json:"friction"`
	Stat     string   `json:"stat"`
	Tags     []string `json:"tags"`
}

const MaxAttempts = 8

// AttemptsForRank is legacy mapping retained for compatibility in non-quest systems.
func AttemptsForRank(rank QuestRank) int {
	switch rank {
	case RankE, RankD:
		return 1
	case RankC, RankB:
		return 2
	case RankA, RankS:
		return 3
	default:
		return 1
	}
}

// AttemptsForQuestEXP returns attempts awarded by quest EXP.
func AttemptsForQuestEXP(exp int) int {
	switch {
	case exp < 15:
		return 1
	case exp <= 30:
		return 2
	default:
		return 3
	}
}

type StatLevel struct {
	ID         int64
	CharID     int64
	StatType   StatType
	Level      int
	CurrentEXP int
	TotalEXP   int
}

func ExpForLevel(level int) int {
	return 50 + (level-1)*30
}

type Quest struct {
	ID               int64
	CharID           int64
	Title            string
	Description      string
	Congratulations  string
	Exp              int
	Rank             QuestRank
	TargetStat       StatType
	Status           QuestStatus
	CreatedAt        time.Time
	CompletedAt      *time.Time
	IsDaily          bool
	TemplateID       *int64 // link to daily_quest_templates
	ExpeditionID     *int64 // link to expeditions if this is an expedition quest
	ExpeditionTaskID *int64 // link to expedition_tasks for task progress updates
}

type Skill struct {
	ID          int64
	CharID      int64
	Name        string
	Description string
	StatType    StatType
	Multiplier  float64
	UnlockedAt  int
	Active      bool
}

type QuestHistoryEntry struct {
	Quest       Quest
	EXPAwarded  int
	CompletedAt time.Time
}

// --- Daily Quest Templates ---

type DailyQuestTemplate struct {
	ID              int64
	CharID          int64
	Title           string
	Description     string
	Congratulations string
	Exp             int
	Rank            QuestRank
	TargetStat      StatType
	Active          bool // whether this template is still active (user can disable)
	CreatedAt       time.Time
}

// --- Daily Activity ---

type DailyActivity struct {
	ID             int64
	CharID         int64
	Date           string // YYYY-MM-DD
	QuestsComplete int
	QuestsFailed   int
	EXPEarned      int
}

// --- Statistics ---

type Statistics struct {
	TotalQuestsCompleted int
	TotalQuestsFailed    int
	QuestsByRank         map[QuestRank]int
	TotalEXPEarned       int
	BestStat             StatType
	BestStatLevel        int
	CurrentStreak        int
	SuccessRate          float64
	StatLevels           []StatLevel
}

// --- Expeditions ---

type ExpeditionStatus string

const (
	ExpeditionActive    ExpeditionStatus = "active"
	ExpeditionCompleted ExpeditionStatus = "completed"
	ExpeditionFailed    ExpeditionStatus = "failed"
)

type Expedition struct {
	ID           int64
	Name         string
	Description  string
	Deadline     *time.Time
	RewardEXP    int
	RewardStats  map[StatType]int
	IsRepeatable bool
	Status       ExpeditionStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Tasks        []ExpeditionTask
}

type ExpeditionTask struct {
	ID              int64
	ExpeditionID    int64
	Title           string
	Description     string
	IsCompleted     bool
	ProgressCurrent int
	ProgressTarget  int
	RewardEXP       int
	TargetStat      StatType
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CompletedExpedition struct {
	ID           int64
	CharID       int64
	ExpeditionID int64
	CompletedAt  time.Time
}

// --- Enemy System ---

type EnemyType string

const (
	EnemyRegular EnemyType = "regular"
	EnemyBoss    EnemyType = "boss"
)

type Enemy struct {
	ID               int64
	Name             string
	Description      string
	Rank             QuestRank
	Type             EnemyType
	Level            int
	HP               int
	Attack           int
	Floor            int
	Zone             int
	IsBoss           bool
	Biome            string
	Role             string
	IsTransition     bool
	TargetWinRateMin float64
	TargetWinRateMax float64
}

// StreakTitle returns the title earned at a given streak milestone, or empty string.
func StreakTitle(streak int) string {
	switch {
	case streak >= 365:
		return "Неостановимый (365 дней)"
	case streak >= 100:
		return "Легенда Дисциплины (100 дней)"
	case streak >= 30:
		return "Стальная Воля (30 дней)"
	case streak >= 7:
		return "Начинающий Охотник (7 дней)"
	default:
		return ""
	}
}

// AllStreakMilestones returns all streak milestones in order.
func AllStreakMilestones() []struct {
	Days  int
	Title string
} {
	return []struct {
		Days  int
		Title string
	}{
		{7, "Начинающий Охотник"},
		{30, "Стальная Воля"},
		{100, "Легенда Дисциплины"},
		{365, "Неостановимый"},
	}
}

// --- Battle System ---

type BattleResult string

const (
	BattleWin  BattleResult = "win"
	BattleLose BattleResult = "lose"
)

type BattleRecord struct {
	ID           int64
	CharID       int64
	EnemyID      int64
	EnemyName    string
	Result       BattleResult
	DamageDealt  int
	DamageTaken  int
	Accuracy     float64
	CriticalHits int
	Dodges       int
	FoughtAt     time.Time
	// Runtime-only reward info (not persisted)
	RewardTitle       string
	RewardBadge       string
	UnlockedEnemyName string
}

// BattleState holds the live state of a memory-game battle
type BattleState struct {
	Enemy         Enemy
	PlayerHP      int
	PlayerMaxHP   int
	EnemyHP       int
	EnemyMaxHP    int
	Round         int
	GridSize      int
	CellsToShow   int
	ShowTimeMs    int
	ShownCells    []int // highlighted cells for current round
	PlayerChoices []int // player's chosen cells
	TotalHits     int
	TotalMisses   int
	TotalCrits    int
	TotalDodges   int
	DamageDealt   int
	DamageTaken   int
	BattleOver    bool
	Result        BattleResult

	// Per-round info for battle log / UI feedback
	LastRoundDamage   int
	LastRoundEnemyDmg int
	LastRoundHits     int
	LastRoundTotal    int
	LastRoundAccuracy float64
	LastRoundCrit     bool
	RoundLog          []string // last N log lines
}

// --- Extended Statistics ---

type BattleStatistics struct {
	TotalBattles    int
	Wins            int
	Losses          int
	WinRate         float64
	TotalDamage     int
	TotalCrits      int
	TotalDodges     int
	EnemiesDefeated map[string]int // enemy name -> defeat count
}

// --- Battle Rewards ---

type BattleReward struct {
	ID        int64
	CharID    int64
	EnemyID   int64
	Title     string
	Badge     string
	AwardedAt time.Time
}

type DefeatedEnemy struct {
	EnemyID     int64
	Name        string
	Description string
	Rank        QuestRank
	Zone        int
	IsBoss      bool
	DefeatedAt  *time.Time
}
