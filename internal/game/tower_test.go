package game

import (
	"sort"
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

func isBossEnemy(enemy models.Enemy) bool {
	return enemy.IsBoss || enemy.Type == models.EnemyBoss
}

func zoneEnemies(t *testing.T, e *Engine, zone int) ([]models.Enemy, models.Enemy) {
	t.Helper()

	all, err := e.DB.GetAllEnemies()
	if err != nil {
		t.Fatalf("get enemies: %v", err)
	}
	var regular []models.Enemy
	var boss models.Enemy
	foundBoss := false
	for _, enemy := range all {
		if enemy.Zone != zone {
			continue
		}
		if isBossEnemy(enemy) {
			boss = enemy
			foundBoss = true
			continue
		}
		regular = append(regular, enemy)
	}
	if !foundBoss {
		t.Fatalf("zone %d boss not found", zone)
	}
	sort.Slice(regular, func(i, j int) bool { return regular[i].ID < regular[j].ID })
	return regular, boss
}

func markDefeated(t *testing.T, e *Engine, enemyID int64) {
	t.Helper()
	defeated, err := e.DB.GetDefeatedEnemyIDs(e.Character.ID)
	if err != nil {
		t.Fatalf("get defeated ids: %v", err)
	}
	if defeated[enemyID] {
		return
	}
	enemy, err := e.DB.GetEnemyByID(enemyID)
	if err != nil {
		t.Fatalf("get enemy by id: %v", err)
	}
	if err := e.DB.InsertBattle(&models.BattleRecord{
		CharID:    e.Character.ID,
		EnemyID:   enemyID,
		EnemyName: enemy.Name,
		Result:    models.BattleWin,
	}); err != nil {
		t.Fatalf("insert battle win: %v", err)
	}
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

	defeated, err := e.DB.GetDefeatedEnemyIDs(e.Character.ID)
	if err != nil {
		t.Fatalf("get defeated ids: %v", err)
	}
	if !defeated[firstEnemyID] {
		t.Fatal("expected first enemy to be marked as defeated by battle win")
	}

	if next.Zone < current.Zone {
		t.Fatalf("next enemy moved backwards by zone: current=%d next=%d", current.Zone, next.Zone)
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

func TestPickNextEnemyReturnsRegularWhenRegularRemaining(t *testing.T) {
	e := newTestEngine(t)

	regular, _ := zoneEnemies(t, e, 1)
	if len(regular) == 0 {
		t.Fatal("expected regular enemies in zone 1")
	}

	for i := 1; i < len(regular); i++ {
		markDefeated(t, e, regular[i].ID)
	}

	next, err := e.PickNextEnemy(e.Character.ID)
	if err != nil {
		t.Fatalf("pick next enemy: %v", err)
	}
	if next == nil {
		t.Fatal("expected next enemy")
	}
	if next.ID != regular[0].ID {
		t.Fatalf("expected regular enemy id=%d, got id=%d", regular[0].ID, next.ID)
	}
	if isBossEnemy(*next) {
		t.Fatal("expected regular enemy, got boss")
	}
}

func TestPickNextEnemyReturnsBossWhenZoneRegularDefeated(t *testing.T) {
	e := newTestEngine(t)

	regular, boss := zoneEnemies(t, e, 1)
	for _, enemy := range regular {
		markDefeated(t, e, enemy.ID)
	}

	next, err := e.PickNextEnemy(e.Character.ID)
	if err != nil {
		t.Fatalf("pick next enemy: %v", err)
	}
	if next == nil {
		t.Fatal("expected next enemy")
	}
	if next.ID != boss.ID {
		t.Fatalf("expected boss id=%d, got id=%d", boss.ID, next.ID)
	}
	if !isBossEnemy(*next) {
		t.Fatal("expected boss enemy")
	}
}

func TestCurrentZoneAdvancesAfterBossDefeated(t *testing.T) {
	e := newTestEngine(t)

	regular, boss := zoneEnemies(t, e, 1)
	for _, enemy := range regular {
		markDefeated(t, e, enemy.ID)
	}
	markDefeated(t, e, boss.ID)

	zone, err := e.GetCurrentZone(e.Character.ID)
	if err != nil {
		t.Fatalf("get current zone: %v", err)
	}
	if zone != 2 {
		t.Fatalf("expected current zone 2, got %d", zone)
	}
}

func TestZoneHasOneBoss(t *testing.T) {
	e := newTestEngine(t)

	all, err := e.DB.GetAllEnemies()
	if err != nil {
		t.Fatalf("get enemies: %v", err)
	}

	bossByZone := map[int]int{}
	for _, enemy := range all {
		if isBossEnemy(enemy) {
			bossByZone[enemy.Zone]++
		}
	}

	for zone := 1; zone <= 5; zone++ {
		if bossByZone[zone] != 1 {
			t.Fatalf("expected exactly 1 boss in zone %d, got %d", zone, bossByZone[zone])
		}
	}
}
