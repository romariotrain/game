package game

import (
	"fmt"
	"sort"

	"solo-leveling/internal/models"
)

// GetCurrentZone returns the first zone where the boss is not defeated.
// If all bosses are defeated, it returns the highest known zone.
func (e *Engine) GetCurrentZone(userID int64) (int, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return 1, err
	}
	if len(allEnemies) == 0 {
		return 1, nil
	}

	defeated, err := e.defeatedEnemyIDs(userID)
	if err != nil {
		return 1, err
	}

	minOpenZone := 0
	maxZone := 1
	hasBoss := false
	for _, enemy := range allEnemies {
		if enemy.Zone > maxZone {
			maxZone = enemy.Zone
		}
		if !isEnemyBoss(enemy) {
			continue
		}
		hasBoss = true
		if defeated[enemy.ID] {
			continue
		}
		if minOpenZone == 0 || enemy.Zone < minOpenZone {
			minOpenZone = enemy.Zone
		}
	}

	if minOpenZone > 0 {
		return minOpenZone, nil
	}
	if !hasBoss {
		return 1, nil
	}
	return maxZone, nil
}

// IsBossAvailable returns true when all regular enemies in the zone are defeated.
func (e *Engine) IsBossAvailable(userID int64, zone int) (bool, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return false, err
	}
	defeated, err := e.defeatedEnemyIDs(userID)
	if err != nil {
		return false, err
	}

	for _, enemy := range allEnemies {
		if enemy.Zone != zone || isEnemyBoss(enemy) {
			continue
		}
		if !defeated[enemy.ID] {
			return false, nil
		}
	}
	return true, nil
}

// PickNextEnemy chooses a single next enemy for Today:
// regular enemies in current zone first, then the zone boss.
func (e *Engine) PickNextEnemy(userID int64) (*models.Enemy, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return nil, err
	}
	if len(allEnemies) == 0 {
		return nil, nil
	}

	defeated, err := e.defeatedEnemyIDs(userID)
	if err != nil {
		return nil, err
	}

	zone, err := e.GetCurrentZone(userID)
	if err != nil {
		return nil, err
	}

	var zoneEnemies []models.Enemy
	for _, enemy := range allEnemies {
		if enemy.Zone == zone {
			zoneEnemies = append(zoneEnemies, enemy)
		}
	}
	if len(zoneEnemies) == 0 {
		return nil, nil
	}

	sort.Slice(zoneEnemies, func(i, j int) bool {
		return zoneEnemies[i].ID < zoneEnemies[j].ID
	})

	var boss *models.Enemy
	for i := range zoneEnemies {
		if isEnemyBoss(zoneEnemies[i]) {
			candidate := zoneEnemies[i]
			boss = &candidate
			break
		}
	}

	if boss != nil && !defeated[boss.ID] {
		available, err := e.IsBossAvailable(userID, zone)
		if err != nil {
			return nil, err
		}
		if available {
			return boss, nil
		}
	}

	for i := range zoneEnemies {
		enemy := zoneEnemies[i]
		if isEnemyBoss(enemy) {
			continue
		}
		if !defeated[enemy.ID] {
			candidate := enemy
			return &candidate, nil
		}
	}

	if boss != nil && !defeated[boss.ID] {
		return boss, nil
	}
	return nil, nil
}

// GetNextEnemyForPlayer keeps compatibility with existing callers.
func (e *Engine) GetNextEnemyForPlayer() (*models.Enemy, error) {
	return e.PickNextEnemy(e.Character.ID)
}

// GetCurrentEnemy keeps backward compatibility with existing callers/tests.
func (e *Engine) GetCurrentEnemy() (*models.Enemy, error) {
	return e.GetNextEnemyForPlayer()
}

// CanFightCurrentEnemy checks if the provided enemy is the current target
// and the player has at least one battle attempt.
func (e *Engine) CanFightCurrentEnemy(enemyID int64) (bool, error) {
	current, err := e.GetNextEnemyForPlayer()
	if err != nil {
		return false, err
	}
	if current == nil || current.ID != enemyID {
		return false, nil
	}
	return e.GetAttempts() > 0, nil
}

func (e *Engine) validateCurrentEnemyForFight(enemyID int64) (*models.Enemy, error) {
	current, err := e.GetNextEnemyForPlayer()
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("все враги уже побеждены")
	}
	if current.ID != enemyID {
		return nil, fmt.Errorf("доступен только текущий враг")
	}
	return current, nil
}

func (e *Engine) spendBattleAttempt() error {
	if err := e.DB.SpendAttempt(e.Character.ID); err != nil {
		return fmt.Errorf("нет попыток для боя! Выполняйте задания, чтобы получить попытки")
	}
	attempts, err := e.DB.GetAttempts(e.Character.ID)
	if err == nil {
		e.Character.Attempts = attempts
	}
	return nil
}

func (e *Engine) defeatedEnemyIDs(userID int64) (map[int64]bool, error) {
	return e.DB.GetDefeatedEnemyIDs(userID)
}

func isEnemyBoss(enemy models.Enemy) bool {
	return enemy.IsBoss || enemy.Type == models.EnemyBoss
}
