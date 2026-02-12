package tabs

import (
	"fyne.io/fyne/v2"

	"solo-leveling/internal/config"
	"solo-leveling/internal/game"
	"solo-leveling/internal/models"
)

// Context provides shared state and services for tab builders.
type Context struct {
	Engine   *game.Engine
	Window   fyne.Window
	App      fyne.App
	Features config.Features

	CharacterPanel    *fyne.Container
	QuestsPanel       *fyne.Container
	StatsPanel        *fyne.Container
	DungeonsPanel     *fyne.Container
	AchievementsPanel *fyne.Container

	RefreshAll          func()
	RefreshCharacter    func()
	RefreshQuests       func()
	RefreshStats        func()
	RefreshDungeons     func()
	RefreshAchievements func()
	RefreshHistory      func()
	StartBattle         func(enemy models.Enemy)
	QuestThemeMode      string
}
