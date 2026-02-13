package tabs

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"solo-leveling/internal/models"
)

type importExpedition struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Deadline     string          `json:"deadline"`
	RewardEXP    int             `json:"reward_exp"`
	RewardStats  map[string]int  `json:"reward_stats"`
	IsRepeatable bool            `json:"is_repeatable"`
	Tasks        []importExpTask `json:"tasks"`
}

type importExpTask struct {
	Name            string `json:"name"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Desc            string `json:"desc"`
	IsCompleted     bool   `json:"is_completed"`
	ProgressCurrent int    `json:"progress_current"`
	ProgressTarget  int    `json:"progress_target"`
	RewardEXP       int    `json:"reward_exp"`
	TargetStat      string `json:"target_stat"`
	Stat            string `json:"stat"`
}

func showImportExpeditionsJSONDialog(ctx *Context) {
	open := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ctx.Window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(fmt.Errorf("не удалось прочитать файл: %w", err), ctx.Window)
			return
		}

		expeditions, err := parseImportedExpeditions(body)
		if err != nil {
			dialog.ShowError(err, ctx.Window)
			return
		}

		created := 0
		var errors []string
		for i, item := range expeditions {
			model, convErr := item.toModel()
			if convErr != nil {
				errors = append(errors, fmt.Sprintf("#%d: %s", i+1, convErr.Error()))
				continue
			}
			if err := ctx.Engine.CreateExpedition(model); err != nil {
				errors = append(errors, fmt.Sprintf("#%d (%s): %s", i+1, model.Name, err.Error()))
				continue
			}
			created++
		}

		msg := fmt.Sprintf("Создано экспедиций: %d", created)
		if len(errors) > 0 {
			msg += fmt.Sprintf("\n\nОшибки (%d):\n%s", len(errors), strings.Join(errors, "\n"))
		}
		dialog.ShowInformation("Импорт экспедиций", msg, ctx.Window)

		RefreshExpeditions(ctx)
		if ctx.RefreshQuests != nil {
			ctx.RefreshQuests()
		}
	}, ctx.Window)
	open.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	open.Show()
}

func parseImportedExpeditions(raw []byte) ([]importExpedition, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, fmt.Errorf("JSON пуст")
	}

	var list []importExpedition
	if err := json.Unmarshal([]byte(trimmed), &list); err == nil {
		if len(list) == 0 {
			return nil, fmt.Errorf("массив экспедиций пуст")
		}
		return list, nil
	}

	var wrapped struct {
		Expeditions []importExpedition `json:"expeditions"`
	}
	if err := json.Unmarshal([]byte(trimmed), &wrapped); err == nil && len(wrapped.Expeditions) > 0 {
		return wrapped.Expeditions, nil
	}

	var single importExpedition
	if err := json.Unmarshal([]byte(trimmed), &single); err == nil {
		if strings.TrimSpace(single.Name) == "" && len(single.Tasks) == 0 {
			return nil, fmt.Errorf("объект экспедиции не содержит name/tasks")
		}
		return []importExpedition{single}, nil
	}

	return nil, fmt.Errorf("невалидный JSON: ожидается объект экспедиции, массив или {\"expeditions\":[...]} ")
}

func (src importExpedition) toModel() (*models.Expedition, error) {
	name := strings.TrimSpace(src.Name)
	if name == "" {
		return nil, fmt.Errorf("expedition.name обязателен")
	}
	if len(src.Tasks) == 0 {
		return nil, fmt.Errorf("expedition %q: требуется минимум 1 task", name)
	}

	var deadline *time.Time
	if strings.TrimSpace(src.Deadline) != "" {
		parsed, err := parseExpeditionDeadline(src.Deadline)
		if err != nil {
			return nil, fmt.Errorf("expedition %q: неверный deadline (%v)", name, err)
		}
		deadline = &parsed
	}

	rewardStats := make(map[models.StatType]int)
	for key, value := range src.RewardStats {
		if value == 0 {
			continue
		}
		if stat, ok := parseStatTypeAlias(key); ok {
			rewardStats[stat] += value
		}
	}

	tasks := make([]models.ExpeditionTask, 0, len(src.Tasks))
	for i, task := range src.Tasks {
		parsedTask, err := task.toModel(i)
		if err != nil {
			return nil, fmt.Errorf("expedition %q: %w", name, err)
		}
		tasks = append(tasks, parsedTask)
	}

	rewardExp := src.RewardEXP
	if rewardExp < 0 {
		rewardExp = 0
	}

	return &models.Expedition{
		Name:         name,
		Description:  strings.TrimSpace(src.Description),
		Deadline:     deadline,
		RewardEXP:    rewardExp,
		RewardStats:  rewardStats,
		IsRepeatable: src.IsRepeatable,
		Status:       models.ExpeditionActive,
		Tasks:        tasks,
	}, nil
}

func (src importExpTask) toModel(index int) (models.ExpeditionTask, error) {
	title := strings.TrimSpace(src.Title)
	if title == "" {
		title = strings.TrimSpace(src.Name)
	}
	if title == "" {
		return models.ExpeditionTask{}, fmt.Errorf("task #%d: title обязателен", index+1)
	}

	target := src.ProgressTarget
	if target <= 0 {
		target = 1
	}
	current := src.ProgressCurrent
	if current < 0 {
		current = 0
	}
	if current > target {
		current = target
	}

	completed := src.IsCompleted || current >= target
	if completed {
		current = target
	}

	rewardExp := src.RewardEXP
	if rewardExp <= 0 {
		rewardExp = 20
	}

	statRaw := strings.TrimSpace(src.TargetStat)
	if statRaw == "" {
		statRaw = strings.TrimSpace(src.Stat)
	}
	stat, ok := parseStatTypeAlias(statRaw)
	if !ok {
		stat = models.StatStrength
	}

	description := strings.TrimSpace(src.Description)
	if description == "" {
		description = strings.TrimSpace(src.Desc)
	}

	return models.ExpeditionTask{
		Title:           title,
		Description:     description,
		IsCompleted:     completed,
		ProgressCurrent: current,
		ProgressTarget:  target,
		RewardEXP:       rewardExp,
		TargetStat:      stat,
	}, nil
}

func parseStatTypeAlias(raw string) (models.StatType, bool) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "STR", "STRENGTH", "СИЛА":
		return models.StatStrength, true
	case "AGI", "AGILITY", "ЛОВКОСТЬ":
		return models.StatAgility, true
	case "INT", "INTELLECT", "ИНТЕЛЛЕКТ":
		return models.StatIntellect, true
	case "STA", "ENDURANCE", "ВЫНОСЛИВОСТЬ":
		return models.StatEndurance, true
	default:
		return "", false
	}
}

func parseExpeditionDeadline(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			if layout == "2006-01-02" {
				return time.Date(ts.Year(), ts.Month(), ts.Day(), 23, 59, 59, 0, time.Local), nil
			}
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("ожидался RFC3339 или YYYY-MM-DD")
}
