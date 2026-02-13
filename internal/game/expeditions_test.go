package game

import (
	"testing"
	"time"

	"solo-leveling/internal/models"
)

func statTotalsByType(stats []models.StatLevel) map[models.StatType]int {
	res := make(map[models.StatType]int, len(stats))
	for _, s := range stats {
		res[s.StatType] = s.TotalEXP
	}
	return res
}

func TestExpeditionCompletionGrantsRewards(t *testing.T) {
	e := newTestEngine(t)

	beforeStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats before: %v", err)
	}
	before := statTotalsByType(beforeStats)

	expedition := models.Expedition{
		Name:         "Тестовая экспедиция",
		Description:  "Проверка завершения",
		RewardEXP:    40,
		RewardStats:  map[models.StatType]int{models.StatStrength: 30},
		IsRepeatable: false,
		Status:       models.ExpeditionActive,
		Tasks: []models.ExpeditionTask{
			{Title: "Силовая задача", Description: "", ProgressTarget: 1, RewardEXP: 20, TargetStat: models.StatStrength},
			{Title: "Интеллект задача", Description: "", ProgressTarget: 1, RewardEXP: 20, TargetStat: models.StatIntellect},
		},
	}
	if err := e.DB.InsertExpedition(&expedition); err != nil {
		t.Fatalf("insert expedition: %v", err)
	}

	spawned, err := e.StartExpedition(expedition.ID)
	if err != nil {
		t.Fatalf("start expedition: %v", err)
	}
	if spawned != 2 {
		t.Fatalf("expected 2 spawned quests, got %d", spawned)
	}

	active, err := e.DB.GetExpeditionActiveQuests(e.Character.ID, expedition.ID)
	if err != nil {
		t.Fatalf("get expedition quests: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active expedition quests, got %d", len(active))
	}

	for _, q := range active {
		if _, err := e.CompleteQuest(q.ID); err != nil {
			t.Fatalf("complete quest %d: %v", q.ID, err)
		}
	}

	afterExpedition, err := e.DB.GetExpeditionByID(expedition.ID)
	if err != nil {
		t.Fatalf("get expedition after completion: %v", err)
	}
	if afterExpedition.Status != models.ExpeditionCompleted {
		t.Fatalf("expected expedition status completed, got %s", afterExpedition.Status)
	}

	completed, err := e.DB.GetCompletedExpeditions(e.Character.ID)
	if err != nil {
		t.Fatalf("get completed expeditions: %v", err)
	}
	found := false
	for _, c := range completed {
		if c.ExpeditionID == expedition.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected completed expedition record")
	}

	afterStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats after: %v", err)
	}
	after := statTotalsByType(afterStats)

	if got := after[models.StatStrength] - before[models.StatStrength]; got < 90 {
		t.Fatalf("expected STR gain >= 90, got %d", got)
	}
	if got := after[models.StatIntellect] - before[models.StatIntellect]; got < 60 {
		t.Fatalf("expected INT gain >= 60, got %d", got)
	}
	if got := after[models.StatAgility] - before[models.StatAgility]; got < 40 {
		t.Fatalf("expected AGI gain >= 40, got %d", got)
	}
	if got := after[models.StatEndurance] - before[models.StatEndurance]; got < 40 {
		t.Fatalf("expected STA gain >= 40, got %d", got)
	}
}

func TestExpeditionDeadlineFailureNoRewards(t *testing.T) {
	e := newTestEngine(t)

	beforeStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats before: %v", err)
	}
	before := statTotalsByType(beforeStats)

	deadline := time.Now().Add(-2 * time.Hour)
	expedition := models.Expedition{
		Name:         "Просроченная экспедиция",
		Description:  "Должна провалиться по дедлайну",
		Deadline:     &deadline,
		RewardEXP:    500,
		RewardStats:  map[models.StatType]int{models.StatStrength: 500},
		IsRepeatable: false,
		Status:       models.ExpeditionActive,
		Tasks: []models.ExpeditionTask{
			{Title: "Задача", Description: "", ProgressTarget: 1, RewardEXP: 20, TargetStat: models.StatStrength},
		},
	}
	if err := e.DB.InsertExpedition(&expedition); err != nil {
		t.Fatalf("insert expedition: %v", err)
	}

	_, err = e.StartExpedition(expedition.ID)
	if err != nil {
		t.Fatalf("start expedition: %v", err)
	}

	if err := e.RefreshExpeditionStatuses(true); err != nil {
		t.Fatalf("refresh expedition statuses: %v", err)
	}

	afterExpedition, err := e.DB.GetExpeditionByID(expedition.ID)
	if err != nil {
		t.Fatalf("get expedition after refresh: %v", err)
	}
	if afterExpedition.Status != models.ExpeditionFailed {
		t.Fatalf("expected expedition failed, got %s", afterExpedition.Status)
	}

	activeQuests, err := e.DB.GetExpeditionActiveQuests(e.Character.ID, expedition.ID)
	if err != nil {
		t.Fatalf("get active expedition quests: %v", err)
	}
	if len(activeQuests) != 0 {
		t.Fatalf("expected 0 active expedition quests after fail, got %d", len(activeQuests))
	}

	completed, err := e.DB.GetCompletedExpeditions(e.Character.ID)
	if err != nil {
		t.Fatalf("get completed expeditions: %v", err)
	}
	for _, c := range completed {
		if c.ExpeditionID == expedition.ID {
			t.Fatal("did not expect completed expedition record for failed expedition")
		}
	}

	afterStats, err := e.GetStatLevels()
	if err != nil {
		t.Fatalf("get stats after: %v", err)
	}
	after := statTotalsByType(afterStats)
	for statType, beforeTotal := range before {
		if after[statType] != beforeTotal {
			t.Fatalf("expected no %s EXP change on failed expedition, got %d -> %d", statType, beforeTotal, after[statType])
		}
	}
}
