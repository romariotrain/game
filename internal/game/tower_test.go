package game

import (
	"testing"

	"solo-leveling/internal/database"
	"solo-leveling/internal/models"
)

func newTestEngine(t *testing.T) *Engine {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)

	db, err := database.New()
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	engine, err := NewEngine(db)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if err := engine.InitEnemies(); err != nil {
		t.Fatalf("init enemies: %v", err)
	}
	return engine
}

func battleWinState(t *testing.T, e *Engine, enemyID int64) *models.BattleState {
	t.Helper()

	state, err := e.StartBattle(enemyID)
	if err != nil {
		t.Fatalf("start battle: %v", err)
	}
	state.BattleOver = true
	state.Result = models.BattleWin
	return state
}

func totalStatEXP(stats []models.StatLevel) int {
	total := 0
	for _, s := range stats {
		total += s.TotalEXP
	}
	return total
}

func TestTowerWinUnlocksNextEnemy(t *testing.T) {
	e := newTestEngine(t)

	if _, err := e.DB.AddAttempts(e.Character.ID, models.MaxAttempts); err != nil {
		t.Fatalf("add attempts: %v", err)
	}

	current, err := e.GetCurrentEnemy()
	if err != nil {
		t.Fatalf("get current enemy: %v", err)
	}
	if current == nil {
		t.Fatal("expected current enemy")
	}
	firstEnemyID := current.ID

	record, err := e.FinishBattle(battleWinState(t, e, firstEnemyID))
	if err != nil {
		t.Fatalf("finish battle: %v", err)
	}
	if record.Result != models.BattleWin {
		t.Fatalf("expected win, got %s", record.Result)
	}

	next, err := e.GetCurrentEnemy()
	if err != nil {
		t.Fatalf("get next current enemy: %v", err)
	}
	if next == nil {
		t.Fatal("expected next current enemy after first win")
	}
	if next.ID == firstEnemyID {
		t.Fatal("expected next enemy to differ from cleared enemy")
	}

	reward, err := e.DB.GetBattleReward(e.Character.ID, firstEnemyID)
	if err != nil {
		t.Fatalf("get reward: %v", err)
	}
	if reward == nil {
		t.Fatal("expected first-win reward to be created")
	}

	unlocked, err := e.DB.GetUnlockedEnemyIDs(e.Character.ID)
	if err != nil {
		t.Fatalf("get unlocked enemies: %v", err)
	}
	if !unlocked[next.ID] {
		t.Fatal("expected next enemy to be unlocked")
	}
}

func TestTowerPreventsFarmingClearedEnemy(t *testing.T) {
	e := newTestEngine(t)

	if _, err := e.DB.AddAttempts(e.Character.ID, models.MaxAttempts); err != nil {
		t.Fatalf("add attempts: %v", err)
	}

	current, err := e.GetCurrentEnemy()
	if err != nil {
		t.Fatalf("get current enemy: %v", err)
	}
	if current == nil {
		t.Fatal("expected current enemy")
	}
	firstEnemyID := current.ID

	if _, err := e.FinishBattle(battleWinState(t, e, firstEnemyID)); err != nil {
		t.Fatalf("finish battle: %v", err)
	}

	before := e.GetAttempts()
	if _, err := e.StartBattle(firstEnemyID); err == nil {
		t.Fatal("expected error when trying to farm a cleared enemy")
	}
	after := e.GetAttempts()
	if after != before {
		t.Fatalf("attempts changed on rejected battle: before=%d after=%d", before, after)
	}
}

func TestTowerSpendsAttemptsOncePerBattle(t *testing.T) {
	e := newTestEngine(t)

	if _, err := e.DB.AddAttempts(e.Character.ID, 1); err != nil {
		t.Fatalf("add attempts: %v", err)
	}
	current, err := e.GetCurrentEnemy()
	if err != nil {
		t.Fatalf("get current enemy: %v", err)
	}
	if current == nil {
		t.Fatal("expected current enemy")
	}

	if _, err := e.StartBattle(current.ID); err != nil {
		t.Fatalf("start battle with one attempt: %v", err)
	}
	if got := e.GetAttempts(); got != 0 {
		t.Fatalf("expected attempts to be 0 after spending one, got %d", got)
	}

	if _, err := e.StartBattle(current.ID); err == nil {
		t.Fatal("expected error when no attempts left")
	}
	if got := e.GetAttempts(); got != 0 {
		t.Fatalf("attempts must not go negative, got %d", got)
	}
}

func TestBattleDoesNotGrantEXP(t *testing.T) {
	e := newTestEngine(t)

	if _, err := e.DB.AddAttempts(e.Character.ID, 1); err != nil {
		t.Fatalf("add attempts: %v", err)
	}
	current, err := e.GetCurrentEnemy()
	if err != nil {
		t.Fatalf("get current enemy: %v", err)
	}
	if current == nil {
		t.Fatal("expected current enemy")
	}

	beforeStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats before battle: %v", err)
	}
	beforeEXP := totalStatEXP(beforeStats)

	if _, err := e.FinishBattle(battleWinState(t, e, current.ID)); err != nil {
		t.Fatalf("finish battle: %v", err)
	}

	afterStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats after battle: %v", err)
	}
	afterEXP := totalStatEXP(afterStats)
	if afterEXP != beforeEXP {
		t.Fatalf("battle changed EXP: before=%d after=%d", beforeEXP, afterEXP)
	}
}
