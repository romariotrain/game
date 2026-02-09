package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"solo-leveling/internal/database"
	"solo-leveling/internal/models"
)

type Engine struct {
	DB        *database.DB
	Character *models.Character
}

func NewEngine(db *database.DB) (*Engine, error) {
	char, err := db.GetOrCreateCharacter("Hunter")
	if err != nil {
		return nil, fmt.Errorf("init character: %w", err)
	}
	return &Engine{DB: db, Character: char}, nil
}

func (e *Engine) GetStatLevels() ([]models.StatLevel, error) {
	return e.DB.GetStatLevels(e.Character.ID)
}

func (e *Engine) GetOverallLevel() (int, error) {
	stats, err := e.GetStatLevels()
	if err != nil {
		return 0, err
	}
	total := 0
	for _, s := range stats {
		total += s.Level
	}
	return total / len(stats), nil
}

func (e *Engine) GetEXPMultiplier(stat models.StatType) (float64, error) {
	skills, err := e.DB.GetSkills(e.Character.ID)
	if err != nil {
		return 1.0, err
	}
	mult := 1.0
	for _, s := range skills {
		if s.Active && s.StatType == stat {
			mult *= s.Multiplier
		}
	}
	return mult, nil
}

type CompleteResult struct {
	EXPAwarded int
	LeveledUp  bool
	OldLevel   int
	NewLevel   int
	StatType   models.StatType
}

func (e *Engine) CompleteQuest(questID int64) (*CompleteResult, error) {
	active, err := e.DB.GetActiveQuests(e.Character.ID)
	if err != nil {
		return nil, err
	}

	var quest *models.Quest
	for i := range active {
		if active[i].ID == questID {
			quest = &active[i]
			break
		}
	}
	if quest == nil {
		return nil, fmt.Errorf("quest not found or not active")
	}

	baseEXP := quest.Rank.BaseEXP()
	mult, err := e.GetEXPMultiplier(quest.TargetStat)
	if err != nil {
		return nil, err
	}
	expAwarded := int(math.Round(float64(baseEXP) * mult))

	stats, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}

	var stat *models.StatLevel
	for i := range stats {
		if stats[i].StatType == quest.TargetStat {
			stat = &stats[i]
			break
		}
	}
	if stat == nil {
		return nil, fmt.Errorf("stat not found: %s", quest.TargetStat)
	}

	oldLevel := stat.Level
	stat.CurrentEXP += expAwarded
	stat.TotalEXP += expAwarded

	for {
		required := models.ExpForLevel(stat.Level)
		if stat.CurrentEXP >= required {
			stat.CurrentEXP -= required
			stat.Level++
		} else {
			break
		}
	}

	if err := e.DB.UpdateStatLevel(stat); err != nil {
		return nil, err
	}
	if err := e.DB.CompleteQuest(questID); err != nil {
		return nil, err
	}

	// Record daily activity
	e.DB.RecordDailyActivity(e.Character.ID, 1, 0, expAwarded)

	return &CompleteResult{
		EXPAwarded: expAwarded,
		LeveledUp:  stat.Level > oldLevel,
		OldLevel:   oldLevel,
		NewLevel:   stat.Level,
		StatType:   quest.TargetStat,
	}, nil
}

