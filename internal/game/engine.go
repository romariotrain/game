package game

import (
	"fmt"
	"math"
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
	case floor <= 10:
		return "Туманные Болота"
	case floor <= 20:
		return "Забытые Руины"
	case floor <= 30:
		return "Ледяные Пики"
	case floor <= 40:
		return "Пепельные Разломы"
	case floor <= 50:
		return "Цитадель Бездны"
	default:
		return fmt.Sprintf("Башня: Этаж %d", floor)
	}
}

func GetPresetEnemies() []models.Enemy {
	type zoneTemplate struct {
		Zone      int
		Biome     string
		MinLevel  int
		MaxLevel  int
		Names     []string
		LoreLabel string
	}

	zones := []zoneTemplate{
		{
			Zone:      1,
			Biome:     "swamp",
			MinLevel:  2,
			MaxLevel:  6,
			LoreLabel: "Туманные Болота",
			Names: []string{
				"Квакающий Разведчик",
				"Болотный Пиявочник",
				"Гнилотный Шаман",
				"Трясинный Волк",
				"Моховой Голем",
				"Токсичный Удильщик",
				"Ведьмин Грибник",
				"Слизень-Разъедатель",
				"Пасть Топи",
				"Хозяйка Туманов Морра",
			},
		},
		{
			Zone:      2,
			Biome:     "ruins",
			MinLevel:  7,
			MaxLevel:  12,
			LoreLabel: "Забытые Руины",
			Names: []string{
				"Ржавый Страж Портала",
				"Пыльный Скелет-Рыцарь",
				"Летучая Моль Проклятий",
				"Крипт-Охотник",
				"Костяной Арбалетчик",
				"Каменный Идол",
				"Тень Архива",
				"Пожиратель Реликвий",
				"Жрец Разломанных Печатей",
				"Архонт Руин Кальдрос",
			},
		},
		{
			Zone:      3,
			Biome:     "frost",
			MinLevel:  13,
			MaxLevel:  18,
			LoreLabel: "Ледяные Пики",
			Names: []string{
				"Снежный Падальщик",
				"Морозный Пехотинец",
				"Ледяная Гарпия",
				"Вьюжный Волк",
				"Осколочный Голем",
				"Северный Берсерк",
				"Хрустальный Охотник",
				"Ледяной Колдун",
				"Белый Йети",
				"Король Вьюги Хельгрим",
			},
		},
		{
			Zone:      4,
			Biome:     "volcanic",
			MinLevel:  19,
			MaxLevel:  24,
			LoreLabel: "Пепельные Разломы",
			Names: []string{
				"Пепельный Разбойник",
				"Обугленный Скелет",
				"Лавовый Плевун",
				"Огненный Гончий",
				"Шлаковый Голем",
				"Жрец Пепла",
				"Крылатый Угольник",
				"Демон Искр",
				"Плавильщик Костей",
				"Владыка Разломов Азгар",
			},
		},
		{
			Zone:      5,
			Biome:     "void",
			MinLevel:  25,
			MaxLevel:  30,
			LoreLabel: "Цитадель Бездны",
			Names: []string{
				"Безликий Смотритель",
				"Паразит Пустоты",
				"Теневой Дуэлянт",
				"Пожиратель Света",
				"Хор Бездны",
				"Клеймённый Инквизитор",
				"Рыцарь Нулевой Тени",
				"Коготь Монарха",
				"Оракул Тишины",
				"Монарх Бездны Ноктэрн",
			},
		},
	}

	enemies := make([]models.Enemy, 0, 50)
	floor := 1

	for _, z := range zones {
		for i, name := range z.Names {
			slot := i + 1
			level := levelForSlot(z.MinLevel, z.MaxLevel, slot)
			role := roleForSlot(slot)
			targetMin, targetMax := targetWinrateForSlot(slot)
			isBoss := slot == 10

			typeValue := models.EnemyRegular
			if isBoss {
				typeValue = models.EnemyBoss
			}

			baseHP := 120 + level*26
			baseATK := 8 + int(math.Round(float64(level)*0.9))
			power := rolePowerMultiplier(slot)
			hp := int(math.Round(float64(baseHP) * power))
			atk := int(math.Round(float64(baseATK) * (0.8 + power*0.35)))

			enemies = append(enemies, models.Enemy{
				Name:             name,
				Description:      fmt.Sprintf("%s: %s", z.LoreLabel, roleLore(role)),
				Rank:             rankForZoneAndSlot(z.Zone, slot),
				Type:             typeValue,
				Level:            level,
				HP:               hp,
				Attack:           atk,
				Floor:            floor,
				Zone:             z.Zone,
				IsBoss:           isBoss,
				Biome:            z.Biome,
				Role:             role,
				IsTransition:     slot <= 2,
				TargetWinRateMin: targetMin,
				TargetWinRateMax: targetMax,
			})
			floor++
		}
	}

	return enemies
}

