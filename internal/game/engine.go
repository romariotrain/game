package game

import (
	"fmt"
	"math/rand"
	"time"

	"solo-leveling/internal/database"
	"solo-leveling/internal/game/combat/memory"
	"solo-leveling/internal/models"
)

type Engine struct {
	DB                    *database.DB
	Character             *models.Character
	RecommendationSource  string
	RecommendationDetails string
}

func NewEngine(db *database.DB) (*Engine, error) {
	char, err := db.GetOrCreateCharacter("Hunter")
	if err != nil {
		return nil, fmt.Errorf("init character: %w", err)
	}
	return &Engine{
		DB:                    db,
		Character:             char,
		RecommendationSource:  "rule-based",
		RecommendationDetails: "инициализация",
	}, nil
}

// ============================================================
// Enemies
// ============================================================

// FloorName returns the display name for a tower floor range.
func FloorName(floor int) string {
	switch {
	case floor <= 5:
		return "Башня: Новичок"
	case floor <= 10:
		return "Башня: Искатель"
	case floor <= 15:
		return "Башня: Ветеран"
	default:
		return fmt.Sprintf("Башня: Этаж %d", floor)
	}
}

func GetPresetEnemies() []models.Enemy {
	return []models.Enemy{
		// === Floor 1-5: Novice ===
		{Name: "Гоблин", Description: "Слабый, но хитрый монстр", Rank: models.RankE, Type: models.EnemyRegular,
			HP: 80, Attack: 8, Floor: 1},
		{Name: "Скелет-страж", Description: "Неупокоенный воин подземелья", Rank: models.RankE, Type: models.EnemyRegular,
			HP: 90, Attack: 9, Floor: 2},
		{Name: "Волк-тень", Description: "Быстрый хищник из тёмного мира", Rank: models.RankD, Type: models.EnemyRegular,
			HP: 120, Attack: 12, Floor: 3},
		{Name: "Каменный голем", Description: "Медленный, но невероятно прочный", Rank: models.RankD, Type: models.EnemyRegular,
			HP: 150, Attack: 14, Floor: 4},
		{Name: "Игрис — Рыцарь Крови", Description: "Легендарный рыцарь, верный страж данжа. Его клинок не знает пощады.",
			Rank: models.RankC, Type: models.EnemyBoss,
			HP: 300, Attack: 20, Floor: 5},

		// === Floor 6-10: Seeker ===
		{Name: "Тёмный рыцарь", Description: "Опытный воин, павший во тьму", Rank: models.RankC, Type: models.EnemyRegular,
			HP: 200, Attack: 18, Floor: 6},
		{Name: "Ядовитый виверн", Description: "Крылатый ящер с отравленным жалом", Rank: models.RankC, Type: models.EnemyRegular,
			HP: 220, Attack: 20, Floor: 7},
		{Name: "Демон-маг", Description: "Владеет разрушительной магией", Rank: models.RankB, Type: models.EnemyRegular,
			HP: 280, Attack: 24, Floor: 8},
		{Name: "Ледяной великан", Description: "Древний страж ледяных пещер", Rank: models.RankB, Type: models.EnemyRegular,
			HP: 320, Attack: 26, Floor: 9},
		{Name: "Барука — Король муравьёв", Description: "Монструозный повелитель насекомых. Его панцирь почти непробиваем.",
			Rank: models.RankA, Type: models.EnemyBoss,
			HP: 500, Attack: 32, Floor: 10},

		// === Floor 11-15: Veteran ===
		{Name: "Архидемон", Description: "Один из высших демонов, невероятная мощь", Rank: models.RankA, Type: models.EnemyRegular,
			HP: 450, Attack: 30, Floor: 11},
		{Name: "Драконид", Description: "Полудракон-получеловек, воин огня", Rank: models.RankA, Type: models.EnemyRegular,
			HP: 500, Attack: 34, Floor: 12},
		{Name: "Страж Бездны", Description: "Сущность из глубин тёмного измерения", Rank: models.RankS, Type: models.EnemyRegular,
			HP: 600, Attack: 38, Floor: 13},
		{Name: "Монарх Хаоса", Description: "Повелитель разрушения и безумия", Rank: models.RankS, Type: models.EnemyRegular,
			HP: 700, Attack: 42, Floor: 14},
		{Name: "Монарх Теней", Description: "Повелитель теней и смерти. Последнее испытание для сильнейших.",
			Rank: models.RankS, Type: models.EnemyBoss,
			HP: 1000, Attack: 50, Floor: 15},
	}
}

