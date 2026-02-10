package game

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"solo-leveling/internal/models"
)

type suggestionTemplate struct {
	TitlePrefix string
	Description string
	Stat        models.StatType
	Rank        models.QuestRank
	Minutes     int
}

func (e *Engine) GetHunterProfile() (*models.HunterProfile, error) {
	return e.DB.GetHunterProfile(e.Character.ID)
}

func (e *Engine) SaveHunterProfile(p *models.HunterProfile) error {
	p.CharID = e.Character.ID
	return e.DB.SaveHunterProfile(p)
}

func (e *Engine) SuggestQuestOptions(limit int) ([]models.QuestSuggestion, error) {
	if limit < 3 {
		limit = 3
	}
	if limit > 5 {
		limit = 5
	}

	profile, err := e.GetHunterProfile()
	if err != nil {
		e.RecommendationSource = "error"
		e.RecommendationDetails = safeDetails(err.Error())
		return nil, err
	}
	if profile == nil {
		e.RecommendationSource = "error"
		e.RecommendationDetails = "профиль охотника не заполнен"
		return nil, fmt.Errorf("профиль охотника не заполнен")
	}

	stats, err := e.GetStatLevels()
	if err != nil {
		e.RecommendationSource = "error"
		e.RecommendationDetails = safeDetails(err.Error())
		return nil, err
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].Level < stats[j].Level })

	maxRank := maxAllowedRank(profile)
	maxMinutes := parseTimeBudget(profile.TimeBudget)

	activeTitles, err := e.getActiveQuestTitles()
	if err != nil {
		e.RecommendationSource = "error"
		e.RecommendationDetails = safeDetails(err.Error())
		return nil, err
	}

	// Prefer LLM suggestions when API key is configured.
	llmSuggestions, provider, err := e.suggestQuestOptionsLLM(limit, profile, stats, maxRank, maxMinutes, activeTitles)
	if err == nil && len(llmSuggestions) >= 3 {
		e.RecommendationSource = "llm"
		e.RecommendationDetails = "генерация через " + provider + " API"
		return llmSuggestions, nil
	}
	if err != nil {
		e.RecommendationSource = "rule-based"
		e.RecommendationDetails = "fallback: " + safeDetails(err.Error())
	}

	// Fallback must always work offline.
	suggestions := e.suggestQuestOptionsRuleBased(limit, profile, stats, maxRank, maxMinutes)
	suggestions = dedupeAgainstActive(suggestions, activeTitles, limit)
	if len(suggestions) == 0 {
		e.RecommendationSource = "error"
		e.RecommendationDetails = "не удалось сформировать рекомендации"
		return nil, fmt.Errorf("не удалось сформировать рекомендации")
	}
	if e.RecommendationSource == "" {
		e.RecommendationSource = "rule-based"
		e.RecommendationDetails = "локальная генерация"
	}
	return suggestions, nil
}

func safeDetails(raw string) string {
	msg := strings.TrimSpace(raw)
	if msg == "" {
		return ""
	}
	keyLike := regexp.MustCompile(`sk-[A-Za-z0-9_-]+`)
	return keyLike.ReplaceAllString(msg, "sk-***")
}

func (e *Engine) suggestQuestOptionsRuleBased(limit int, profile *models.HunterProfile, stats []models.StatLevel, maxRank models.QuestRank, maxMinutes int) []models.QuestSuggestion {
	var pool []suggestionTemplate
	pool = append(pool, baseSuggestionPool()...)
	pool = append(pool, goalSuggestionPool(profile.Goals)...)
	pool = filterPool(pool, maxRank, maxMinutes)

	var suggestions []models.QuestSuggestion
	used := make(map[string]bool)

	// 1) First pick for weakest stats.
	for _, s := range stats {
		picked, ok := pickByStat(pool, s.StatType, used)
		if !ok {
			continue
		}
		suggestions = append(suggestions, toSuggestion(picked))
		used[picked.TitlePrefix] = true
		if len(suggestions) >= limit {
			return suggestions
		}
	}

	// 2) Sometimes include a streak-support suggestion if today has no completions.
	needStreak, _ := needsStreakSupport(e)
	if needStreak {
		if picked, ok := pickByTitle(pool, "Мини-стабильность", used); ok {
			suggestions = append(suggestions, toSuggestion(picked))
			used[picked.TitlePrefix] = true
		}
	}
	if len(suggestions) >= limit {
		return suggestions[:limit]
	}

	// 3) Fill remaining slots with easiest useful options.
	sort.Slice(pool, func(i, j int) bool {
		if pool[i].Rank != pool[j].Rank {
			return rankWeight(pool[i].Rank) < rankWeight(pool[j].Rank)
		}
		return pool[i].Minutes < pool[j].Minutes
	})
	for _, item := range pool {
		if used[item.TitlePrefix] {
			continue
		}
		suggestions = append(suggestions, toSuggestion(item))
		used[item.TitlePrefix] = true
		if len(suggestions) >= limit {
			break
		}
	}
	return suggestions
}

