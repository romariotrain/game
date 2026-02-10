package game

import (
	"solo-leveling/internal/models"
)

// ============================================================
// Statistics
// ============================================================

func (e *Engine) GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{
		QuestsByRank: make(map[models.QuestRank]int),
	}

	var err error

	stats.TotalQuestsCompleted, err = e.DB.GetTotalCompletedCount(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.TotalQuestsFailed, err = e.DB.GetTotalFailedCount(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.QuestsByRank, err = e.DB.GetCompletedCountByRank(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.TotalEXPEarned, err = e.DB.GetTotalEXPEarned(e.Character.ID)
	if err != nil {
		return nil, err
	}

	stats.CurrentStreak, err = e.DB.GetStreak(e.Character.ID)
	if err != nil {
		return nil, err
	}

	statLevels, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}
	stats.StatLevels = statLevels

	// Calculate best stat
	bestLevel := 0
	for _, s := range statLevels {
		if s.Level > bestLevel {
			bestLevel = s.Level
			stats.BestStat = s.StatType
			stats.BestStatLevel = s.Level
		}
	}

	// Success rate
	total := stats.TotalQuestsCompleted + stats.TotalQuestsFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalQuestsCompleted) / float64(total) * 100
	}

	return stats, nil
}