func (e *Engine) InitEnemies() error {
	preset := GetPresetEnemies()
	needsReseed, err := e.DB.EnemyCatalogNeedsReseed(preset)
	if err != nil {
		return err
	}
	if needsReseed {
		if err := e.DB.ReplaceEnemyCatalog(preset); err != nil {
			return err
		}
	}
	return e.DB.NormalizeEnemyZones()
}

func levelForSlot(minLevel, maxLevel, slot int) int {
	pattern := []int{0, 1, 1, 2, 2, 3, 4, 4, 5, 5}
	if slot < 1 {
		slot = 1
	}
	if slot > len(pattern) {
		slot = len(pattern)
	}
	level := minLevel + pattern[slot-1]
	if level > maxLevel {
		level = maxLevel
	}
	return level
}

func roleForSlot(slot int) string {
	switch slot {
	case 1:
		return "TRANSITION"
	case 2:
		return "TRANSITION_ELITE"
	case 3:
		return "NORMAL"
	case 4:
		return "HARD"
	case 5:
		return "EASY"
	case 6:
		return "HARD"
	case 7:
		return "ELITE"
	case 8:
		return "NORMAL"
	case 9:
		return "MINIBOSS"
	case 10:
		return "BOSS"
	default:
		return "NORMAL"
	}
}

func rolePowerMultiplier(slot int) float64 {
	switch slot {
	case 1:
		return 1.15
	case 2:
		return 1.22
	case 3:
		return 1.00
	case 4:
		return 1.08
	case 5:
		return 0.90
	case 6:
		return 1.10
	case 7:
		return 1.20
	case 8:
		return 1.02
	case 9:
		return 1.30
	case 10:
		return 1.38
	default:
		return 1.00
	}
}

func targetWinrateForSlot(slot int) (float64, float64) {
	switch slot {
	case 1:
		return 15, 25
	case 2:
		return 12, 20
	case 3:
		return 30, 45
	case 4:
		return 20, 30
	case 5:
		return 45, 60
	case 6:
		return 20, 30
	case 7:
		return 12, 20
	case 8:
		return 30, 45
	case 9:
		return 8, 15
	case 10:
		return 5, 12
	default:
		return 30, 45
	}
}

func rankForZoneAndSlot(zone, slot int) models.QuestRank {
	switch zone {
	case 1:
		if slot <= 5 {
			return models.RankE
		}
		return models.RankD
	case 2:
		if slot <= 5 {
			return models.RankC
		}
		return models.RankB
	case 3:
		if slot <= 4 {
			return models.RankB
		}
		if slot <= 8 {
			return models.RankA
		}
		return models.RankS
	case 4:
		if slot <= 5 {
			return models.RankA
		}
		return models.RankS
	default:
		return models.RankS
	}
}

func roleLore(role string) string {
	switch role {
	case "TRANSITION":
		return "входной страж зоны: опасен с первых секунд."
	case "TRANSITION_ELITE":
		return "пограничный элитный противник, проверяет базу билда."
	case "HARD":
		return "усиливает давление и наказывает ошибки."
	case "EASY":
		return "тактическая передышка, но не бесплатная."
	case "ELITE":
		return "элитный враг с усиленной выживаемостью."
	case "MINIBOSS":
		return "мини-босс, близок к порогу зоны."
	case "BOSS":
		return "властитель зоны и ключ к следующему этапу."
	default:
		return "боевой противник башни."
	}
}

func (e *Engine) GetEnemies() ([]models.Enemy, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return nil, err
	}
	currentZone, err := e.GetCurrentZone(e.Character.ID)
	if err != nil {
		return nil, err
	}

	var available []models.Enemy
	for _, en := range allEnemies {
		if en.Zone == currentZone {
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

			nextEnemy, err := e.GetNextEnemyForPlayer()
			if err != nil {
				return nil, err
			}
			if nextEnemy != nil && nextEnemy.ID != state.Enemy.ID {
				record.UnlockedEnemyName = nextEnemy.Name
			}
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