func (e *Engine) InitEnemies() error {
	count, err := e.DB.GetEnemyCount()
	if err != nil {
		return err
	}
	if count == 0 {
		enemies := GetPresetEnemies()
		for i := range enemies {
			if err := e.DB.InsertEnemy(&enemies[i]); err != nil {
				return err
			}
		}
	}

	// Ensure the first enemy is unlocked for the player
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return err
	}
	if len(allEnemies) > 0 {
		_ = e.DB.UnlockEnemy(e.Character.ID, allEnemies[0].ID)
	}
	return nil
}

func (e *Engine) GetEnemies() ([]models.Enemy, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return nil, err
	}
	unlocked, err := e.DB.GetUnlockedEnemyIDs(e.Character.ID)
	if err != nil {
		return nil, err
	}
	if len(unlocked) == 0 && len(allEnemies) > 0 {
		_ = e.DB.UnlockEnemy(e.Character.ID, allEnemies[0].ID)
		unlocked[allEnemies[0].ID] = true
	}

	var available []models.Enemy
	for _, en := range allEnemies {
		if unlocked[en.ID] {
			available = append(available, en)
		}
	}
	return available, nil
}

// ============================================================
// Battle System
// ============================================================

func (e *Engine) StartBattle(enemyID int64) (*models.BattleState, error) {
	// Spend 1 attempt
	if err := e.DB.SpendAttempt(e.Character.ID); err != nil {
		return nil, fmt.Errorf("нет попыток для боя! Выполняйте задания, чтобы получить попытки")
	}
	e.Character.Attempts--

	enemy, err := e.DB.GetEnemyByID(enemyID)
	if err != nil {
		return nil, err
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}

	statMap := make(map[models.StatType]int)
	for _, s := range stats {
		statMap[s.StatType] = s.Level
	}

	memStats := memory.Stats{
		STR: statMap[models.StatStrength],
		AGI: statMap[models.StatAgility],
		INT: statMap[models.StatIntellect],
		STA: statMap[models.StatEndurance],
	}
	memDiff := memory.DifficultyFor(enemy.Rank, memStats)

	// Base player HP from endurance
	endurance := statMap[models.StatEndurance]
	if endurance == 0 {
		endurance = 1
	}
	baseHP := 100 + endurance*10

	playerHP := baseHP

	// Generate pattern (Tactical Memory)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	pattern, err := memory.GeneratePattern(memDiff.GridSize, memDiff.PatternLength, rng)
	if err != nil {
		return nil, err
	}

	return &models.BattleState{
		Enemy:         *enemy,
		PlayerHP:      playerHP,
		PlayerMaxHP:   playerHP,
		EnemyHP:       enemy.HP,
		EnemyMaxHP:    enemy.HP,
		Round:         1,
		GridSize:      memDiff.GridSize,
		PatternLength: memDiff.PatternLength,
		ShowTimeMs:    memDiff.ShowTimeMs,
		AllowedErrors: memDiff.AllowedErrors,
		Pattern:       pattern,
	}, nil
}

