package tabs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/game"
	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildQuests(ctx *Context) fyne.CanvasObject {
	ctx.QuestsPanel = container.NewVBox()
	addBtn := components.MakeStyledButton("Новое задание", theme.ContentAddIcon(), func() {
		showCreateQuestDialog(ctx)
	})
	importBtn := widget.NewButtonWithIcon("Импорт JSON", theme.FolderOpenIcon(), func() {
		showImportJSONDialog(ctx)
	})
	topBar := container.NewHBox(components.MakeSectionHeader("Активные Задания"), layout.NewSpacer(), importBtn, addBtn)
	RefreshQuests(ctx)
	return container.NewVScroll(container.NewPadded(container.NewVBox(topBar, ctx.QuestsPanel)))
}

func RefreshQuests(ctx *Context) {
	if ctx.QuestsPanel == nil {
		return
	}
	ctx.QuestsPanel.Objects = nil

	quests, err := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	if err != nil {
		ctx.QuestsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		ctx.QuestsPanel.Refresh()
		return
	}

	if len(quests) == 0 {
		ctx.QuestsPanel.Add(components.MakeEmptyState("Нет активных заданий. Создайте новое!"))
		ctx.QuestsPanel.Refresh()
		return
	}

	for _, quest := range quests {
		card := buildQuestCard(ctx, quest)
		ctx.QuestsPanel.Add(card)
	}

	ctx.QuestsPanel.Refresh()
}

func buildQuestCard(ctx *Context, q models.Quest) *fyne.Container {
	rankBadge := components.MakeRankBadge(q.Rank)
	titleText := components.MakeTitle(q.Title, components.ColorText, 15)

	var dailyIndicator fyne.CanvasObject
	if q.IsDaily {
		dailyLabel := components.MakeLabel("Ежедневное", components.ColorBlue)
		dailyLabel.TextSize = 11
		dailyIndicator = dailyLabel
	} else if q.DungeonID != nil {
		dungeonLabel := components.MakeLabel("Данж", components.ColorPurple)
		dungeonLabel.TextSize = 11
		dailyIndicator = dungeonLabel
	} else {
		dailyIndicator = layout.NewSpacer()
	}

	statText := components.MakeLabel(
		fmt.Sprintf("Цель: %s %s", q.TargetStat.Icon(), q.TargetStat.DisplayName()),
		components.ColorTextDim,
	)
	rewardText := components.MakeLabel(
		fmt.Sprintf("Награда: +%d EXP | Ранг: %s | +%d попыток", q.Exp, q.Rank, models.AttemptsForQuestEXP(q.Exp)),
		components.ColorAccentBright,
	)

	var descLabel fyne.CanvasObject
	if q.Description != "" {
		descLabel = components.MakeLabel(q.Description, components.ColorTextDim)
	} else {
		descLabel = layout.NewSpacer()
	}

	completeBtn := widget.NewButtonWithIcon("Выполнить", theme.ConfirmIcon(), func() {
		completeQuest(ctx, q)
	})
	completeBtn.Importance = widget.HighImportance

	failBtn := widget.NewButtonWithIcon("Провал", theme.CancelIcon(), func() {
		dialog.ShowConfirm("Провалить задание?",
			fmt.Sprintf("Провалить \"%s\"? EXP не будет начислен.", q.Title),
			func(ok bool) {
				if ok {
					ctx.Engine.FailQuest(q.ID)
					RefreshQuests(ctx)
					if ctx.RefreshStats != nil {
						ctx.RefreshStats()
					}
				}
			}, ctx.Window)
	})

	deleteBtn := widget.NewButtonWithIcon("Удалить", theme.DeleteIcon(), func() {
		msg := fmt.Sprintf("Удалить \"%s\"?", q.Title)
		if q.IsDaily {
			msg += "\nЕжедневный шаблон тоже будет деактивирован."
		}
		dialog.ShowConfirm("Удалить задание?", msg, func(ok bool) {
			if ok {
				ctx.Engine.DeleteQuest(q.ID)
				RefreshQuests(ctx)
			}
		}, ctx.Window)
	})

	topRow := container.NewHBox(rankBadge, titleText, dailyIndicator, layout.NewSpacer(), completeBtn, failBtn, deleteBtn)
	content := container.NewVBox(topRow, statText, rewardText, descLabel)
	return components.MakeCard(content)
}

