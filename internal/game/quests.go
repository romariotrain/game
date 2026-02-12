package game

import (
	"fmt"
	"time"

	"solo-leveling/internal/models"
)

func (e *Engine) GetStatLevels() ([]models.StatLevel, error) {
	return e.DB.GetStatLevels(e.Character.ID)
}

func (e *Engine) GetOverallLevel() (int, error) {
	stats, err := e.GetStatLevels()
	if err != nil {
		return 0, err
	}
	total := 0
	for _, s := range stats {
		total += s.Level
	}
	return total / len(stats), nil
}

func (e *Engine) GetEXPMultiplier(stat models.StatType) (float64, error) {
	skills, err := e.DB.GetSkills(e.Character.ID)
	if err != nil {
		return 1.0, err
	}
	mult := 1.0
	for _, s := range skills {
		if s.Active && s.StatType == stat {
			mult *= s.Multiplier
		}
	}
	return mult, nil
}

type CompleteResult struct {
	EXPAwarded      int
	LeveledUp       bool
	OldLevel        int
	NewLevel        int
	StatType        models.StatType
	AttemptsAwarded int
	TotalAttempts   int
}

func (e *Engine) CompleteQuest(questID int64) (*CompleteResult, error) {
	active, err := e.DB.GetActiveQuests(e.Character.ID)
	if err != nil {
		return nil, err
	}

	var quest *models.Quest
	for i := range active {
		if active[i].ID == questID {
			quest = &active[i]
			break
		}
	}
	if quest == nil {
		return nil, fmt.Errorf("quest not found or not active")
	}

	expAwarded := quest.Exp
	if expAwarded <= 0 {
		expAwarded = 1
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}

	var stat *models.StatLevel
	for i := range stats {
		if stats[i].StatType == quest.TargetStat {
			stat = &stats[i]
			break
		}
	}
	if stat == nil {
		return nil, fmt.Errorf("stat not found: %s", quest.TargetStat)
	}

	oldLevel := stat.Level
	stat.CurrentEXP += expAwarded
	stat.TotalEXP += expAwarded

	for {
		required := models.ExpForLevel(stat.Level)
		if stat.CurrentEXP >= required {
			stat.CurrentEXP -= required
			stat.Level++
		} else {
			break
		}
	}

	if err := e.DB.UpdateStatLevel(stat); err != nil {
		return nil, err
	}
	if err := e.DB.CompleteQuest(questID); err != nil {
		return nil, err
	}

	// Record daily activity
	e.DB.RecordDailyActivity(e.Character.ID, 1, 0, expAwarded)

	// Award battle attempts based on quest EXP.
	attemptsAwarded := models.AttemptsForQuestEXP(quest.Exp)
	totalAttempts, _ := e.DB.AddAttempts(e.Character.ID, attemptsAwarded)
	e.Character.Attempts = totalAttempts

	// First completion achievement (idempotent).
	if err := e.UnlockAchievement(AchievementFirstTask); err != nil {
		return nil, err
	}

	// Check streak milestones
	e.CheckStreakMilestones()

	return &CompleteResult{
		EXPAwarded:      expAwarded,
		LeveledUp:       stat.Level > oldLevel,
		OldLevel:        oldLevel,
		NewLevel:        stat.Level,
		StatType:        quest.TargetStat,
		AttemptsAwarded: attemptsAwarded,
		TotalAttempts:   totalAttempts,
	}, nil
}

func (e *Engine) CreateQuest(title, description, congratulations string, exp int, targetStat models.StatType, isDaily bool) (*models.Quest, error) {
	if exp <= 0 {
		exp = 1
	}
	q := &models.Quest{
		CharID:          e.Character.ID,
		Title:           title,
		Description:     description,
		Congratulations: congratulations,
		Exp:             exp,
		Rank:            models.RankFromEXP(exp),
		TargetStat:      targetStat,
		IsDaily:         isDaily,
	}

	// If daily, create a template first
	if isDaily {
		tmpl := &models.DailyQuestTemplate{
			CharID:          e.Character.ID,
			Title:           title,
			Description:     description,
			Congratulations: congratulations,
			Exp:             exp,
			Rank:            models.RankFromEXP(exp),
			TargetStat:      targetStat,
		}
		if err := e.DB.CreateDailyTemplate(tmpl); err != nil {
			return nil, err
		}
		q.TemplateID = &tmpl.ID
	}

	if err := e.DB.CreateQuest(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (e *Engine) FailQuest(questID int64) error {
	// Record daily activity for failure
	e.DB.RecordDailyActivity(e.Character.ID, 0, 1, 0)
	return e.DB.FailQuest(questID)
}

func (e *Engine) DeleteQuest(questID int64) error {
	// If it's a daily quest with a template, disable the template too
	q, err := e.DB.GetQuestByID(questID)
	if err == nil && q.IsDaily && q.TemplateID != nil {
		e.DB.DisableDailyTemplate(*q.TemplateID)
	}
	return e.DB.DeleteQuest(questID)
}

// CheckStreakMilestones awards titles for streak milestones
func (e *Engine) CheckStreakMilestones() {
	streak, err := e.DB.GetStreak(e.Character.ID)
	if err != nil {
		return
	}
	if streak >= 7 {
		_ = e.UnlockAchievement(AchievementStreak7)
	}
	for _, m := range models.AllStreakMilestones() {
		if streak >= m.Days {
			e.DB.InsertStreakTitle(e.Character.ID, m.Title, m.Days)
		}
	}
}

// GetAttempts returns current battle attempts
func (e *Engine) GetAttempts() int {
	attempts, err := e.DB.GetAttempts(e.Character.ID)
	if err != nil {
		return e.Character.Attempts
	}
	e.Character.Attempts = attempts
	return attempts
}

// AutoFailUnfinishedQuests marks unfinished main/daily quests from previous days as failed.
func (e *Engine) AutoFailUnfinishedQuests() (int, error) {
	active, err := e.DB.GetActiveQuests(e.Character.ID)
	if err != nil {
		return 0, err
	}

	today := time.Now().Format("2006-01-02")
	failed := 0
	for _, q := range active {
		// Keep dungeon chains untouched; fail only regular/daily quest flow.
		if q.DungeonID != nil {
			continue
		}
		if q.CreatedAt.Format("2006-01-02") == today {
			continue
		}
		if err := e.DB.FailQuest(q.ID); err != nil {
			return failed, err
		}
		failed++
	}

	if failed > 0 {
		_ = e.DB.RecordDailyActivity(e.Character.ID, 0, failed, 0)
	}
	return failed, nil
}

// SpawnDailyQuests checks all active daily templates and creates today's quests if missing
func (e *Engine) SpawnDailyQuests() (int, error) {
	templates, err := e.DB.GetActiveDailyTemplates(e.Character.ID)
	if err != nil {
		return 0, err
	}

	spawned := 0
	for _, tmpl := range templates {
		exists, err := e.DB.HasDailyQuestForToday(e.Character.ID, tmpl.ID)
		if err != nil {
			return spawned, err
		}
		if exists {
			continue
		}

		templateID := tmpl.ID
		q := &models.Quest{
			CharID:          e.Character.ID,
			Title:           tmpl.Title,
			Description:     tmpl.Description,
			Congratulations: tmpl.Congratulations,
			Exp:             tmpl.Exp,
			Rank:            tmpl.Rank,
			TargetStat:      tmpl.TargetStat,
			IsDaily:         true,
			TemplateID:      &templateID,
		}
		if err := e.DB.CreateQuest(q); err != nil {
			return spawned, err
		}
		spawned++
	}
	return spawned, nil
}
