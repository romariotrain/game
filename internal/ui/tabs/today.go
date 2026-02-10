package tabs

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/game"
	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildToday(ctx *Context) fyne.CanvasObject {
	ctx.CharacterPanel = container.NewVBox()
	RefreshToday(ctx)
	return container.NewVScroll(container.NewPadded(ctx.CharacterPanel))
}

func RefreshToday(ctx *Context) {
	if ctx.CharacterPanel == nil {
		return
	}
	ctx.CharacterPanel.Objects = nil

	stats, err := ctx.Engine.GetStatLevels()
	if err != nil {
		ctx.CharacterPanel.Add(components.MakeLabel("Ошибка загрузки: "+err.Error(), components.ColorRed))
		return
	}

	overallLevel, _ := ctx.Engine.GetOverallLevel()
	rankTitle := game.HunterRank(overallLevel)

	completedDungeons, _ := ctx.Engine.DB.GetCompletedDungeons(ctx.Engine.Character.ID)

	charCard := buildCharacterCard(ctx, overallLevel, rankTitle, stats, completedDungeons)
	ctx.CharacterPanel.Add(charCard)

	// Streak + Attempts in one row
	streakAttemptsCard := buildStreakAttemptsCard(ctx)
	ctx.CharacterPanel.Add(streakAttemptsCard)

	// Today's active quests
	ctx.CharacterPanel.Add(widget.NewSeparator())
	ctx.CharacterPanel.Add(components.MakeSectionHeader("Задания на сегодня"))
	buildTodayQuests(ctx)

	// Stats summary
	ctx.CharacterPanel.Add(widget.NewSeparator())
	ctx.CharacterPanel.Add(components.MakeSectionHeader("Характеристики"))
	for _, stat := range stats {
		row := components.MakeStatRow(stat)
		card := components.MakeCard(row)
		ctx.CharacterPanel.Add(card)
	}

	ctx.CharacterPanel.Refresh()
}

func buildCharacterCard(ctx *Context, level int, rank string, stats []models.StatLevel, completedDungeons []models.CompletedDungeon) *fyne.Container {
	nameText := components.MakeTitle(ctx.Engine.Character.Name, components.ColorGold, 24)
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		entry := widget.NewEntry()
		entry.SetText(ctx.Engine.Character.Name)
		dialog.ShowForm("Имя Охотника", "Сохранить", "Отмена",
			[]*widget.FormItem{widget.NewFormItem("Имя", entry)},
			func(ok bool) {
				if ok && strings.TrimSpace(entry.Text) != "" {
					ctx.Engine.RenameCharacter(strings.TrimSpace(entry.Text))
					RefreshToday(ctx)
				}
			}, ctx.Window)
	})

	nameRow := container.NewHBox(nameText, editBtn)

	rankColor := components.ParseHexColor(game.HunterRankColor(level))
	rankText := canvas.NewText(rank, rankColor)
	rankText.TextSize = 16
	rankText.TextStyle = fyne.TextStyle{Bold: true}

	levelText := components.MakeTitle(fmt.Sprintf("Общий уровень: %d", level), components.ColorText, 16)

	var statSummary []fyne.CanvasObject
	for _, s := range stats {
		txt := components.MakeLabel(fmt.Sprintf("%s %s: %d", s.StatType.Icon(), s.StatType.DisplayName(), s.Level), components.ColorTextDim)
		statSummary = append(statSummary, txt)
	}

	top := container.NewVBox(nameRow, rankText, levelText)
	statsRow := container.NewHBox(statSummary...)

	contentItems := []fyne.CanvasObject{top, widget.NewSeparator(), statsRow}

	// Show titles: dungeon titles + battle titles + streak titles
	var allTitles []string

	for _, cd := range completedDungeons {
		if cd.EarnedTitle != "" {
			allTitles = append(allTitles, cd.EarnedTitle)
		}
	}

	battleRewards, _ := ctx.Engine.DB.GetAllBattleRewards(ctx.Engine.Character.ID)
	for _, r := range battleRewards {
		if r.Title != "" {
			allTitles = append(allTitles, r.Title)
		}
	}

	streakTitles, _ := ctx.Engine.DB.GetStreakTitles(ctx.Engine.Character.ID)
	allTitles = append(allTitles, streakTitles...)

	if len(allTitles) > 0 {
		var titleWidgets []fyne.CanvasObject
		titleWidgets = append(titleWidgets, components.MakeTitle("Титулы:", components.ColorGold, 13))
		for _, t := range allTitles {
			titleWidgets = append(titleWidgets, components.MakeLabel("  "+t, components.ColorPurple))
		}
		contentItems = append(contentItems, widget.NewSeparator())
		contentItems = append(contentItems, container.NewVBox(titleWidgets...))
	}

	content := container.NewVBox(contentItems...)
	return components.MakeCard(content)
}