func completeQuest(ctx *Context, q models.Quest) {
	result, err := ctx.Engine.CompleteQuest(q.ID)
	if err != nil {
		dialog.ShowError(err, ctx.Window)
		return
	}

	msg := fmt.Sprintf("Задание выполнено!\n\n+%d EXP к %s %s\n+%d попыток боя (всего: %d)",
		result.EXPAwarded, result.StatType.Icon(), result.StatType.DisplayName(),
		result.AttemptsAwarded, result.TotalAttempts)
	if result.LeveledUp {
		msg += fmt.Sprintf("\n\nУРОВЕНЬ ПОВЫШЕН! %s: %d -> %d",
			result.StatType.DisplayName(), result.OldLevel, result.NewLevel)
		for lvl := result.OldLevel + 1; lvl <= result.NewLevel; lvl++ {
			options := game.GetSkillOptions(result.StatType, lvl)
			if len(options) > 0 {
				msg += fmt.Sprintf("\n\nНовый скил доступен на уровне %d!", lvl)
				showSkillUnlockDialog(ctx, result.StatType, lvl)
			}
		}
	}

	if q.DungeonID != nil {
		done, err := ctx.Engine.CheckDungeonCompletion(*q.DungeonID)
		if err == nil && done {
			if err := ctx.Engine.CompleteDungeon(*q.DungeonID); err == nil {
				msg += "\n\nДАНЖ ПРОЙДЕН! Получена награда!"
			}
		}
	}
	if text := strings.TrimSpace(q.Congratulations); text != "" {
		msg += "\n\n" + text
	}

	dialog.ShowInformation("Задание выполнено!", msg, ctx.Window)

	if ctx.RefreshAll != nil {
		ctx.RefreshAll()
		return
	}
	RefreshQuests(ctx)
	RefreshDungeons(ctx)
}

func parseIntWithDefault(raw string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}
	return v
}

func showCreateQuestDialog(ctx *Context) {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название задания...")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Описание (необязательно)...")
	descEntry.SetMinRowsVisible(2)

	statNames := []string{"Сила", "Ловкость", "Интеллект", "Выносливость"}
	statSelect := widget.NewSelect(statNames, nil)
	statSelect.SetSelected("Сила")

	dailyCheck := widget.NewCheck("Ежедневное задание", nil)

	minutesEntry := widget.NewEntry()
	minutesEntry.SetText("20")
	minutesEntry.SetPlaceHolder("Минуты (например 20)")

	effortSelect := widget.NewSelect([]string{"1", "2", "3", "4", "5"}, nil)
	effortSelect.SetSelected("2")

	frictionSelect := widget.NewSelect([]string{"1", "2", "3"}, nil)
	frictionSelect.SetSelected("1")

	expLabel := components.MakeLabel("", components.ColorAccentBright)
	updatePreview := func() {
		minutes := parseIntWithDefault(minutesEntry.Text, 20)
		effort := parseIntWithDefault(effortSelect.Selected, 2)
		friction := parseIntWithDefault(frictionSelect.Selected, 1)
		exp := models.CalculateQuestEXP(minutes, effort, friction)
		expLabel.Text = fmt.Sprintf("EXP: +%d | Ранг: %s | Попытки: +%d", exp, models.RankFromEXP(exp), models.AttemptsForQuestEXP(exp))
		expLabel.Refresh()
	}
	minutesEntry.OnChanged = func(string) { updatePreview() }
	effortSelect.OnChanged = func(string) { updatePreview() }
	frictionSelect.OnChanged = func(string) { updatePreview() }
	updatePreview()

	formItems := []*widget.FormItem{
		widget.NewFormItem("Задание", titleEntry),
		widget.NewFormItem("Описание", descEntry),
		widget.NewFormItem("Минуты", minutesEntry),
		widget.NewFormItem("Effort (1-5)", effortSelect),
		widget.NewFormItem("Friction (1-3)", frictionSelect),
		widget.NewFormItem("Награда", expLabel),
		widget.NewFormItem("Стат", statSelect),
		widget.NewFormItem("Тип", dailyCheck),
	}

	dialog.ShowForm("Новое Задание", "Создать", "Отмена", formItems, func(ok bool) {
		if !ok || strings.TrimSpace(titleEntry.Text) == "" {
			return
		}

		minutes := parseIntWithDefault(minutesEntry.Text, 20)
		effort := parseIntWithDefault(effortSelect.Selected, 2)
		friction := parseIntWithDefault(frictionSelect.Selected, 1)
		exp := models.CalculateQuestEXP(minutes, effort, friction)
		statMap := map[string]models.StatType{
			"Сила": models.StatStrength, "Ловкость": models.StatAgility,
			"Интеллект": models.StatIntellect, "Выносливость": models.StatEndurance,
		}
		stat := statMap[statSelect.Selected]

		_, err := ctx.Engine.CreateQuest(
			strings.TrimSpace(titleEntry.Text),
			strings.TrimSpace(descEntry.Text),
			"",
			exp, stat, dailyCheck.Checked,
		)
		if err != nil {
			dialog.ShowError(err, ctx.Window)
			return
		}

		RefreshQuests(ctx)
	}, ctx.Window)
}

