package game

import (
	"fmt"

	"solo-leveling/internal/models"
)

// Legacy recommendations from the old profile flow are intentionally disabled.
// JSON import of generated tasks is available in the Quests tab.
func (e *Engine) suggestQuestOptionsLLM(
	_ int,
	_ *models.HunterProfile,
	_ []models.StatLevel,
	_ models.QuestRank,
	_ int,
	_ map[string]bool,
) ([]models.QuestSuggestion, string, error) {
	return nil, "", fmt.Errorf("legacy LLM flow disabled; use local AI recommendations in Today")
}