func (e *Engine) CreateQuest(title, description string, rank models.QuestRank, targetStat models.StatType, isDaily bool) (*models.Quest, error) {
	q := &models.Quest{
		CharID:      e.Character.ID,
		Title:       title,
		Description: description,
		Rank:        rank,
		TargetStat:  targetStat,
		IsDaily:     isDaily,
	}

	// If daily, create a template first
	if isDaily {
		tmpl := &models.DailyQuestTemplate{
			CharID:      e.Character.ID,
			Title:       title,
			Description: description,
			Rank:        rank,
			TargetStat:  targetStat,
		}
		if err := e.DB.CreateDailyTemplate(tmpl); err != nil {
			return nil, err
		}
		q.TemplateID = &tmpl.ID
	}

	if err := e.DB.CreateQuest(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (e *Engine) FailQuest(questID int64) error {
	// Record daily activity for failure
	e.DB.RecordDailyActivity(e.Character.ID, 0, 1, 0)
	return e.DB.FailQuest(questID)
}

func (e *Engine) DeleteQuest(questID int64) error {
	// If it's a daily quest with a template, disable the template too
	q, err := e.DB.GetQuestByID(questID)
	if err == nil && q.IsDaily && q.TemplateID != nil {
		e.DB.DisableDailyTemplate(*q.TemplateID)
	}
	return e.DB.DeleteQuest(questID)
}

// SpawnDailyQuests checks all active daily templates and creates today's quests if missing
func (e *Engine) SpawnDailyQuests() (int, error) {
	templates, err := e.DB.GetActiveDailyTemplates(e.Character.ID)
	if err != nil {
		return 0, err
	}

	spawned := 0
	for _, tmpl := range templates {
		exists, err := e.DB.HasDailyQuestForToday(e.Character.ID, tmpl.ID)
		if err != nil {
			return spawned, err
		}
		if exists {
			continue
		}

		templateID := tmpl.ID
		q := &models.Quest{
			CharID:      e.Character.ID,
			Title:       tmpl.Title,
			Description: tmpl.Description,
			Rank:        tmpl.Rank,
			TargetStat:  tmpl.TargetStat,
			IsDaily:     true,
			TemplateID:  &templateID,
		}
		if err := e.DB.CreateQuest(q); err != nil {
			return spawned, err
		}
		spawned++
	}
	return spawned, nil
}

// ============================================================
// Statistics
// ============================================================

func (e *Engine) GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{
		QuestsByRank: make(map[models.QuestRank]int),
	}

	var err error

	stats.TotalQuestsCompleted, err = e.DB.GetTotalCompletedCount(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.TotalQuestsFailed, err = e.DB.GetTotalFailedCount(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.QuestsByRank, err = e.DB.GetCompletedCountByRank(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.TotalEXPEarned, err = e.DB.GetTotalEXPEarned(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.CurrentStreak, err = e.DB.GetStreak(e.Character.ID)
	if err != nil {
		return nil, err
	}

	statLevels, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}
	stats.StatLevels = statLevels

	// Calculate best stat
	bestLevel := 0
	for _, s := range statLevels {
		if s.Level > bestLevel {
			bestLevel = s.Level
			stats.BestStat = s.StatType
			stats.BestStatLevel = s.Level
		}
	}

	// Success rate
	total := stats.TotalQuestsCompleted + stats.TotalQuestsFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalQuestsCompleted) / float64(total) * 100
	}

	return stats, nil
}

// ============================================================
// Dungeons
// ============================================================

func (e *Engine) InitDungeons() error {
	count, err := e.DB.GetDungeonCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	dungeons := GetPresetDungeons()
	for i := range dungeons {
		if err := e.DB.InsertDungeon(&dungeons[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) RefreshDungeonStatuses() error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return err
	}

	statMap := make(map[models.StatType]int)
	for _, s := range stats {
		statMap[s.StatType] = s.Level
	}

	for _, d := range dungeons {
		if d.Status == models.DungeonCompleted || d.Status == models.DungeonInProgress {
			continue
		}

		meetsReqs := true
		for _, req := range d.Requirements {
			if statMap[req.StatType] < req.MinLevel {
				meetsReqs = false
				break
			}
		}

		newStatus := models.DungeonLocked
		if meetsReqs {
			newStatus = models.DungeonAvailable
		}

		if newStatus != d.Status {
			e.DB.UpdateDungeonStatus(d.ID, newStatus)
		}
	}
	return nil
}

func (e *Engine) EnterDungeon(dungeonID int64) error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	var dungeon *models.Dungeon
	for i := range dungeons {
		if dungeons[i].ID == dungeonID {
			dungeon = &dungeons[i]
			break
		}
	}
	if dungeon == nil {
		return fmt.Errorf("dungeon not found")
	}
	if dungeon.Status != models.DungeonAvailable {
		return fmt.Errorf("dungeon is not available")
	}

	// Create quests from dungeon quest definitions
	for _, qd := range dungeon.QuestDefinitions {
		did := dungeon.ID
		q := &models.Quest{
			CharID:      e.Character.ID,
			Title:       qd.Title,
			Description: qd.Description,
			Rank:        qd.Rank,
			TargetStat:  qd.TargetStat,
			DungeonID:   &did,
		}
		if err := e.DB.CreateQuest(q); err != nil {
			return err
		}
	}

	return e.DB.UpdateDungeonStatus(dungeonID, models.DungeonInProgress)
}

// CheckDungeonCompletion checks if all quests in a dungeon are completed
func (e *Engine) CheckDungeonCompletion(dungeonID int64) (bool, error) {
	allQuests, err := e.DB.GetDungeonAllQuests(e.Character.ID, dungeonID)
	if err != nil {
		return false, err
	}
	if len(allQuests) == 0 {
		return false, nil
	}

	for _, q := range allQuests {
		if q.Status != models.QuestCompleted {
			return false, nil
		}
	}
	return true, nil
}

// CompleteDungeon finalizes a dungeon, awards rewards
func (e *Engine) CompleteDungeon(dungeonID int64) error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	var dungeon *models.Dungeon
	for i := range dungeons {
		if dungeons[i].ID == dungeonID {
			dungeon = &dungeons[i]
			break
		}
	}
	if dungeon == nil {
		return fmt.Errorf("dungeon not found")
	}

	// Award bonus EXP to all stats
	if dungeon.RewardEXP > 0 {
		stats, err := e.GetStatLevels()
		if err != nil {
			return err
		}
		for i := range stats {
			stats[i].CurrentEXP += dungeon.RewardEXP
			stats[i].TotalEXP += dungeon.RewardEXP
			for {
				required := models.ExpForLevel(stats[i].Level)
				if stats[i].CurrentEXP >= required {
					stats[i].CurrentEXP -= required
					stats[i].Level++
				} else {
					break
				}
			}
			if err := e.DB.UpdateStatLevel(&stats[i]); err != nil {
				return err
			}
		}
	}

	if err := e.DB.UpdateDungeonStatus(dungeonID, models.DungeonCompleted); err != nil {
		return err
	}
	return e.DB.CompleteDungeon(e.Character.ID, dungeonID, dungeon.RewardTitle)
}

// GetDungeonProgress returns (completed, total) quest counts for a dungeon
func (e *Engine) GetDungeonProgress(dungeonID int64) (int, int, error) {
	allQuests, err := e.DB.GetDungeonAllQuests(e.Character.ID, dungeonID)
	if err != nil {
		return 0, 0, err
	}
	completed := 0
	for _, q := range allQuests {
		if q.Status == models.QuestCompleted {
			completed++
		}
	}
	return completed, len(allQuests), nil
}

// ============================================================
// Skills (unchanged)
// ============================================================

type SkillOption struct {
	Name        string
	Description string
	Multiplier  float64
}

func GetSkillOptions(stat models.StatType, level int) []SkillOption {
	catalog := map[models.StatType]map[int][]SkillOption{
		models.StatStrength: {
			3:  {{Name: "Железная хватка", Description: "Увеличивает получение EXP Силы", Multiplier: 1.10}},
			5:  {{Name: "Берсерк", Description: "Мощный прилив силы", Multiplier: 1.15}},
			8:  {{Name: "Титан", Description: "Сила титана течёт в венах", Multiplier: 1.20}},
			10: {{Name: "Разрушитель", Description: "Нет преград, которые не сломать", Multiplier: 1.25}},
			15: {{Name: "Монарх Силы", Description: "Абсолютная мощь", Multiplier: 1.35}},
		},
		models.StatAgility: {
			3:  {{Name: "Быстрые ноги", Description: "Увеличивает получение EXP Ловкости", Multiplier: 1.10}},
			5:  {{Name: "Тень", Description: "Движения быстрее взгляда", Multiplier: 1.15}},
			8:  {{Name: "Фантом", Description: "Неуловимый как призрак", Multiplier: 1.20}},
			10: {{Name: "Молния", Description: "Скорость молнии", Multiplier: 1.25}},
			15: {{Name: "Монарх Скорости", Description: "Время замедляется вокруг", Multiplier: 1.35}},
		},
		models.StatIntellect: {
			3:  {{Name: "Острый ум", Description: "Увеличивает получение EXP Интеллекта", Multiplier: 1.10}},
			5:  {{Name: "Аналитик", Description: "Видит паттерны во всём", Multiplier: 1.15}},
			8:  {{Name: "Стратег", Description: "На три шага впереди", Multiplier: 1.20}},
			10: {{Name: "Мудрец", Description: "Знания бесконечны", Multiplier: 1.25}},
			15: {{Name: "Монарх Разума", Description: "Абсолютный интеллект", Multiplier: 1.35}},
		},
		models.StatEndurance: {
			3:  {{Name: "Толстая кожа", Description: "Увеличивает получение EXP Выносливости", Multiplier: 1.10}},
			5:  {{Name: "Стойкость", Description: "Боль — лишь иллюзия", Multiplier: 1.15}},
			8:  {{Name: "Непробиваемый", Description: "Тело крепче стали", Multiplier: 1.20}},
			10: {{Name: "Бессмертный", Description: "Ничто не сломит волю", Multiplier: 1.25}},
			15: {{Name: "Монарх Воли", Description: "Абсолютная стойкость", Multiplier: 1.35}},
		},
	}

	if statCat, ok := catalog[stat]; ok {
		if options, ok := statCat[level]; ok {
			return options
		}
	}
	return nil
}

func (e *Engine) UnlockSkill(stat models.StatType, level int, optionIndex int) (*models.Skill, error) {
	options := GetSkillOptions(stat, level)
	if optionIndex < 0 || optionIndex >= len(options) {
		return nil, fmt.Errorf("invalid skill option index")
	}
	opt := options[optionIndex]
	skill := &models.Skill{
		CharID:      e.Character.ID,
		Name:        opt.Name,
		Description: opt.Description,
		StatType:    stat,
		Multiplier:  opt.Multiplier,
		UnlockedAt:  level,
	}
	if err := e.DB.CreateSkill(skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (e *Engine) GetSkills() ([]models.Skill, error) {
	return e.DB.GetSkills(e.Character.ID)
}

func (e *Engine) ToggleSkill(skillID int64, active bool) error {
	return e.DB.ToggleSkill(skillID, active)
}

func (e *Engine) RenameCharacter(name string) error {
	e.Character.Name = name
	return e.DB.UpdateCharacterName(e.Character.ID, name)
}

func HunterRank(level int) string {
	switch {
	case level >= 40:
		return "S-Ранг Охотник"
	case level >= 30:
		return "A-Ранг Охотник"
	case level >= 20:
		return "B-Ранг Охотник"
	case level >= 14:
		return "C-Ранг Охотник"
	case level >= 8:
		return "D-Ранг Охотник"
	default:
		return "E-Ранг Охотник"
	}
}

func HunterRankColor(level int) string {
	switch {
	case level >= 40:
		return "#e74c3c"
	case level >= 30:
		return "#e67e22"
	case level >= 20:
		return "#9b59b6"
	case level >= 14:
		return "#4a7fbf"
	case level >= 8:
		return "#4a9e4a"
	default:
		return "#8a8a8a"
	}
}

// ============================================================
// Enemies
// ============================================================

func GetPresetEnemies() []models.Enemy {
	return []models.Enemy{
		// Regular enemies
		{Name: "Гоблин", Description: "Слабый, но хитрый монстр", Rank: models.RankE, Type: models.EnemyRegular,
			HP: 80, Attack: 8, PatternSize: 4, ShowTime: 3.0, RewardEXP: 15, RewardCrystals: 10,
			DropMaterial: models.MaterialCommon, DropChance: 0.5},
		{Name: "Волк-тень", Description: "Быстрый хищник из тёмного мира", Rank: models.RankD, Type: models.EnemyRegular,
			HP: 120, Attack: 12, PatternSize: 5, ShowTime: 2.8, RewardEXP: 30, RewardCrystals: 20,
			DropMaterial: models.MaterialCommon, DropChance: 0.6},
		{Name: "Каменный голем", Description: "Медленный, но невероятно прочный", Rank: models.RankC, Type: models.EnemyRegular,
			HP: 200, Attack: 15, PatternSize: 6, ShowTime: 2.5, RewardEXP: 55, RewardCrystals: 35,
			DropMaterial: models.MaterialCommon, DropChance: 0.7},
		{Name: "Тёмный рыцарь", Description: "Опытный воин, павший во тьму", Rank: models.RankB, Type: models.EnemyRegular,
			HP: 300, Attack: 22, PatternSize: 7, ShowTime: 2.2, RewardEXP: 90, RewardCrystals: 55,
			DropMaterial: models.MaterialRare, DropChance: 0.4},
		{Name: "Демон-маг", Description: "Владеет разрушительной магией", Rank: models.RankA, Type: models.EnemyRegular,
			HP: 450, Attack: 30, PatternSize: 8, ShowTime: 2.0, RewardEXP: 150, RewardCrystals: 80,
			DropMaterial: models.MaterialRare, DropChance: 0.5},
		{Name: "Архидемон", Description: "Один из высших демонов, невероятная мощь", Rank: models.RankS, Type: models.EnemyRegular,
			HP: 600, Attack: 40, PatternSize: 10, ShowTime: 1.8, RewardEXP: 280, RewardCrystals: 130,
			DropMaterial: models.MaterialEpic, DropChance: 0.3},
		// Bosses
		{Name: "Игрис — Рыцарь Крови", Description: "Легендарный рыцарь, верный страж данжа. Его клинок не знает пощады.",
			Rank: models.RankB, Type: models.EnemyBoss,
			HP: 500, Attack: 28, PatternSize: 8, ShowTime: 2.5, RewardEXP: 200, RewardCrystals: 120,
			DropMaterial: models.MaterialRare, DropChance: 0.8},
		{Name: "Барука — Король муравьёв", Description: "Монструозный повелитель насекомых. Его панцирь почти непробиваем.",
			Rank: models.RankA, Type: models.EnemyBoss,
			HP: 700, Attack: 38, PatternSize: 10, ShowTime: 2.0, RewardEXP: 350, RewardCrystals: 200,
			DropMaterial: models.MaterialEpic, DropChance: 0.6},
		{Name: "Монарх Теней", Description: "Повелитель теней и смерти. Последнее испытание для сильнейших.",
			Rank: models.RankS, Type: models.EnemyBoss,
			HP: 1000, Attack: 50, PatternSize: 12, ShowTime: 1.5, RewardEXP: 500, RewardCrystals: 350,
			DropMaterial: models.MaterialEpic, DropChance: 0.9},
	}
}

func (e *Engine) InitEnemies() error {
	count, err := e.DB.GetEnemyCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	enemies := GetPresetEnemies()
	for i := range enemies {
		if err := e.DB.InsertEnemy(&enemies[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) GetEnemies() ([]models.Enemy, error) {
	return e.DB.GetAllEnemies()
}

// ============================================================
// Battle System
// ============================================================

func (e *Engine) StartBattle(enemyID int64) (*models.BattleState, error) {
	enemy, err := e.DB.GetEnemyByID(enemyID)
	if err != nil {
		return nil, err
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}

	// Get equipped items for bonuses
	equipped, err := e.DB.GetEquippedItems(e.Character.ID)
	if err != nil {
		return nil, err
	}

	// Base player HP from endurance
	endurance := 1
	for _, s := range stats {
		if s.StatType == models.StatEndurance {
			endurance = s.Level
		}
	}
	baseHP := 100 + endurance*10

	// Bonus HP from armor
	bonusHP := 0
	for _, eq := range equipped {
		bonusHP += eq.BonusHP
	}
	playerHP := baseHP + bonusHP

	// Generate pattern
	pattern := generatePattern(enemy.PatternSize, 64) // 8x8 = 64 cells

	return &models.BattleState{
		Enemy:       *enemy,
		PlayerHP:    playerHP,
		PlayerMaxHP: playerHP,
		EnemyHP:     enemy.HP,
		EnemyMaxHP:  enemy.HP,
		Round:       1,
		Pattern:     pattern,
	}, nil
}

func generatePattern(size, gridSize int) []int {
	if size > gridSize {
		size = gridSize
	}
	perm := rand.Perm(gridSize)
	return perm[:size]
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

	equipped, err := e.DB.GetEquippedItems(e.Character.ID)
	if err != nil {
		return err
	}

	statMap := make(map[models.StatType]int)
	for _, s := range stats {
		statMap[s.StatType] = s.Level
	}

	// Calculate accuracy: how many guesses match the pattern
	patternSet := make(map[int]bool)
	for _, p := range state.Pattern {
		patternSet[p] = true
	}

	hits := 0
	for _, g := range guesses {
		if patternSet[g] {
			hits++
		}
	}

	accuracy := 0.0
	if len(state.Pattern) > 0 {
		accuracy = float64(hits) / float64(len(state.Pattern))
	}

	// Base damage from strength
	strength := statMap[models.StatStrength]
	baseDamage := 10 + strength*3

	// Weapon bonus
	for _, eq := range equipped {
		baseDamage += eq.BonusAttack
	}

	// Damage scales with accuracy
	damage := int(float64(baseDamage) * accuracy)

	// Critical hit chance from intellect (intellect * 2%, max 40%)
	intellect := statMap[models.StatIntellect]
	critChance := math.Min(float64(intellect)*0.02, 0.40)
	crits := 0
	if rand.Float64() < critChance && accuracy > 0.5 {
		damage = int(float64(damage) * 1.5)
		crits = 1
	}

	// Enemy attacks
	enemyDamage := state.Enemy.Attack

	// Dodge chance from agility (agility * 2%, max 35%)
	agility := statMap[models.StatAgility]
	dodgeChance := math.Min(float64(agility)*0.02, 0.35)
	dodges := 0
	if rand.Float64() < dodgeChance {
		enemyDamage = 0
		dodges = 1
	}

	// Apply damage
	state.EnemyHP -= damage
	state.PlayerHP -= enemyDamage

	state.TotalHits += hits
	state.TotalMisses += len(state.Pattern) - hits
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
		state.Pattern = generatePattern(state.Enemy.PatternSize, 64)
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
		record.RewardEXP = state.Enemy.RewardEXP
		record.RewardCrystals = state.Enemy.RewardCrystals

		// Award EXP to strength stat
		expPerStat := state.Enemy.RewardEXP / 2
		stats, err := e.GetStatLevels()
		if err != nil {
			return nil, err
		}
		for i := range stats {
			if stats[i].StatType == models.StatStrength {
				stats[i].CurrentEXP += expPerStat
				stats[i].TotalEXP += expPerStat
				for {
					required := models.ExpForLevel(stats[i].Level)
					if stats[i].CurrentEXP >= required {
						stats[i].CurrentEXP -= required
						stats[i].Level++
					} else {
						break
					}
				}
				e.DB.UpdateStatLevel(&stats[i])
			}
		}

		// Award crystals
		e.DB.AddCrystals(e.Character.ID, state.Enemy.RewardCrystals)

		// Material drop
		if state.Enemy.DropMaterial != "" && rand.Float64() < state.Enemy.DropChance {
			record.MaterialDrop = string(state.Enemy.DropMaterial)
			e.DB.AddMaterial(e.Character.ID, state.Enemy.DropMaterial, 1)
		}

		// Record daily activity
		e.DB.RecordDailyActivity(e.Character.ID, 0, 0, expPerStat)
	}

	_ = totalRounds
	if err := e.DB.InsertBattle(record); err != nil {
		return nil, err
	}

	return record, nil
}

// GetShowTime returns adjusted show time based on accessory bonuses
func (e *Engine) GetShowTime(baseTime float64) (float64, error) {
	equipped, err := e.DB.GetEquippedItems(e.Character.ID)
	if err != nil {
		return baseTime, err
	}
	bonus := 0.0
	for _, eq := range equipped {
		bonus += eq.BonusTime
	}
	return baseTime + bonus, nil
}

// ============================================================
// Resources
// ============================================================

func (e *Engine) GetResources() (*models.PlayerResources, error) {
	return e.DB.GetOrCreateResources(e.Character.ID)
}

func (e *Engine) InitResources() error {
	_, err := e.DB.GetOrCreateResources(e.Character.ID)
	return err
}

// ============================================================
// Gacha System
// ============================================================

var equipmentNames = map[models.EquipmentSlot][]string{
	models.SlotWeapon: {
		"Кинжал тени", "Клинок рассвета", "Меч демона", "Копьё судьбы",
		"Секира хаоса", "Молот грома", "Жезл тьмы", "Посох мудреца",
	},
	models.SlotArmor: {
		"Кираса охотника", "Мантия тени", "Доспех дракона", "Роба мага",
		"Латы стража", "Плащ невидимости", "Панцирь титана", "Облачение монарха",
	},
	models.SlotAccessory: {
		"Кольцо силы", "Амулет удачи", "Ожерелье мудрости", "Браслет скорости",
		"Серьга тьмы", "Пояс титана", "Корона монарха", "Перчатки ловкости",
	},
}

func (e *Engine) GachaPull(banner models.GachaBanner) (*models.Equipment, error) {
	res, err := e.DB.GetOrCreateResources(e.Character.ID)
	if err != nil {
		return nil, err
	}

	cost := banner.Cost()
	if res.Crystals < cost {
		return nil, fmt.Errorf("недостаточно кристаллов: нужно %d, есть %d", cost, res.Crystals)
	}

	res.Crystals -= cost
	if err := e.DB.UpdateResources(res); err != nil {
		return nil, err
	}

	pity, err := e.DB.GetOrCreateGachaPity(e.Character.ID)
	if err != nil {
		return nil, err
	}

	// Determine rarity
	rarity := e.rollGachaRarity(banner, pity)

	// Reset or increment pity
	switch banner {
	case models.BannerNormal:
		if rarity >= models.RarityRare {
			pity.NormalPity = 0
		} else {
			pity.NormalPity++
		}
	case models.BannerAdvanced:
		if rarity >= models.RarityEpic {
			pity.AdvancedPity = 0
		} else {
			pity.AdvancedPity++
		}
	}
	e.DB.UpdateGachaPity(e.Character.ID, pity)

	// Generate equipment
	slot := randomSlot()
	eq := e.generateEquipment(slot, rarity)

	if err := e.DB.InsertEquipment(eq); err != nil {
		return nil, err
	}

	// Record gacha history
	h := &models.GachaHistory{
		CharID:      e.Character.ID,
		Banner:      banner,
		EquipmentID: eq.ID,
		Rarity:      rarity,
	}
	e.DB.InsertGachaHistory(h)

	return eq, nil
}

func (e *Engine) GachaMultiPull(banner models.GachaBanner, count int) ([]*models.Equipment, error) {
	var results []*models.Equipment
	for i := 0; i < count; i++ {
		eq, err := e.GachaPull(banner)
		if err != nil {
			return results, err
		}
		results = append(results, eq)
	}
	return results, nil
}

func (e *Engine) rollGachaRarity(banner models.GachaBanner, pity *models.GachaPity) models.EquipmentRarity {
	roll := rand.Float64() * 100

	switch banner {
	case models.BannerNormal:
		// Pity: guaranteed rare at 30 pulls
		if pity.NormalPity >= 29 {
			return models.RarityRare
		}
		switch {
		case roll < 1: // 1% legendary
			return models.RarityLegendary
		case roll < 5: // 4% epic
			return models.RarityEpic
		case roll < 15: // 10% rare
			return models.RarityRare
		case roll < 40: // 25% uncommon
			return models.RarityUncommon
		default: // 60% common
			return models.RarityCommon
		}

	case models.BannerAdvanced:
		// Pity: guaranteed epic at 50 pulls
		if pity.AdvancedPity >= 49 {
			return models.RarityEpic
		}
		switch {
		case roll < 3: // 3% legendary
			return models.RarityLegendary
		case roll < 13: // 10% epic
			return models.RarityEpic
		case roll < 33: // 20% rare
			return models.RarityRare
		case roll < 63: // 30% uncommon
			return models.RarityUncommon
		default: // 37% common
			return models.RarityCommon
		}
	}

	return models.RarityCommon
}

func randomSlot() models.EquipmentSlot {
	slots := []models.EquipmentSlot{models.SlotWeapon, models.SlotArmor, models.SlotAccessory}
	return slots[rand.Intn(len(slots))]
}

func (e *Engine) generateEquipment(slot models.EquipmentSlot, rarity models.EquipmentRarity) *models.Equipment {
	names := equipmentNames[slot]
	name := names[rand.Intn(len(names))]

	baseStats := rarity.BaseStats()

	eq := &models.Equipment{
		CharID: e.Character.ID,
		Name:   name,
		Slot:   slot,
		Rarity: rarity,
		Level:  1,
	}

	switch slot {
	case models.SlotWeapon:
		eq.BonusAttack = baseStats
	case models.SlotArmor:
		eq.BonusHP = baseStats * 5
	case models.SlotAccessory:
		eq.BonusTime = float64(baseStats) * 0.1
	}

	return eq
}

// ============================================================
// Equipment Management
// ============================================================

func (e *Engine) GetEquipment() ([]models.Equipment, error) {
	return e.DB.GetAllEquipment(e.Character.ID)
}

func (e *Engine) EquipItem(equipmentID int64) error {
	eq, err := e.DB.GetEquipmentByID(equipmentID)
	if err != nil {
		return err
	}
	// Unequip current item in same slot
	if err := e.DB.UnequipSlot(e.Character.ID, eq.Slot); err != nil {
		return err
	}
	eq.Equipped = true
	return e.DB.UpdateEquipment(eq)
}

func (e *Engine) UnequipItem(equipmentID int64) error {
	eq, err := e.DB.GetEquipmentByID(equipmentID)
	if err != nil {
		return err
	}
	eq.Equipped = false
	return e.DB.UpdateEquipment(eq)
}

func (e *Engine) DismantleEquipment(equipmentID int64) (int, models.MaterialTier, int, error) {
	eq, err := e.DB.GetEquipmentByID(equipmentID)
	if err != nil {
		return 0, "", 0, err
	}

	crystals := eq.Rarity.DismantleCrystals()
	matTier, matCount := eq.Rarity.DismantleMaterial()

	// Add level bonus
	crystals += (eq.Level - 1) * 5

	// Add resources
	e.DB.AddCrystals(e.Character.ID, crystals)
	if matCount > 0 {
		e.DB.AddMaterial(e.Character.ID, matTier, matCount)
	}

	// Delete equipment
	e.DB.DeleteEquipment(equipmentID)

	return crystals, matTier, matCount, nil
}

func (e *Engine) SellEquipment(equipmentID int64) (int, error) {
	eq, err := e.DB.GetEquipmentByID(equipmentID)
	if err != nil {
		return 0, err
	}

	crystals := eq.Rarity.DismantleCrystals() / 2
	if crystals < 1 {
		crystals = 1
	}
	crystals += (eq.Level - 1) * 2

	e.DB.AddCrystals(e.Character.ID, crystals)
	e.DB.DeleteEquipment(equipmentID)

	return crystals, nil
}

// UpgradeEquipment feeds one equipment into another for EXP
func (e *Engine) UpgradeEquipment(targetID, feedID int64) (*models.Equipment, bool, error) {
	target, err := e.DB.GetEquipmentByID(targetID)
	if err != nil {
		return nil, false, err
	}
	feed, err := e.DB.GetEquipmentByID(feedID)
	if err != nil {
		return nil, false, err
	}

	// EXP from feeding: base stats * 10 + feed level * 5
	expGain := feed.Rarity.BaseStats()*10 + feed.Level*5

	target.CurrentEXP += expGain
	leveledUp := false

	for {
		required := models.EquipmentEXPForLevel(target.Level)
		if target.CurrentEXP >= required {
			target.CurrentEXP -= required
			target.Level++
			leveledUp = true
			// Increase stats on level up
			switch target.Slot {
			case models.SlotWeapon:
				target.BonusAttack += 1 + target.Rarity.BaseStats()/5
			case models.SlotArmor:
				target.BonusHP += 3 + target.Rarity.BaseStats()/2
			case models.SlotAccessory:
				target.BonusTime += 0.05
			}
		} else {
			break
		}
	}

	if err := e.DB.UpdateEquipment(target); err != nil {
		return nil, false, err
	}
	e.DB.DeleteEquipment(feedID)

	return target, leveledUp, nil
}

// ============================================================
// Crafting System
// ============================================================

func GetPresetRecipes() []models.CraftRecipe {
	return []models.CraftRecipe{
		{Name: "Ковка: Обычное оружие", ResultSlot: models.SlotWeapon, ResultRarity: models.RarityCommon,
			CostCrystals: 50, CostCommon: 5},
		{Name: "Ковка: Необычное оружие", ResultSlot: models.SlotWeapon, ResultRarity: models.RarityUncommon,
			CostCrystals: 150, CostCommon: 15, CostRare: 3},
		{Name: "Ковка: Редкое оружие", ResultSlot: models.SlotWeapon, ResultRarity: models.RarityRare,
			CostCrystals: 400, CostCommon: 20, CostRare: 10, CostEpic: 2},
		{Name: "Ковка: Эпическое оружие", ResultSlot: models.SlotWeapon, ResultRarity: models.RarityEpic,
			CostCrystals: 1000, CostRare: 20, CostEpic: 8},
		{Name: "Ковка: Обычная броня", ResultSlot: models.SlotArmor, ResultRarity: models.RarityCommon,
			CostCrystals: 50, CostCommon: 5},
		{Name: "Ковка: Необычная броня", ResultSlot: models.SlotArmor, ResultRarity: models.RarityUncommon,
			CostCrystals: 150, CostCommon: 15, CostRare: 3},
		{Name: "Ковка: Редкая броня", ResultSlot: models.SlotArmor, ResultRarity: models.RarityRare,
			CostCrystals: 400, CostCommon: 20, CostRare: 10, CostEpic: 2},
		{Name: "Ковка: Эпическая броня", ResultSlot: models.SlotArmor, ResultRarity: models.RarityEpic,
			CostCrystals: 1000, CostRare: 20, CostEpic: 8},
		{Name: "Ковка: Обычный аксессуар", ResultSlot: models.SlotAccessory, ResultRarity: models.RarityCommon,
			CostCrystals: 50, CostCommon: 5},
		{Name: "Ковка: Необычный аксессуар", ResultSlot: models.SlotAccessory, ResultRarity: models.RarityUncommon,
			CostCrystals: 150, CostCommon: 15, CostRare: 3},
		{Name: "Ковка: Редкий аксессуар", ResultSlot: models.SlotAccessory, ResultRarity: models.RarityRare,
			CostCrystals: 400, CostCommon: 20, CostRare: 10, CostEpic: 2},
		{Name: "Ковка: Эпический аксессуар", ResultSlot: models.SlotAccessory, ResultRarity: models.RarityEpic,
			CostCrystals: 1000, CostRare: 20, CostEpic: 8},
	}
}

func (e *Engine) InitRecipes() error {
	count, err := e.DB.GetCraftRecipeCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	recipes := GetPresetRecipes()
	for i := range recipes {
		if err := e.DB.InsertCraftRecipe(&recipes[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) GetRecipes() ([]models.CraftRecipe, error) {
	return e.DB.GetAllCraftRecipes()
}

func (e *Engine) CraftItem(recipeID int64) (*models.Equipment, error) {
	recipes, err := e.DB.GetAllCraftRecipes()
	if err != nil {
		return nil, err
	}

	var recipe *models.CraftRecipe
	for i := range recipes {
		if recipes[i].ID == recipeID {
			recipe = &recipes[i]
			break
		}
	}
	if recipe == nil {
		return nil, fmt.Errorf("рецепт не найден")
	}

	res, err := e.DB.GetOrCreateResources(e.Character.ID)
	if err != nil {
		return nil, err
	}

	// Check resources
	if res.Crystals < recipe.CostCrystals {
		return nil, fmt.Errorf("недостаточно кристаллов")
	}
	if res.MaterialCommon < recipe.CostCommon {
		return nil, fmt.Errorf("недостаточно обычных материалов")
	}
	if res.MaterialRare < recipe.CostRare {
		return nil, fmt.Errorf("недостаточно редких материалов")
	}
	if res.MaterialEpic < recipe.CostEpic {
		return nil, fmt.Errorf("недостаточно эпических материалов")
	}

	// Deduct resources
	res.Crystals -= recipe.CostCrystals
	res.MaterialCommon -= recipe.CostCommon
	res.MaterialRare -= recipe.CostRare
	res.MaterialEpic -= recipe.CostEpic
	if err := e.DB.UpdateResources(res); err != nil {
		return nil, err
	}

	// Generate equipment
	eq := e.generateEquipment(recipe.ResultSlot, recipe.ResultRarity)
	if err := e.DB.InsertEquipment(eq); err != nil {
		return nil, err
	}

	return eq, nil
}

// ============================================================
// Daily Rewards
// ============================================================

func (e *Engine) CanClaimDailyReward() (bool, error) {
	return e.DB.CanClaimDailyReward(e.Character.ID)
}

func (e *Engine) ClaimDailyReward() (*models.DailyReward, error) {
	canClaim, err := e.DB.CanClaimDailyReward(e.Character.ID)
	if err != nil {
		return nil, err
	}
	if !canClaim {
		return nil, fmt.Errorf("ежедневная награда уже получена сегодня")
	}

	streak, err := e.DB.GetDailyRewardStreak(e.Character.ID)
	if err != nil {
		return nil, err
	}

	// Check if streak is continuous
	last, err := e.DB.GetLastDailyReward(e.Character.ID)
	if err != nil {
		return nil, err
	}

	nextDay := streak + 1
	if last != nil {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		lastDay := last.ClaimedAt.Format("2006-01-02")
		if lastDay != yesterday {
			nextDay = 1 // streak broken
		}
	}

	crystals := models.DailyRewardCrystals(nextDay)

	reward := &models.DailyReward{
		CharID:   e.Character.ID,
		Day:      nextDay,
		Crystals: crystals,
	}

	if err := e.DB.InsertDailyReward(reward); err != nil {
		return nil, err
	}

	// Add crystals
	e.DB.AddCrystals(e.Character.ID, crystals)

	return reward, nil
}

func (e *Engine) GetDailyRewardInfo() (bool, int, int, error) {
	canClaim, err := e.DB.CanClaimDailyReward(e.Character.ID)
	if err != nil {
		return false, 0, 0, err
	}

	streak, err := e.DB.GetDailyRewardStreak(e.Character.ID)
	if err != nil {
		return false, 0, 0, err
	}

	nextDay := streak + 1
	last, err := e.DB.GetLastDailyReward(e.Character.ID)
	if err != nil {
		return false, 0, 0, err
	}
	if last != nil {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		lastDay := last.ClaimedAt.Format("2006-01-02")
		if lastDay != yesterday && lastDay != time.Now().Format("2006-01-02") {
			nextDay = 1
		}
	}

	crystals := models.DailyRewardCrystals(nextDay)
	return canClaim, nextDay, crystals, nil
}

// ============================================================
// Battle & Gacha Statistics
// ============================================================

func (e *Engine) GetBattleStats() (*models.BattleStatistics, error) {
	return e.DB.GetBattleStats(e.Character.ID)
}

func (e *Engine) GetGachaStats() (*models.GachaStatistics, error) {
	return e.DB.GetGachaStats(e.Character.ID)
}

func (e *Engine) GetGachaHistory(limit int) ([]models.GachaHistory, error) {
	return e.DB.GetGachaHistory(e.Character.ID, limit)
}

func (e *Engine) GetBattleHistory(limit int) ([]models.BattleRecord, error) {
	return e.DB.GetBattleHistory(e.Character.ID, limit)
}

func (e *Engine) GetGachaPity() (*models.GachaPity, error) {
	return e.DB.GetOrCreateGachaPity(e.Character.ID)
}
