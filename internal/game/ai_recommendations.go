package game

import (
	"fmt"
	"sort"
	"strings"

	"solo-leveling/internal/ai/suggestions"
	"solo-leveling/internal/models"
)

type AIQuestOption struct {
	Title       string
	Description string
	Minutes     int
	Exp         int
	Stat        models.StatType
	Rank        models.QuestRank
	Attempts    int
}

func EXPFromSuggestion(s models.AISuggestion) int {
	return models.CalculateQuestEXP(s.Minutes, s.Effort, s.Friction)
}

func RankFromSuggestion(s models.AISuggestion) models.QuestRank {
	return models.RankFromEXP(EXPFromSuggestion(s))
}

func (e *Engine) GetAIProfileText() (string, error) {
	return e.DB.GetAIProfileText()
}

func (e *Engine) SaveAIProfileText(text string) error {
	return e.DB.SaveAIProfileText(strings.TrimSpace(text))
}

func (e *Engine) ImportAIQuestOptions(rawJSON string) ([]AIQuestOption, error) {
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" {
		return nil, fmt.Errorf("json пуст")
	}

	profileText, err := e.GetAIProfileText()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(profileText) == "" {
		return nil, fmt.Errorf("профиль AI пуст")
	}

	rawSuggestions, err := suggestions.ParseJSON(rawJSON)
	if err != nil {
		_ = e.DB.LogAISuggestions(rawJSON, "manual_import", err.Error())
		return nil, fmt.Errorf("не удалось распарсить JSON: %w", err)
	}

	statLevels, err := e.GetStatLevels()
	if err != nil {
		return nil, err
	}
	playerStats := toPlayerStats(statLevels)
	weakOrder := weakStatOrder(playerStats)

	options := mapSuggestionsToOptions(rawSuggestions)
	sortOptions(options, weakOrder)
	if len(options) > 5 {
		options = options[:5]
	}

	_ = e.DB.LogAISuggestions(rawJSON, "manual_import", "")

	if len(options) == 0 {
		return nil, fmt.Errorf("json не содержит пригодных предложений")
	}

	return options, nil
}

func toPlayerStats(levels []models.StatLevel) models.PlayerStats {
	out := models.PlayerStats{STR: 1, AGI: 1, INT: 1, STA: 1}
	for _, s := range levels {
		switch s.StatType {
		case models.StatStrength:
			out.STR = s.Level
		case models.StatAgility:
			out.AGI = s.Level
		case models.StatIntellect:
			out.INT = s.Level
		case models.StatEndurance:
			out.STA = s.Level
		}
	}
	return out
}

func mapSuggestionsToOptions(in []models.AISuggestion) []AIQuestOption {
	out := make([]AIQuestOption, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, s := range in {
		stat, ok := parseAISuggestionStat(s.Stat)
		if !ok {
			continue
		}
		title := strings.TrimSpace(s.Title)
		desc := oneLine(s.Desc)
		if title == "" || desc == "" {
			continue
		}
		key := strings.ToLower(title)
		if seen[key] {
			continue
		}
		seen[key] = true

		exp := EXPFromSuggestion(s)
		rank := models.RankFromEXP(exp)
		out = append(out, AIQuestOption{
			Title:       title,
			Description: desc,
			Minutes:     s.Minutes,
			Exp:         exp,
			Stat:        stat,
			Rank:        rank,
			Attempts:    models.AttemptsForQuestEXP(exp),
		})
	}
	return out
}

func oneLine(s string) string {
	line := strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	line = strings.Join(strings.Fields(line), " ")
	if len(line) > 140 {
		return line[:137] + "..."
	}
	return line
}

func parseAISuggestionStat(raw string) (models.StatType, bool) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "STR":
		return models.StatStrength, true
	case "AGI":
		return models.StatAgility, true
	case "INT":
		return models.StatIntellect, true
	case "STA":
		return models.StatEndurance, true
	default:
		return "", false
	}
}

func weakStatOrder(stats models.PlayerStats) map[models.StatType]int {
	type pair struct {
		stat  models.StatType
		level int
	}
	all := []pair{
		{stat: models.StatStrength, level: stats.STR},
		{stat: models.StatAgility, level: stats.AGI},
		{stat: models.StatIntellect, level: stats.INT},
		{stat: models.StatEndurance, level: stats.STA},
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].level != all[j].level {
			return all[i].level < all[j].level
		}
		return all[i].stat < all[j].stat
	})
	order := make(map[models.StatType]int, len(all))
	for i, p := range all {
		order[p.stat] = i
	}
	return order
}

func sortOptions(options []AIQuestOption, weakOrder map[models.StatType]int) {
	sort.Slice(options, func(i, j int) bool {
		pi := weakOrder[options[i].Stat]
		pj := weakOrder[options[j].Stat]
		if pi != pj {
			return pi < pj
		}
		if options[i].Minutes != options[j].Minutes {
			return options[i].Minutes < options[j].Minutes
		}
		if rankWeight(options[i].Rank) != rankWeight(options[j].Rank) {
			return rankWeight(options[i].Rank) < rankWeight(options[j].Rank)
		}
		return options[i].Title < options[j].Title
	})
}
