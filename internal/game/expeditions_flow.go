package game

import (
	"fmt"
	"time"

	"solo-leveling/internal/models"
)

// ============================================================
// Expeditions
// ============================================================

func (e *Engine) InitExpeditions() error {
	count, err := e.DB.GetExpeditionCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	expeditions := GetPresetExpeditions()
	for i := range expeditions {
		if err := e.DB.InsertExpedition(&expeditions[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) RefreshExpeditionStatuses(failExpired bool) error {
	expeditions, err := e.DB.GetAllExpeditions()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, ex := range expeditions {
		if ex.Status != models.ExpeditionActive {
			continue
		}
		if failExpired && ex.Deadline != nil && now.After(*ex.Deadline) {
			if err := e.DB.UpdateExpeditionStatus(ex.ID, models.ExpeditionFailed); err != nil {
				return err
			}
			if err := e.DB.FailActiveQuestsByExpedition(e.Character.ID, ex.ID); err != nil {
				return err
			}
			continue
		}

		done, err := e.CheckExpeditionCompletion(ex.ID)
		if err != nil {
			return err
		}
		if done {
			if err := e.CompleteExpedition(ex.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Engine) StartExpedition(expeditionID int64) (int, error) {
	expedition, err := e.DB.GetExpeditionByID(expeditionID)
	if err != nil {
		return 0, err
	}

	if expedition.Status == models.ExpeditionFailed {
		if !expedition.IsRepeatable {
			return 0, fmt.Errorf("expedition failed and is not repeatable")
		}
		if err := e.DB.ResetExpeditionTasks(expeditionID); err != nil {
			return 0, err
		}
		if err := e.DB.UpdateExpeditionStatus(expeditionID, models.ExpeditionActive); err != nil {
			return 0, err
		}
	}

	if expedition.Status == models.ExpeditionCompleted {
		if !expedition.IsRepeatable {
			return 0, fmt.Errorf("expedition already completed")
		}
		if err := e.DB.ResetExpeditionTasks(expeditionID); err != nil {
			return 0, err
		}
		if err := e.DB.UpdateExpeditionStatus(expeditionID, models.ExpeditionActive); err != nil {
			return 0, err
		}
	}

	tasks, err := e.DB.GetExpeditionTasks(expeditionID)
	if err != nil {
		return 0, err
	}

	spawned := 0
	for _, task := range tasks {
		if task.IsCompleted {
			continue
		}
		exists, err := e.DB.HasActiveQuestForExpeditionTask(e.Character.ID, task.ID)
		if err != nil {
			return spawned, err
		}
		if exists {
			continue
		}
		if err := e.createExpeditionQuest(expeditionID, task); err != nil {
			return spawned, err
		}
		spawned++
	}

	return spawned, nil
}

func (e *Engine) createExpeditionQuest(expeditionID int64, task models.ExpeditionTask) error {
	expID := expeditionID
	taskID := task.ID
	q := &models.Quest{
		CharID:           e.Character.ID,
		Title:            task.Title,
		Description:      task.Description,
		Exp:              task.RewardEXP,
		Rank:             models.RankFromEXP(task.RewardEXP),
		TargetStat:       task.TargetStat,
		ExpeditionID:     &expID,
		ExpeditionTaskID: &taskID,
	}
	return e.DB.CreateQuest(q)
}

func (e *Engine) resolveExpeditionTaskForQuest(q models.Quest) (*models.ExpeditionTask, error) {
	if q.ExpeditionTaskID != nil {
		task, err := e.DB.GetExpeditionTaskByID(*q.ExpeditionTaskID)
		if err == nil {
			if task.ExpeditionID == *q.ExpeditionID {
				return task, nil
			}
		}
	}
	return e.DB.FindNextIncompleteExpeditionTaskByTitle(*q.ExpeditionID, q.Title)
}

func (e *Engine) AdvanceExpeditionByQuest(q models.Quest) (*models.Expedition, bool, error) {
	if q.ExpeditionID == nil {
		return nil, false, nil
	}

	expedition, err := e.DB.GetExpeditionByID(*q.ExpeditionID)
	if err != nil {
		return nil, false, err
	}
	if expedition.Status != models.ExpeditionActive {
		return expedition, false, nil
	}

	task, err := e.resolveExpeditionTaskForQuest(q)
	if err != nil {
		return expedition, false, err
	}
	if task == nil {
		return expedition, false, nil
	}

	updatedTask, err := e.DB.IncrementExpeditionTaskProgress(task.ID, 1)
	if err != nil {
		return expedition, false, err
	}
	if !updatedTask.IsCompleted {
		if err := e.createExpeditionQuest(*q.ExpeditionID, *updatedTask); err != nil {
			return expedition, false, err
		}
	}

	done, err := e.CheckExpeditionCompletion(*q.ExpeditionID)
	if err != nil {
		return expedition, false, err
	}
	if !done {
		expedition, _ = e.DB.GetExpeditionByID(*q.ExpeditionID)
		return expedition, false, nil
	}

	if err := e.CompleteExpedition(*q.ExpeditionID); err != nil {
		return expedition, false, err
	}
	expedition, _ = e.DB.GetExpeditionByID(*q.ExpeditionID)
	return expedition, true, nil
}

// CheckExpeditionCompletion returns true when all expedition tasks are complete.
func (e *Engine) CheckExpeditionCompletion(expeditionID int64) (bool, error) {
	tasks, err := e.DB.GetExpeditionTasks(expeditionID)
	if err != nil {
		return false, err
	}
	if len(tasks) == 0 {
		return false, nil
	}
	for _, task := range tasks {
		if !task.IsCompleted {
			return false, nil
		}
	}
	return true, nil
}

// CompleteExpedition finalizes an expedition, applies rewards and stores completion.
func (e *Engine) CompleteExpedition(expeditionID int64) error {
	expedition, err := e.DB.GetExpeditionByID(expeditionID)
	if err != nil {
		return err
	}
	if expedition.Status == models.ExpeditionCompleted {
		return nil
	}
	if expedition.Status == models.ExpeditionFailed {
		return fmt.Errorf("expedition is failed")
	}

	done, err := e.CheckExpeditionCompletion(expeditionID)
	if err != nil {
		return err
	}
	if !done {
		return nil
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return err
	}

	if expedition.RewardEXP > 0 {
		for i := range stats {
			applyEXPToStat(&stats[i], expedition.RewardEXP)
		}
	}

	for statType, exp := range expedition.RewardStats {
		if exp <= 0 {
			continue
		}
		for i := range stats {
			if stats[i].StatType == statType {
				applyEXPToStat(&stats[i], exp)
				break
			}
		}
	}

	for i := range stats {
		if err := e.DB.UpdateStatLevel(&stats[i]); err != nil {
			return err
		}
	}

	if err := e.DB.UpdateExpeditionStatus(expeditionID, models.ExpeditionCompleted); err != nil {
		return err
	}
	if err := e.DB.CompleteExpedition(e.Character.ID, expeditionID); err != nil {
		return err
	}
	return e.UnlockAchievement(AchievementFirstExpedition)
}

// GetExpeditionProgress returns completed task count, total task count and completion percentage.
func (e *Engine) GetExpeditionProgress(expeditionID int64) (int, int, float64, error) {
	tasks, err := e.DB.GetExpeditionTasks(expeditionID)
	if err != nil {
		return 0, 0, 0, err
	}
	if len(tasks) == 0 {
		return 0, 0, 0, nil
	}
	completed := 0
	for _, task := range tasks {
		if task.IsCompleted {
			completed++
		}
	}
	percent := float64(completed) * 100.0 / float64(len(tasks))
	return completed, len(tasks), percent, nil
}

func applyEXPToStat(stat *models.StatLevel, exp int) {
	if exp <= 0 {
		return
	}
	stat.CurrentEXP += exp
	stat.TotalEXP += exp
	for {
		required := models.ExpForLevel(stat.Level)
		if stat.CurrentEXP < required {
			break
		}
		stat.CurrentEXP -= required
		stat.Level++
	}
}
