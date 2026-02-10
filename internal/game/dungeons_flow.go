package game

import (
	"fmt"

	"solo-leveling/internal/models"
)

// ============================================================
// Dungeons
// ============================================================

func (e *Engine) InitDungeons() error {
	count, err := e.DB.GetDungeonCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	dungeons := GetPresetDungeons()
	for i := range dungeons {
		if err := e.DB.InsertDungeon(&dungeons[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) RefreshDungeonStatuses() error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return err
	}

	statMap := make(map[models.StatType]int)
	for _, s := range stats {
		statMap[s.StatType] = s.Level
	}

	for _, d := range dungeons {
		if d.Status == models.DungeonCompleted || d.Status == models.DungeonInProgress {
			continue
		}

		meetsReqs := true
		for _, req := range d.Requirements {
			if statMap[req.StatType] < req.MinLevel {
				meetsReqs = false
				break
			}
		}

		newStatus := models.DungeonLocked
		if meetsReqs {
			newStatus = models.DungeonAvailable
		}

		if newStatus != d.Status {
			e.DB.UpdateDungeonStatus(d.ID, newStatus)
		}
	}
	return nil
}

func (e *Engine) EnterDungeon(dungeonID int64) error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	var dungeon *models.Dungeon
	for i := range dungeons {
		if dungeons[i].ID == dungeonID {
			dungeon = &dungeons[i]
			break
		}
	}
	if dungeon == nil {
		return fmt.Errorf("dungeon not found")
	}
	if dungeon.Status != models.DungeonAvailable {
		return fmt.Errorf("dungeon is not available")
	}

	// Create quests from dungeon quest definitions
	for _, qd := range dungeon.QuestDefinitions {
		did := dungeon.ID
		q := &models.Quest{
			CharID:      e.Character.ID,
			Title:       qd.Title,
			Description: qd.Description,
			Exp:         qd.Exp,
			Rank:        models.RankFromEXP(qd.Exp),
			TargetStat:  qd.TargetStat,
			DungeonID:   &did,
		}
		if err := e.DB.CreateQuest(q); err != nil {
			return err
		}
	}

	return e.DB.UpdateDungeonStatus(dungeonID, models.DungeonInProgress)
}

// CheckDungeonCompletion checks if all quests in a dungeon are completed
func (e *Engine) CheckDungeonCompletion(dungeonID int64) (bool, error) {
	allQuests, err := e.DB.GetDungeonAllQuests(e.Character.ID, dungeonID)
	if err != nil {
		return false, err
	}
	if len(allQuests) == 0 {
		return false, nil
	}

	for _, q := range allQuests {
		if q.Status != models.QuestCompleted {
			return false, nil
		}
	}
	return true, nil
}

// CompleteDungeon finalizes a dungeon, awards rewards
func (e *Engine) CompleteDungeon(dungeonID int64) error {
	dungeons, err := e.DB.GetAllDungeons()
	if err != nil {
		return err
	}

	var dungeon *models.Dungeon
	for i := range dungeons {
		if dungeons[i].ID == dungeonID {
			dungeon = &dungeons[i]
			break
		}
	}
	if dungeon == nil {
		return fmt.Errorf("dungeon not found")
	}

	// Award bonus EXP to all stats
	if dungeon.RewardEXP > 0 {
		stats, err := e.GetStatLevels()
		if err != nil {
			return err
		}
		for i := range stats {
			stats[i].CurrentEXP += dungeon.RewardEXP
			stats[i].TotalEXP += dungeon.RewardEXP
			for {
				required := models.ExpForLevel(stats[i].Level)
				if stats[i].CurrentEXP >= required {
					stats[i].CurrentEXP -= required
					stats[i].Level++
				} else {
					break
				}
			}
			if err := e.DB.UpdateStatLevel(&stats[i]); err != nil {
				return err
			}
		}
	}

	if err := e.DB.UpdateDungeonStatus(dungeonID, models.DungeonCompleted); err != nil {
		return err
	}
	return e.DB.CompleteDungeon(e.Character.ID, dungeonID, dungeon.RewardTitle)
}

// GetDungeonProgress returns (completed, total) quest counts for a dungeon
func (e *Engine) GetDungeonProgress(dungeonID int64) (int, int, error) {
	allQuests, err := e.DB.GetDungeonAllQuests(e.Character.ID, dungeonID)
	if err != nil {
		return 0, 0, err
	}
	completed := 0
	for _, q := range allQuests {
		if q.Status == models.QuestCompleted {
			completed++
		}
	}
	return completed, len(allQuests), nil
}