func (e *Engine) getActiveQuestTitles() (map[string]bool, error) {
	active, err := e.DB.GetActiveQuests(e.Character.ID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(active))
	for _, q := range active {
		out[normalizeTitle(q.Title)] = true
	}
	return out, nil
}

func dedupeAgainstActive(suggestions []models.QuestSuggestion, activeTitles map[string]bool, limit int) []models.QuestSuggestion {
	out := make([]models.QuestSuggestion, 0, len(suggestions))
	used := make(map[string]bool, len(suggestions))
	for _, s := range suggestions {
		key := normalizeTitle(s.Title)
		if key == "" || activeTitles[key] || used[key] {
			continue
		}
		used[key] = true
		out = append(out, s)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func normalizeTitle(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func toSuggestion(t suggestionTemplate) models.QuestSuggestion {
	return models.QuestSuggestion{
		Title:            t.TitlePrefix,
		Description:      t.Description,
		Rank:             t.Rank,
		TargetStat:       t.Stat,
		EstimatedMinutes: t.Minutes,
	}
}

func parseTimeBudget(raw string) int {
	re := regexp.MustCompile(`\d+`)
	num := re.FindString(raw)
	if num == "" {
		return 45
	}
	var v int
	fmt.Sscanf(num, "%d", &v)
	if v < 10 {
		return 15
	}
	if v > 180 {
		return 180
	}
	return v
}

func maxAllowedRank(profile *models.HunterProfile) models.QuestRank {
	text := strings.ToLower(strings.Join([]string{
		profile.PsychologicalConstraints,
		profile.Dislikes,
		profile.SupportStyle,
	}, " "))

	if strings.Contains(text, "выгора") || strings.Contains(text, "устал") || strings.Contains(text, "сложн") {
		return models.RankD
	}
	if strings.Contains(text, "мягк") || strings.Contains(text, "нейтрал") {
		return models.RankC
	}
	return models.RankB
}

func rankWeight(rank models.QuestRank) int {
	switch rank {
	case models.RankE:
		return 1
	case models.RankD:
		return 2
	case models.RankC:
		return 3
	case models.RankB:
		return 4
	default:
		return 5
	}
}

func filterPool(pool []suggestionTemplate, maxRank models.QuestRank, maxMinutes int) []suggestionTemplate {
	out := make([]suggestionTemplate, 0, len(pool))
	for _, p := range pool {
		if rankWeight(p.Rank) > rankWeight(maxRank) {
			continue
		}
		if p.Minutes > maxMinutes {
			continue
		}
		out = append(out, p)
	}
	return out
}

func pickByStat(pool []suggestionTemplate, stat models.StatType, used map[string]bool) (suggestionTemplate, bool) {
	for _, p := range pool {
		if p.Stat == stat && !used[p.TitlePrefix] {
			return p, true
		}
	}
	return suggestionTemplate{}, false
}

func pickByTitle(pool []suggestionTemplate, title string, used map[string]bool) (suggestionTemplate, bool) {
	for _, p := range pool {
		if p.TitlePrefix == title && !used[p.TitlePrefix] {
			return p, true
		}
	}
	return suggestionTemplate{}, false
}

func needsStreakSupport(e *Engine) (bool, error) {
	activities, err := e.DB.GetDailyActivityLast30(e.Character.ID)
	if err != nil {
		return false, err
	}
	today := time.Now().Format("2006-01-02")
	for _, a := range activities {
		if a.Date == today {
			return a.QuestsComplete == 0, nil
		}
	}
	return true, nil
}

func baseSuggestionPool() []suggestionTemplate {
	return []suggestionTemplate{
		{TitlePrefix: "Мини-стабильность", Description: "Сделать одно маленькое полезное действие для поддержания ритма", Rank: models.RankE, Stat: models.StatEndurance, Minutes: 10},
		{TitlePrefix: "Чистый старт", Description: "Быстро разобрать 1 маленький участок рабочего пространства", Rank: models.RankE, Stat: models.StatStrength, Minutes: 12},
		{TitlePrefix: "План на 20 минут", Description: "Записать 3 приоритетные задачи на сегодня без перегруза", Rank: models.RankD, Stat: models.StatIntellect, Minutes: 15},
		{TitlePrefix: "Прогулка с фокусом", Description: "Короткая ходьба без отвлечений, чтобы перезагрузить внимание", Rank: models.RankD, Stat: models.StatAgility, Minutes: 20},
		{TitlePrefix: "Глубокий блок", Description: "Один сфокусированный блок работы по важной задаче", Rank: models.RankC, Stat: models.StatIntellect, Minutes: 30},
		{TitlePrefix: "Домашний рывок", Description: "Завершить одно отложенное бытовое дело до конца", Rank: models.RankC, Stat: models.StatStrength, Minutes: 25},
		{TitlePrefix: "Восстановление", Description: "Спокойная рутина для снижения усталости и восстановления ресурса", Rank: models.RankE, Stat: models.StatEndurance, Minutes: 15},
		{TitlePrefix: "Дисциплина дня", Description: "Выполнить заранее выбранный микро-ритуал в одно и то же время", Rank: models.RankD, Stat: models.StatAgility, Minutes: 10},
		{TitlePrefix: "Решающий шаг", Description: "Сделать один измеримый шаг к долгосрочной цели", Rank: models.RankB, Stat: models.StatStrength, Minutes: 40},
	}
}

func goalSuggestionPool(goals string) []suggestionTemplate {
	g := strings.ToLower(goals)
	var out []suggestionTemplate

	if strings.Contains(g, "здоров") {
		out = append(out, suggestionTemplate{
			TitlePrefix: "База здоровья",
			Description: "Короткое действие для тела: разминка, прогулка или мягкая активность",
			Stat:        models.StatEndurance,
			Rank:        models.RankD,
			Minutes:     20,
		})
	}
	if strings.Contains(g, "дисцип") || strings.Contains(g, "привыч") {
		out = append(out, suggestionTemplate{
			TitlePrefix: "Точка дисциплины",
			Description: "Выполнить заранее выбранный повторяемый шаг без усложнения",
			Stat:        models.StatAgility,
			Rank:        models.RankD,
			Minutes:     15,
		})
	}
	if strings.Contains(g, "деньг") || strings.Contains(g, "финанс") {
		out = append(out, suggestionTemplate{
			TitlePrefix: "Финансовый контроль",
			Description: "Проверить и зафиксировать один финансовый показатель или расход",
			Stat:        models.StatIntellect,
			Rank:        models.RankC,
			Minutes:     20,
		})
	}
	if strings.Contains(g, "навык") || strings.Contains(g, "обуч") {
		out = append(out, suggestionTemplate{
			TitlePrefix: "Учебный спринт",
			Description: "Короткий блок обучения с одним конкретным результатом",
			Stat:        models.StatIntellect,
			Rank:        models.RankC,
			Minutes:     30,
		})
	}
	if strings.Contains(g, "баланс") {
		out = append(out, suggestionTemplate{
			TitlePrefix: "Баланс дня",
			Description: "Сделать один шаг для личной жизни или отдыха без чувства вины",
			Stat:        models.StatEndurance,
			Rank:        models.RankE,
			Minutes:     15,
		})
	}

	return out
}
