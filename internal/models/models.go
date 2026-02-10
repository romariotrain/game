package models

import "time"

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
	ID       int64
	Name     string
	Attempts int
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

// AttemptsForRank returns how many battle attempts a quest of the given rank awards.
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
	ID          int64
	CharID      int64
	Title       string
	Description string
	Rank        QuestRank
	TargetStat  StatType
	Status      QuestStatus
	CreatedAt   time.Time
	CompletedAt *time.Time
	IsDaily     bool
	TemplateID  *int64 // link to daily_quest_templates
	DungeonID   *int64 // link to dungeon_quests if this is a dungeon quest
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
	ID          int64
	CharID      int64
	Title       string
	Description string
	Rank        QuestRank
	TargetStat  StatType
	Active      bool // whether this template is still active (user can disable)
	CreatedAt   time.Time
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

// --- Dungeons ---

type DungeonStatus string

const (
	DungeonLocked     DungeonStatus = "locked"
	DungeonAvailable  DungeonStatus = "available"
	DungeonInProgress DungeonStatus = "in_progress"
	DungeonCompleted  DungeonStatus = "completed"
)

type StatRequirement struct {
	StatType StatType `json:"stat_type"`
	MinLevel int      `json:"min_level"`
}

type Dungeon struct {
	ID               int64
	Name             string
	Description      string
	Requirements     []StatRequirement
	Status           DungeonStatus
	RewardTitle      string
	RewardEXP        int               // bonus EXP to ALL stats on completion
	QuestDefinitions []DungeonQuestDef // quest templates inside the dungeon
}

type DungeonQuestDef struct {
	ID          int64
	DungeonID   int64
	Title       string
	Description string
	Rank        QuestRank
	TargetStat  StatType
}

type CompletedDungeon struct {
	ID          int64
	CharID      int64
	DungeonID   int64
	CompletedAt time.Time
	EarnedTitle string
}

// --- Enemy System ---

type EnemyType string

const (
	EnemyRegular EnemyType = "regular"
	EnemyBoss    EnemyType = "boss"
)

type Enemy struct {
	ID          int64
	Name        string
	Description string
	Rank        QuestRank
	Type        EnemyType
	HP          int
	Attack      int
	Floor       int
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
	PatternLength int
	ShowTimeMs    int
	AllowedErrors int
	Pattern       []int // ordered sequence of cell indices
	PlayerGuesses []int // player's chosen cells
	TotalHits     int
	TotalMisses   int
	TotalCrits    int
	TotalDodges   int
	DamageDealt   int
	DamageTaken   int
	BattleOver    bool
	Result        BattleResult
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