// ProcessRound evaluates the player's guesses for one round of battle
func (e *Engine) ProcessRound(state *models.BattleState, guesses []int) error {
	if state.BattleOver {
		return fmt.Errorf("battle is already over")
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return err
	}

	statMap := make(map[models.StatType]int)
	for _, s := range stats {
		statMap[s.StatType] = s.Level
	}

	// Tactical Memory accuracy: ordered sequence comparison
	total := len(state.Pattern)
	hits := 0
	for i := 0; i < total; i++ {
		if i < len(guesses) && guesses[i] == state.Pattern[i] {
			hits++
		}
	}
	errors := total - hits
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(hits) / float64(total)
	}

	strength := statMap[models.StatStrength]
	// STR reduces required correct steps
	requiredHits := total - (strength / 5)
	if requiredHits < 1 {
		requiredHits = 1
	}

	baseDamage := 8 + strength*2

	damage := 0
	if hits >= requiredHits {
		damage = int(float64(baseDamage) * accuracy)
	}

	// Enemy attacks
	enemyDamage := state.Enemy.Attack

	// STA reduces incoming damage slightly
	endurance := statMap[models.StatEndurance]
	enemyDamage -= endurance / 5
	if enemyDamage < 0 {
		enemyDamage = 0
	}

	// Penalty for exceeding allowed errors
	if errors > state.AllowedErrors {
		enemyDamage += (errors - state.AllowedErrors) * 5
	}

	crits := 0
	dodges := 0

	// Apply damage
	state.EnemyHP -= damage
	state.PlayerHP -= enemyDamage

	state.TotalHits += hits
	state.TotalMisses += errors
	state.TotalCrits += crits
	state.TotalDodges += dodges
	state.DamageDealt += damage
	state.DamageTaken += enemyDamage

	// Check battle outcome
	if state.EnemyHP <= 0 {
		state.EnemyHP = 0
		state.BattleOver = true
		state.Result = models.BattleWin
	} else if state.PlayerHP <= 0 {
		state.PlayerHP = 0
		state.BattleOver = true
		state.Result = models.BattleLose
	} else {
		// New round, new pattern
		state.Round++
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		pattern, err := memory.GeneratePattern(state.GridSize, state.PatternLength, rng)
		if err != nil {
			return err
		}
		state.Pattern = pattern
	}

	state.PlayerGuesses = guesses
	return nil
}

// FinishBattle records the battle result and awards rewards
func (e *Engine) FinishBattle(state *models.BattleState) (*models.BattleRecord, error) {
	totalRounds := state.Round
	accuracy := 0.0
	totalPatternCells := state.TotalHits + state.TotalMisses
	if totalPatternCells > 0 {
		accuracy = float64(state.TotalHits) / float64(totalPatternCells) * 100
	}

	record := &models.BattleRecord{
		CharID:       e.Character.ID,
		EnemyID:      state.Enemy.ID,
		EnemyName:    state.Enemy.Name,
		Result:       state.Result,
		DamageDealt:  state.DamageDealt,
		DamageTaken:  state.DamageTaken,
		Accuracy:     accuracy,
		CriticalHits: state.TotalCrits,
		Dodges:       state.TotalDodges,
	}

	if state.Result == models.BattleWin {
		// First-win rewards: title, badge, unlock next enemy
		existing, err := e.DB.GetBattleReward(e.Character.ID, state.Enemy.ID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			title := fmt.Sprintf("Покоритель: %s", state.Enemy.Name)
			badge := fmt.Sprintf("Знак: %s", state.Enemy.Name)
			reward := &models.BattleReward{
				CharID:  e.Character.ID,
				EnemyID: state.Enemy.ID,
				Title:   title,
				Badge:   badge,
			}
			if err := e.DB.InsertBattleReward(reward); err != nil {
				return nil, err
			}
			record.RewardTitle = title
			record.RewardBadge = badge

			// Unlock next enemy in sequence
			allEnemies, err := e.DB.GetAllEnemies()
			if err != nil {
				return nil, err
			}
			for i := range allEnemies {
				if allEnemies[i].ID == state.Enemy.ID && i+1 < len(allEnemies) {
					next := allEnemies[i+1]
					_ = e.DB.UnlockEnemy(e.Character.ID, next.ID)
					record.UnlockedEnemyName = next.Name
					break
				}
			}
		}
	}

	_ = totalRounds
	if err := e.DB.InsertBattle(record); err != nil {
		return nil, err
	}

	return record, nil
}

// GetShowTimeMs returns adjusted show time (ms) based on accessory bonuses
func (e *Engine) GetShowTimeMs(baseTimeMs int) (int, error) {
	return baseTimeMs, nil
}

func (e *Engine) GetBattleStats() (*models.BattleStatistics, error) {
	return e.DB.GetBattleStats(e.Character.ID)
}

func (e *Engine) GetBattleHistory(limit int) ([]models.BattleRecord, error) {
	return e.DB.GetBattleHistory(e.Character.ID, limit)
}
