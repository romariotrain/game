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

	endurance := statMap[models.StatEndurance]
	if endurance == 0 {
		endurance = 1
	}
	baseHP := 100 + endurance*10
	playerHP := baseHP

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

func (e *Engine) ProcessBossPuzzleInput(state *boss.State, value int) error {
	return boss.ApplyPuzzleInput(state, value)
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

	if err := e.DB.InsertBattle(record); err != nil {
		return nil, err
	}

	// Add a small delay to mark battle timing variance in history
	rand.Seed(time.Now().UnixNano())

	return record, nil
}
