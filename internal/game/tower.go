package game

import (
	"fmt"

	"solo-leveling/internal/models"
)

// GetNextEnemyForPlayer returns the first enemy in sequence that is not yet cleared.
// If all enemies are cleared, it returns nil, nil.
func (e *Engine) GetNextEnemyForPlayer() (*models.Enemy, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return nil, err
	}
	if len(allEnemies) == 0 {
		return nil, nil
	}

	// Always keep the first enemy unlocked as an entry point.
	_ = e.DB.UnlockEnemy(e.Character.ID, allEnemies[0].ID)

	for i := range allEnemies {
		reward, err := e.DB.GetBattleReward(e.Character.ID, allEnemies[i].ID)
		if err != nil {
			return nil, err
		}
		if reward == nil {
			_ = e.DB.UnlockEnemy(e.Character.ID, allEnemies[i].ID)
			enemy := allEnemies[i]
			return &enemy, nil
		}
	}
	return nil, nil
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

func (e *Engine) unlockNextEnemy(currentEnemyID int64) (string, error) {
	allEnemies, err := e.DB.GetAllEnemies()
	if err != nil {
		return "", err
	}
	for i := range allEnemies {
		if allEnemies[i].ID == currentEnemyID && i+1 < len(allEnemies) {
			next := allEnemies[i+1]
			if err := e.DB.UnlockEnemy(e.Character.ID, next.ID); err != nil {
				return "", err
			}
			return next.Name, nil
		}
	}
	return "", nil
}
