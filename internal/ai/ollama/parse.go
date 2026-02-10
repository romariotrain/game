package ollama

import (
	"encoding/json"
	"fmt"
	"strings"

	"solo-leveling/internal/models"
)

func ParseSuggestionsJSON(raw string) ([]models.AISuggestion, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	// Try direct array first.
	var arr []models.AISuggestion
	if err := json.Unmarshal([]byte(clean), &arr); err == nil {
		return arr, nil
	}

	// Try object wrapper: {"suggestions":[...]}
	var wrapper struct {
		Suggestions []models.AISuggestion `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(clean), &wrapper); err == nil && len(wrapper.Suggestions) > 0 {
		return wrapper.Suggestions, nil
	}

	// Extract first JSON array substring as a fallback.
	start := strings.Index(clean, "[")
	end := strings.LastIndex(clean, "]")
	if start >= 0 && end > start {
		chunk := clean[start : end+1]
		if err := json.Unmarshal([]byte(chunk), &arr); err == nil {
			return arr, nil
		}
	}

	return nil, fmt.Errorf("invalid suggestions json")
}
