package game

import (
	"testing"
	"time"

	"solo-leveling/internal/models"
)

func TestSpawnDailyQuests_NoDuplicateSameDay(t *testing.T) {
	e := newTestEngine(t)

	daily, err := e.CreateQuest("Daily Seed", "test", "", 20, models.StatIntellect, true)
	if err != nil {
		t.Fatalf("create daily quest: %v", err)
	}
	if daily.TemplateID == nil {
		t.Fatal("expected daily template id")
	}

	spawned, err := e.SpawnDailyQuests()
	if err != nil {
		t.Fatalf("spawn daily quests: %v", err)
	}
	if spawned != 0 {
		t.Fatalf("expected 0 spawned on same day, got %d", spawned)
	}

	spawnedAgain, err := e.SpawnDailyQuests()
	if err != nil {
		t.Fatalf("spawn daily quests again: %v", err)
	}
	if spawnedAgain != 0 {
		t.Fatalf("expected 0 spawned on second run, got %d", spawnedAgain)
	}

	active, err := e.DB.GetActiveQuests(e.Character.ID)
	if err != nil {
		t.Fatalf("get active quests: %v", err)
	}
	countForTemplate := 0
	for _, q := range active {
		if q.TemplateID != nil && *q.TemplateID == *daily.TemplateID {
			countForTemplate++
		}
	}
	if countForTemplate != 1 {
		t.Fatalf("expected exactly 1 active daily for template, got %d", countForTemplate)
	}
}

func TestAutoFailUnfinishedQuests_FailsStaleMainAndDaily(t *testing.T) {
	e := newTestEngine(t)

	mainQ, err := e.CreateQuest("Main Quest", "test", "", 24, models.StatStrength, false)
	if err != nil {
		t.Fatalf("create main quest: %v", err)
	}

	dailyQ, err := e.CreateQuest("Daily Quest", "test", "", 22, models.StatIntellect, true)
	if err != nil {
		t.Fatalf("create daily quest: %v", err)
	}
	if dailyQ.TemplateID == nil {
		t.Fatal("expected daily template id")
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	if err := e.DB.SetQuestCreatedAt(mainQ.ID, yesterday); err != nil {
		t.Fatalf("set main created_at: %v", err)
	}
	if err := e.DB.SetQuestCreatedAt(dailyQ.ID, yesterday); err != nil {
		t.Fatalf("set daily created_at: %v", err)
	}

	failed, err := e.AutoFailUnfinishedQuests()
	if err != nil {
		t.Fatalf("auto-fail unfinished: %v", err)
	}
	if failed != 2 {
		t.Fatalf("expected 2 failed quests, got %d", failed)
	}

	gotMain, err := e.DB.GetQuestByID(mainQ.ID)
	if err != nil {
		t.Fatalf("get main quest: %v", err)
	}
	if gotMain.Status != models.QuestFailed {
		t.Fatalf("expected main quest status failed, got %s", gotMain.Status)
	}

	gotDaily, err := e.DB.GetQuestByID(dailyQ.ID)
	if err != nil {
		t.Fatalf("get daily quest: %v", err)
	}
	if gotDaily.Status != models.QuestFailed {
		t.Fatalf("expected daily quest status failed, got %s", gotDaily.Status)
	}

	spawned, err := e.SpawnDailyQuests()
	if err != nil {
		t.Fatalf("spawn daily after autofail: %v", err)
	}
	if spawned != 1 {
		t.Fatalf("expected 1 spawned daily after autofail, got %d", spawned)
	}
}