func buildStreakAttemptsCard(ctx *Context) *fyne.Container {
	streak, _ := ctx.Engine.DB.GetStreak(ctx.Engine.Character.ID)
	attempts := ctx.Engine.GetAttempts()

	// Streak section
	streakText := components.MakeTitle(fmt.Sprintf("🔥 Streak: %d дней", streak), components.ColorAccentBright, 16)

	var milestoneText fyne.CanvasObject
	title := models.StreakTitle(streak)
	if title != "" {
		milestoneText = components.MakeLabel(title, components.ColorGold)
	} else {
		nextMilestone := 7
		for _, m := range models.AllStreakMilestones() {
			if streak < m.Days {
				nextMilestone = m.Days
				break
			}
		}
		milestoneText = components.MakeLabel(
			fmt.Sprintf("До достижения: %d дн.", nextMilestone-streak),
			components.ColorTextDim,
		)
	}

	// Attempts section
	attemptsLabel := components.MakeTitle(
		fmt.Sprintf("⚔️ Попытки: %d/%d", attempts, models.MaxAttempts),
		components.ColorText, 16,
	)
	attemptsBar := components.MakeEXPBar(attempts, models.MaxAttempts, components.ColorAccent)

	left := container.NewVBox(streakText, milestoneText)
	right := container.NewVBox(attemptsLabel, attemptsBar)
	row := container.NewGridWithColumns(2, left, right)
	return components.MakeCard(row)
}

func buildTodayQuests(ctx *Context) {
	quests, err := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	if err != nil {
		ctx.CharacterPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		return
	}

	if len(quests) == 0 {
		ctx.CharacterPanel.Add(components.MakeEmptyState("Нет активных заданий. Создайте новое!"))
		return
	}

	for _, quest := range quests {
		q := quest
		rankBadge := components.MakeRankBadge(q.Rank)
		titleText := components.MakeTitle(q.Title, components.ColorText, 14)

		var typeIndicator fyne.CanvasObject
		if q.IsDaily {
			lbl := components.MakeLabel("Ежедневное", components.ColorBlue)
			lbl.TextSize = 11
			typeIndicator = lbl
		} else if q.DungeonID != nil {
			lbl := components.MakeLabel("Данж", components.ColorPurple)
			lbl.TextSize = 11
			typeIndicator = lbl
		} else {
			typeIndicator = layout.NewSpacer()
		}

		statText := components.MakeLabel(
			fmt.Sprintf("+%d EXP -> %s | Ранг: %s | +%d попыток",
				q.Exp, q.TargetStat.DisplayName(), q.Rank, models.AttemptsForQuestEXP(q.Exp)),
			components.ColorTextDim,
		)

		completeBtn := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
			result, err := ctx.Engine.CompleteQuest(q.ID)
			if err != nil {
				dialog.ShowError(err, ctx.Window)
				return
			}

			msg := fmt.Sprintf("+%d EXP к %s %s\n+%d попыток боя (всего: %d)",
				result.EXPAwarded, result.StatType.Icon(), result.StatType.DisplayName(),
				result.AttemptsAwarded, result.TotalAttempts)
			if result.LeveledUp {
				msg += fmt.Sprintf("\n\nУРОВЕНЬ ПОВЫШЕН! %s: %d -> %d",
					result.StatType.DisplayName(), result.OldLevel, result.NewLevel)
				for lvl := result.OldLevel + 1; lvl <= result.NewLevel; lvl++ {
					options := game.GetSkillOptions(result.StatType, lvl)
					if len(options) > 0 {
						msg += fmt.Sprintf("\nНовый скил на уровне %d!", lvl)
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
			}
		})
		completeBtn.Importance = widget.HighImportance

		topRow := container.NewHBox(rankBadge, titleText, typeIndicator, layout.NewSpacer(), completeBtn)
		card := components.MakeCard(container.NewVBox(topRow, statText))
		ctx.CharacterPanel.Add(card)
	}
}

func showSkillUnlockDialog(ctx *Context, stat models.StatType, level int) {
	options := game.GetSkillOptions(stat, level)
	if len(options) == 0 {
		return
	}

	opt := options[0]
	msg := fmt.Sprintf("Новый скил для %s на уровне %d!\n\n%s\n%s\nМультипликатор: x%.2f",
		stat.DisplayName(), level, opt.Name, opt.Description, opt.Multiplier)

	dialog.ShowConfirm("Скил разблокирован!", msg, func(ok bool) {
		if ok {
			ctx.Engine.UnlockSkill(stat, level, 0)
		}
	}, ctx.Window)
}
