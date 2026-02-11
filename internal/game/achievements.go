package game

import (
	"log"

	"solo-leveling/internal/models"
)

const (
	AchievementFirstTask    = "first_task"
	AchievementFirstBattle  = "first_battle"
	AchievementStreak7      = "streak_7"
	AchievementFirstDungeon = "first_dungeon"
)

func defaultAchievements() []models.Achievement {
	return []models.Achievement{
		{
			Key:         AchievementFirstTask,
			Title:       "Первое задание",
			Description: "Первая отметка в архиве пути охотника.",
			Category:    "discipline",
		},
		{
			Key:         AchievementFirstBattle,
			Title:       "Первая победа",
			Description: "Первый победный бой в башне.",
			Category:    "combat",
		},
		{
			Key:         AchievementStreak7,
			Title:       "Семидневный ритм",
			Description: "Стабильный темп на протяжении семи дней.",
			Category:    "streak",
		},
		{
			Key:         AchievementFirstDungeon,
			Title:       "Первый данж",
			Description: "Первый закрытый данж в истории охотника.",
			Category:    "dungeon",
		},
	}
}

func (e *Engine) InitAchievements() error {
	return e.DB.SeedAchievements(defaultAchievements())
}

func (e *Engine) UnlockAchievement(key string) error {
	unlocked, err := e.DB.UnlockAchievement(key)
	if err != nil {
		return err
	}
	if unlocked {
		log.Printf("achievement unlocked: %s", key)
	}
	return nil
}

func (e *Engine) GetAchievements() ([]models.Achievement, error) {
	return e.DB.GetAchievements()
}
