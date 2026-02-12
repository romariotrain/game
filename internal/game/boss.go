package game

import (
	"fmt"
	"math/rand"
	"time"

	"solo-leveling/internal/game/combat/boss"
	"solo-leveling/internal/game/combat/memory"
	"solo-leveling/internal/models"
)

func (e *Engine) StartBossBattle(enemyID int64) (*boss.State, error) {
	enemy, err := e.validateCurrentEnemyForFight(enemyID)
	if err != nil {
		return nil, err
	}
	if enemy.Type != models.EnemyBoss {
		return nil, fmt.Errorf("босс не выбран: текущий враг не является боссом")
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

	return boss.NewState(*enemy, memStats, playerHP)
}

func (e *Engine) ProcessBossMemory(state *boss.State, guesses []int) error {
	stats, err := e.GetStatLevels()
	if err != nil {
		return err
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

	return boss.ApplyMemoryInput(state, guesses, memStats, 0)
}

func (e *Engine) FailBoss(state *boss.State) (*models.BattleRecord, error) {
	record := &models.BattleRecord{
		CharID:       e.Character.ID,
		EnemyID:      state.Enemy.ID,
		EnemyName:    state.Enemy.Name,
		Result:       models.BattleLose,
		DamageDealt:  state.DamageDealt,
		DamageTaken:  state.DamageTaken,
		Accuracy:     boss.CalcAccuracy(state.TotalHits, state.TotalMisses),
		CriticalHits: 0,
		Dodges:       0,
	}
	if err := e.DB.InsertBattle(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (e *Engine) FinishBoss(state *boss.State) (*models.BattleRecord, error) {
	record := &models.BattleRecord{
		CharID:       e.Character.ID,
		EnemyID:      state.Enemy.ID,
		EnemyName:    state.Enemy.Name,
		Result:       models.BattleWin,
		DamageDealt:  state.DamageDealt,
		DamageTaken:  state.DamageTaken,
		Accuracy:     boss.CalcAccuracy(state.TotalHits, state.TotalMisses),
		CriticalHits: 0,
		Dodges:       0,
	}

	existing, err := e.DB.GetBattleReward(e.Character.ID, state.Enemy.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		title := fmt.Sprintf("Босс-победа: %s", state.Enemy.Name)
		badge := fmt.Sprintf("Босс: %s", state.Enemy.Name)
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

	if err := e.DB.InsertBattle(record); err != nil {
		return nil, err
	}

	// Add a small delay to mark battle timing variance in history
	rand.Seed(time.Now().UnixNano())

	return record, nil
}
