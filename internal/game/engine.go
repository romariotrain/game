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
	e := &Engine{
		DB:                    db,
		Character:             char,
		RecommendationSource:  "rule-based",
		RecommendationDetails: "инициализация",
	}
	if err := e.InitAchievements(); err != nil {
		return nil, fmt.Errorf("init achievements: %w", err)
	}
	return e, nil
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
	enemy, err := e.validateCurrentEnemyForFight(enemyID)
	if err != nil {
		return nil, err
	}
	if enemy.Type == models.EnemyBoss {
		return nil, fmt.Errorf("этот противник требует бой с боссом")
	}
	if err := e.spendBattleAttempt(); err != nil {
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
	playerHP := memory.PlayerHP(memStats.STA)
	gridSize := memory.GridSize(*enemy)
	cellsToShow := memory.CellsToShow(*enemy, memStats)
	showTimeMs := memory.TimeToShow(memStats)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	shown, err := memory.GenerateShownCells(gridSize, cellsToShow, rng)
	if err != nil {
		return nil, err
	}

	return &models.BattleState{
		Enemy:       *enemy,
		PlayerHP:    playerHP,
		PlayerMaxHP: playerHP,
		EnemyHP:     enemy.HP,
		EnemyMaxHP:  enemy.HP,
		Round:       1,
		GridSize:    gridSize,
		CellsToShow: cellsToShow,
		ShowTimeMs:  showTimeMs,
		ShownCells:  shown,
	}, nil
}

// ProcessRound evaluates one visual-memory round.
func (e *Engine) ProcessRound(state *models.BattleState, choices []int) error {
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

	str := statMap[models.StatStrength]
	statsForRound := memory.Stats{
		STR: str,
		AGI: statMap[models.StatAgility],
		INT: statMap[models.StatIntellect],
		STA: statMap[models.StatEndurance],
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	accuracy := memory.ComputeAccuracy(state.ShownCells, choices)
	hits := memory.CorrectClicks(state.ShownCells, choices)
	total := state.CellsToShow
	misses := total - hits
	if misses < 0 {
		misses = 0
	}

	damage, isCrit := memory.ComputePlayerDamage(statsForRound, accuracy, rng)
	enemyDamage := memory.ComputeEnemyDamage(state.Enemy, statsForRound, accuracy, rng)
	if isCrit {
		state.TotalCrits++
	}

	// Apply damage
	state.EnemyHP -= damage
	state.PlayerHP -= enemyDamage

	state.TotalHits += hits
	state.TotalMisses += misses
	state.DamageDealt += damage
	state.DamageTaken += enemyDamage

	// Per-round info for UI
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
		state.Round++
		state.CellsToShow = memory.CellsToShow(state.Enemy, statsForRound)
		state.ShowTimeMs = memory.TimeToShow(statsForRound)
		shown, err := memory.GenerateShownCells(state.GridSize, state.CellsToShow, rng)
		if err != nil {
			return err
		}
		state.ShownCells = shown
	}

	state.PlayerChoices = choices
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

			nextName, err := e.unlockNextEnemy(state.Enemy.ID)
			if err != nil {
				return nil, err
			}
			record.UnlockedEnemyName = nextName
		}
		if err := e.UnlockAchievement(AchievementFirstBattle); err != nil {
			return nil, err
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