// showSkillUnlockDialog is defined in today.go

type importQuest struct {
	Name            string          `json:"name"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Desc            string          `json:"desc"`
	Congratulations string          `json:"congratulations"`
	Minutes         int             `json:"minutes"`
	Effort          int             `json:"effort"`
	Friction        int             `json:"friction"`
	Stat            string          `json:"stat"`
	Stats           json.RawMessage `json:"stats"` // legacy fallback
	IsDaily         bool            `json:"is_daily"`
}

func (q *importQuest) parseStat() models.StatType {
	statMap := map[string]models.StatType{
		"STR": models.StatStrength,
		"AGI": models.StatAgility,
		"INT": models.StatIntellect,
		"STA": models.StatEndurance,
	}

	if st, ok := statMap[strings.ToUpper(strings.TrimSpace(q.Stat))]; ok {
		return st
	}

	// Legacy: stats as string or array.
	var single string
	if json.Unmarshal(q.Stats, &single) == nil {
		if st, ok := statMap[strings.ToUpper(strings.TrimSpace(single))]; ok {
			return st
		}
	}
	var arr []string
	if json.Unmarshal(q.Stats, &arr) == nil {
		for _, item := range arr {
			if st, ok := statMap[strings.ToUpper(strings.TrimSpace(item))]; ok {
				return st
			}
		}
	}

	return models.StatIntellect
}

func (q *importQuest) parseTitle(stat models.StatType) string {
	name := strings.TrimSpace(q.Name)
	if name != "" {
		return name
	}
	title := strings.TrimSpace(q.Title)
	if title != "" {
		return title
	}
	minutes := q.Minutes
	if minutes <= 0 {
		minutes = 20
	}
	return fmt.Sprintf("Задача: %s (%d мин)", stat.DisplayName(), minutes)
}

func (q *importQuest) parseDescription() string {
	desc := strings.TrimSpace(q.Description)
	if desc != "" {
		return desc
	}
	return strings.TrimSpace(q.Desc)
}

func (q *importQuest) calculateEXP() int {
	return models.CalculateQuestEXP(q.Minutes, q.Effort, q.Friction)
}

func (q *importQuest) parseCongratulations() string {
	return strings.TrimSpace(q.Congratulations)
}

func showImportJSONDialog(ctx *Context) {
	entry := widget.NewMultiLineEntry()
	entry.SetMinRowsVisible(10)
	entry.SetPlaceHolder(`[{"title":"...","desc":"...","minutes":25,"effort":3,"friction":2,"stat":"INT","is_daily":false}]`)

	hint := components.MakeLabel("Вставьте JSON массив заданий", components.ColorTextDim)

	formItems := []*widget.FormItem{
		widget.NewFormItem("JSON", container.NewVBox(entry, hint)),
	}

	dialog.ShowForm("Импорт заданий из JSON", "Импортировать", "Отмена", formItems, func(ok bool) {
		if !ok {
			return
		}

		text := strings.TrimSpace(entry.Text)
		if text == "" {
			dialog.ShowError(fmt.Errorf("JSON пуст"), ctx.Window)
			return
		}

		var quests []importQuest
		if err := json.Unmarshal([]byte(text), &quests); err != nil {
			dialog.ShowError(fmt.Errorf("невалидный JSON: %w", err), ctx.Window)
			return
		}

		if len(quests) == 0 {
			dialog.ShowError(fmt.Errorf("массив заданий пуст"), ctx.Window)
			return
		}

		var created int
		var errors []string
		for i, q := range quests {
			stat := q.parseStat()
			title := q.parseTitle(stat)
			desc := q.parseDescription()
			congrats := q.parseCongratulations()
			exp := q.calculateEXP()

			_, err := ctx.Engine.CreateQuest(title, desc, congrats, exp, stat, q.IsDaily)
			if err != nil {
				errors = append(errors, fmt.Sprintf("#%d (%s → %s): %s", i+1, title, stat.DisplayName(), err.Error()))
				continue
			}
			created++
		}

		msg := fmt.Sprintf("Импортировано %d заданий", created)
		if len(errors) > 0 {
			msg += fmt.Sprintf("\n\nОшибки (%d):\n%s", len(errors), strings.Join(errors, "\n"))
		}
		dialog.ShowInformation("Импорт завершён", msg, ctx.Window)

		RefreshQuests(ctx)
		if ctx.RefreshAll != nil {
			ctx.RefreshAll()
		}
	}, ctx.Window)
}
