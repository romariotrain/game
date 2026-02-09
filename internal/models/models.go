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
	ID   int64
	Name string
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
	ID             int64
	Name           string
	Description    string
	Rank           QuestRank
	Type           EnemyType
	HP             int
	Attack         int
	PatternSize    int     // number of cells to memorize
	ShowTime       float64 // seconds the pattern is displayed
	RewardEXP      int
	RewardCrystals int
	DropMaterial   MaterialTier // which material tier can drop
	DropChance     float64      // 0.0 - 1.0 probability of material drop
}

// --- Currency & Materials ---

type MaterialTier string

const (
	MaterialCommon MaterialTier = "common"
	MaterialRare   MaterialTier = "rare"
	MaterialEpic   MaterialTier = "epic"
)

func (m MaterialTier) DisplayName() string {
	switch m {
	case MaterialCommon:
		return "Обычный"
	case MaterialRare:
		return "Редкий"
	case MaterialEpic:
		return "Эпический"
	default:
		return string(m)
	}
}

func (m MaterialTier) Color() string {
	switch m {
	case MaterialCommon:
		return "#8a8a8a"
	case MaterialRare:
		return "#4a7fbf"
	case MaterialEpic:
		return "#9b59b6"
	default:
		return "#ffffff"
	}
}

type PlayerResources struct {
	ID             int64
	CharID         int64
	Crystals       int
	MaterialCommon int
	MaterialRare   int
	MaterialEpic   int
}

// --- Equipment System ---

type EquipmentRarity string

const (
	RarityCommon    EquipmentRarity = "common"
	RarityUncommon  EquipmentRarity = "uncommon"
	RarityRare      EquipmentRarity = "rare"
	RarityEpic      EquipmentRarity = "epic"
	RarityLegendary EquipmentRarity = "legendary"
)

var AllRarities = []EquipmentRarity{RarityCommon, RarityUncommon, RarityRare, RarityEpic, RarityLegendary}

func (r EquipmentRarity) DisplayName() string {
	switch r {
	case RarityCommon:
		return "Обычное"
	case RarityUncommon:
		return "Необычное"
	case RarityRare:
		return "Редкое"
	case RarityEpic:
		return "Эпическое"
	case RarityLegendary:
		return "Легендарное"
	default:
		return string(r)
	}
}

func (r EquipmentRarity) Color() string {
	switch r {
	case RarityCommon:
		return "#8a8a8a"
	case RarityUncommon:
		return "#4a9e4a"
	case RarityRare:
		return "#4a7fbf"
	case RarityEpic:
		return "#9b59b6"
	case RarityLegendary:
		return "#e67e22"
	default:
		return "#ffffff"
	}
}

func (r EquipmentRarity) BaseStats() int {
	switch r {
	case RarityCommon:
		return 2
	case RarityUncommon:
		return 5
	case RarityRare:
		return 10
	case RarityEpic:
		return 18
	case RarityLegendary:
		return 30
	default:
		return 0
	}
}

func (r EquipmentRarity) DismantleCrystals() int {
	switch r {
	case RarityCommon:
		return 5
	case RarityUncommon:
		return 15
	case RarityRare:
		return 40
	case RarityEpic:
		return 100
	case RarityLegendary:
		return 250
	default:
		return 0
	}
}

func (r EquipmentRarity) DismantleMaterial() (MaterialTier, int) {
	switch r {
	case RarityCommon:
		return MaterialCommon, 1
	case RarityUncommon:
		return MaterialCommon, 3
	case RarityRare:
		return MaterialRare, 2
	case RarityEpic:
		return MaterialRare, 3
	case RarityLegendary:
		return MaterialEpic, 2
	default:
		return MaterialCommon, 0
	}
}

type EquipmentSlot string

const (
	SlotWeapon    EquipmentSlot = "weapon"
	SlotArmor     EquipmentSlot = "armor"
	SlotAccessory EquipmentSlot = "accessory"
)

func (s EquipmentSlot) DisplayName() string {
	switch s {
	case SlotWeapon:
		return "Оружие"
	case SlotArmor:
		return "Броня"
	case SlotAccessory:
		return "Аксессуар"
	default:
		return string(s)
	}
}

type Equipment struct {
	ID          int64
	CharID      int64
	Name        string
	Slot        EquipmentSlot
	Rarity      EquipmentRarity
	Level       int
	CurrentEXP  int
	BonusAttack int     // weapon: bonus damage in battle
	BonusHP     int     // armor: bonus HP in battle
	BonusTime   float64 // accessory: bonus memorize time (seconds)
	Equipped    bool
	CreatedAt   time.Time
}

func EquipmentEXPForLevel(level int) int {
	return 30 + (level-1)*20
}

// --- Gacha System ---

type GachaBanner string

const (
	BannerNormal   GachaBanner = "normal"
	BannerAdvanced GachaBanner = "advanced"
)

func (b GachaBanner) DisplayName() string {
	switch b {
	case BannerNormal:
		return "Обычный призыв"
	case BannerAdvanced:
		return "Продвинутый призыв"
	default:
		return string(b)
	}
}

func (b GachaBanner) Cost() int {
	switch b {
	case BannerNormal:
		return 100
	case BannerAdvanced:
		return 300
	default:
		return 100
	}
}

type GachaHistory struct {
	ID          int64
	CharID      int64
	Banner      GachaBanner
	EquipmentID int64
	Rarity      EquipmentRarity
	PulledAt    time.Time
}

type GachaPity struct {
	NormalPity   int // pulls since last rare+ on normal banner
	AdvancedPity int // pulls since last epic+ on advanced banner
}

// --- Battle System ---

type BattleResult string

const (
	BattleWin  BattleResult = "win"
	BattleLose BattleResult = "lose"
)

type BattleRecord struct {
	ID             int64
	CharID         int64
	EnemyID        int64
	EnemyName      string
	Result         BattleResult
	DamageDealt    int
	DamageTaken    int
	Accuracy       float64
	CriticalHits   int
	Dodges         int
	RewardEXP      int
	RewardCrystals int
	MaterialDrop   string // empty or material tier
	FoughtAt       time.Time
}

// BattleState holds the live state of a memory-game battle
type BattleState struct {
	Enemy         Enemy
	PlayerHP      int
	PlayerMaxHP   int
	EnemyHP       int
	EnemyMaxHP    int
	Round         int
	Pattern       []int // cell indices to memorize
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

// --- Crafting System ---

type CraftRecipe struct {
	ID           int64
	Name         string
	ResultSlot   EquipmentSlot
	ResultRarity EquipmentRarity
	CostCrystals int
	CostCommon   int
	CostRare     int
	CostEpic     int
}

// --- Daily Rewards ---

type DailyReward struct {
	ID        int64
	CharID    int64
	ClaimedAt time.Time
	Day       int // streak day (1-7, then repeats)
	Crystals  int
}

func DailyRewardCrystals(streakDay int) int {
	day := ((streakDay - 1) % 7) + 1
	switch day {
	case 1:
		return 50
	case 2:
		return 50
	case 3:
		return 75
	case 4:
		return 75
	case 5:
		return 100
	case 6:
		return 100
	case 7:
		return 200
	default:
		return 50
	}
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

type GachaStatistics struct {
	TotalPulls    int
	PullsByBanner map[GachaBanner]int
	PullsByRarity map[EquipmentRarity]int
	CrystalsSpent int
}
