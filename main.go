package main

import (
	"log"

	fyneApp "fyne.io/fyne/v2/app"

	"solo-leveling/internal/config"
	"solo-leveling/internal/database"
	"solo-leveling/internal/game"
	"solo-leveling/internal/sim"
	"solo-leveling/internal/ui"
)

func main() {
	// Headless simulation mode — no DB, no UI
	if sim.RunCLIAutoTune() {
		return
	}
	if sim.RunCLI() {
		return
	}
	if sim.RunCLICompact() {
		return
	}

	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	engine, err := game.NewEngine(db)
	if err != nil {
		log.Fatalf("Failed to initialize game engine: %v", err)
	}

	// Seed preset dungeons if not yet created
	if err := engine.InitDungeons(); err != nil {
		log.Printf("Warning: failed to init dungeons: %v", err)
	}

	// Spawn daily quests for today
	spawned, err := engine.SpawnDailyQuests()
	if err != nil {
		log.Printf("Warning: failed to spawn daily quests: %v", err)
	}
	if spawned > 0 {
		log.Printf("Spawned %d daily quests", spawned)
	}

	// Refresh dungeon availability based on current stats
	if err := engine.RefreshDungeonStatuses(); err != nil {
		log.Printf("Warning: failed to refresh dungeon statuses: %v", err)
	}

	// Seed preset enemies if not yet created
	if err := engine.InitEnemies(); err != nil {
		log.Printf("Warning: failed to init enemies: %v", err)
	}

	application := fyneApp.New()
	application.Settings().SetTheme(&ui.SoloLevelingTheme{})

	features := config.DefaultFeatures()
	appUI := ui.NewApp(application, engine, features)
	appUI.Run()
}
