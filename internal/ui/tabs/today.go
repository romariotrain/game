package tabs

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

	charCard := buildCharacterCard(ctx, overallLevel, rankTitle, stats)
	todayTile, attemptsTile, enemyTile := buildDailyProgressTiles(ctx)
	rightTopCard := buildDailyProgressCard(todayTile, attemptsTile, enemyTile)

	leftMin := canvas.NewRectangle(color.Transparent)
	leftMin.SetMinSize(fyne.NewSize(320, 250))
	leftPane := container.NewStack(leftMin, charCard)

	rightMin := canvas.NewRectangle(color.Transparent)
	rightMin.SetMinSize(fyne.NewSize(320, 250))
	rightPane := container.NewStack(rightMin, rightTopCard)
	topRow := container.NewGridWithColumns(2, leftPane, rightPane)
	ctx.CharacterPanel.Add(topRow)

	// Streak + Attempts in one row
	streakAttemptsCard := buildStreakAttemptsCard(ctx)
	ctx.CharacterPanel.Add(streakAttemptsCard)

	// Today's active quests
	ctx.CharacterPanel.Add(widget.NewSeparator())
	ctx.CharacterPanel.Add(components.MakeSectionHeader("Задания на сегодня"))
	buildTodayQuests(ctx)

	ctx.CharacterPanel.Refresh()
}

func buildCharacterCard(ctx *Context, level int, rank string, stats []models.StatLevel) *fyne.Container {
	nameText := components.MakeTitle(ctx.Engine.Character.Name, components.ColorGold, 20)
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

	nameRow := container.NewHBox(nameText, layout.NewSpacer(), editBtn)

	rankColor := components.ParseHexColor(game.HunterRankColor(level))
	rankText := canvas.NewText(rank, rankColor)
	rankText.TextSize = 14
	rankText.TextStyle = fyne.TextStyle{Bold: true}

	levelText := components.MakeLabel(fmt.Sprintf("Lv.%d", level), components.ColorTextDim)
	levelText.TextSize = 11
	totalEXP := 0
	for _, stat := range stats {
		totalEXP += stat.TotalEXP
	}
	expText := components.MakeLabel(fmt.Sprintf("EXP %d", totalEXP), components.ColorTextDim)
	expText.TextSize = 11
	statsAura := buildCompactStatBlock(stats)

	metaGap := canvas.NewRectangle(color.Transparent)
	metaGap.SetMinSize(fyne.NewSize(8, 0))
	metaRow := container.NewHBox(rankText, metaGap, levelText, metaGap, expText)

	rightItems := []fyne.CanvasObject{
		nameRow,
		metaRow,
		widget.NewSeparator(),
		statsAura,
	}

	avatarPane := buildHunterAvatarPane()
	rightPane := container.NewVBox(rightItems...)
	rightPaneWithInset := container.New(layout.NewCustomPaddedLayout(0, 0, 10, 0), rightPane)
	row := container.NewBorder(nil, nil, avatarPane, nil, rightPaneWithInset)
	return makeTopCard(row, fyne.NewSize(0, 250))
}

func buildHunterAvatarPane() fyne.CanvasObject {
	const avatarWidth float32 = 150
	const avatarHeight float32 = 190

	bg := canvas.NewRectangle(color.NRGBA{R: 24, G: 28, B: 44, A: 255})
	bg.CornerRadius = 8
	bg.SetMinSize(fyne.NewSize(avatarWidth, avatarHeight))

	avatarPath := resolveAvatarPath()
	if avatarPath != "" {
		avatar := canvas.NewImageFromFile(avatarPath)
		avatar.FillMode = canvas.ImageFillContain
		avatar.SetMinSize(fyne.NewSize(avatarWidth, avatarHeight))
		return container.NewStack(bg, avatar)
	}

	placeholder := canvas.NewImageFromResource(theme.AccountIcon())
	placeholder.FillMode = canvas.ImageFillContain
	placeholder.SetMinSize(fyne.NewSize(92, 92))
	return container.NewStack(bg, container.NewCenter(placeholder))
}

func buildDailyProgressTiles(ctx *Context) (fyne.CanvasObject, fyne.CanvasObject, fyne.CanvasObject) {
	enemyName := "Неизвестный враг"
	enemyRankText := "?"
	if ctx.Features.Combat {
		enemy, err := ctx.Engine.GetCurrentEnemy()
		if err == nil && enemy != nil {
			enemyName = enemy.Name
			enemyRankText = string(enemy.Rank)
		} else if err == nil && enemy == nil {
			enemyName = "Новый враг скоро"
		}
	}

	quests, _ := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	activities, _ := ctx.Engine.DB.GetDailyActivityLast30(ctx.Engine.Character.ID)
	today := time.Now().Format("2006-01-02")
	completedToday := 0
	failedToday := 0
	for _, a := range activities {
		if a.Date == today {
			completedToday = a.QuestsComplete
			failedToday = a.QuestsFailed
			break
		}
	}
	activeToday := len(quests)
	totalToday := completedToday + failedToday + activeToday
	if totalToday < 1 {
		totalToday = 1
	}

	errorLimit := 3
	attempts := ctx.Engine.GetAttempts()

	ctaLabel := "⚔️ Вступить в бой"
	ctaEnabled := ctx.Features.Combat && attempts > 0
	if failedToday > 0 && ctaEnabled {
		ctaLabel = "Продолжить бой"
	}
	if !ctaEnabled {
		ctaLabel = "Бой недоступен"
	}
	ctaBtn := widget.NewButtonWithIcon(ctaLabel, theme.MediaPlayIcon(), func() {
		dialog.ShowInformation("Бой", "Открой вкладку Tower для запуска боя.", ctx.Window)
	})
	ctaBtn.Importance = widget.HighImportance
	if !ctaEnabled {
		ctaBtn.Disable()
	}

	todayTitle := components.MakeLabel("Сегодня", components.ColorTextDim)
	todayTitle.TextSize = 12
	todayValue := components.MakeTitle(fmt.Sprintf("Выполнено %d/%d", completedToday, totalToday), components.ColorText, 15)
	todayErrors := components.MakeLabel(fmt.Sprintf("Ошибки %d/%d", failedToday, errorLimit), components.ColorTextDim)
	todayErrors.TextSize = 12
	todayTile := makeTopCard(
		container.NewVBox(
			todayTitle,
			todayValue,
			makeMiniProgressBar(completedToday, totalToday, components.ColorAccentBright),
			todayErrors,
		),
		fyne.NewSize(0, 100),
	)

	attemptsTitle := components.MakeLabel("Попытки", components.ColorTextDim)
	attemptsTitle.TextSize = 12
	attemptsValue := components.MakeTitle(fmt.Sprintf("%d/%d", attempts, models.MaxAttempts), components.ColorText, 18)
	attemptsState := components.MakeLabel("Ресурс боя", components.ColorTextDim)
	attemptsState.TextSize = 12
	attemptsTile := makeTopCard(
		container.NewVBox(
			attemptsTitle,
			attemptsValue,
			makeMiniProgressBar(attempts, models.MaxAttempts, components.ColorAccent),
			attemptsState,
		),
		fyne.NewSize(0, 100),
	)

	enemyIcon := canvas.NewImageFromResource(theme.VisibilityIcon())
	enemyIcon.FillMode = canvas.ImageFillContain
	enemyIcon.SetMinSize(fyne.NewSize(32, 32))
	enemyIconBg := canvas.NewRectangle(color.NRGBA{R: 22, G: 22, B: 38, A: 255})
	enemyIconBg.CornerRadius = 7
	enemyIconBg.SetMinSize(fyne.NewSize(44, 44))
	enemyIconBox := container.NewStack(enemyIconBg, container.NewCenter(enemyIcon))

	enemyTitle := components.MakeLabel("Враг дня", components.ColorTextDim)
	enemyTitle.TextSize = 12
	enemyNameLabel := components.MakeTitle(enemyName, components.ColorText, 15)
	enemyRankLabel := components.MakeLabel(fmt.Sprintf("Ранг %s", enemyRankText), components.ColorTextDim)
	enemyRankLabel.TextSize = 12
	enemyInfo := container.NewVBox(enemyTitle, enemyNameLabel, enemyRankLabel)

	enemyTile := makeTopCard(
		container.NewVBox(
			container.NewHBox(enemyIconBox, enemyInfo, layout.NewSpacer()),
			layout.NewSpacer(),
			ctaBtn,
		),
		fyne.NewSize(0, 138),
	)

	return todayTile, attemptsTile, enemyTile
}

func buildDailyProgressCard(todayTile, attemptsTile, enemyTile fyne.CanvasObject) *fyne.Container {
	topMetrics := container.NewGridWithColumns(2, todayTile, attemptsTile)
	bottomRow := container.NewGridWithColumns(1, enemyTile)
	return container.NewGridWithRows(2, topMetrics, bottomRow)
}

func buildCompactStatBlock(stats []models.StatLevel) fyne.CanvasObject {
	statByType := make(map[models.StatType]models.StatLevel)
	for _, s := range stats {
		statByType[s.StatType] = s
	}

	statOrder := []models.StatType{
		models.StatStrength,
		models.StatAgility,
		models.StatIntellect,
		models.StatEndurance,
	}

	var chips []fyne.CanvasObject
	for _, statType := range statOrder {
		stat := statByType[statType]
		if stat.Level < 1 {
			stat.Level = 1
		}
		expNeeded := models.ExpForLevel(stat.Level)
		if stat.CurrentEXP < 0 {
			stat.CurrentEXP = 0
		}
		if stat.CurrentEXP > expNeeded {
			stat.CurrentEXP = expNeeded
		}

		code := components.MakeLabel(fmt.Sprintf("%s %d", statShortCode(statType), stat.Level), components.ColorText)
		code.TextSize = 12
		code.TextStyle = fyne.TextStyle{Bold: true}
		expText := components.MakeLabel(fmt.Sprintf("EXP %d/%d", stat.CurrentEXP, expNeeded), components.ColorTextDim)
		expText.TextSize = 11
		chips = append(chips, makeStatChip(container.NewVBox(code, expText)))
	}

	return container.NewGridWithColumns(2, chips...)
}

func statShortCode(statType models.StatType) string {
	switch statType {
	case models.StatStrength:
		return "STR"
	case models.StatAgility:
		return "AGI"
	case models.StatIntellect:
		return "INT"
	case models.StatEndurance:
		return "STA"
	default:
		return "STAT"
	}
}

func makeTopCard(content fyne.CanvasObject, minSize fyne.Size) *fyne.Container {
	bg := canvas.NewRectangle(components.ColorBGCard)
	bg.CornerRadius = 8
	bg.SetMinSize(minSize)
	inset := container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), content)
	return container.NewStack(bg, inset)
}

func makeStatChip(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 22, G: 22, B: 38, A: 230})
	bg.CornerRadius = 6
	bg.SetMinSize(fyne.NewSize(0, 50))
	inset := container.New(layout.NewCustomPaddedLayout(8, 8, 8, 8), content)
	return container.NewStack(bg, inset)
}

func makeMiniProgressBar(current, max int, barColor color.Color) fyne.CanvasObject {
	if max <= 0 {
		max = 1
	}
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}

	ratio := float32(current) / float32(max)

	bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 50, A: 255})
	bg.CornerRadius = 3
	bg.SetMinSize(fyne.NewSize(140, 8))

	fill := canvas.NewRectangle(barColor)
	fill.CornerRadius = 3
	fill.SetMinSize(fyne.NewSize(140*ratio, 8))

	return container.NewStack(bg, container.NewHBox(fill, layout.NewSpacer()))
}

func resolveAvatarPath() string {
	candidates := []string{
		filepath.Join("assets", "avatar.png"),
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, "assets", "avatar.png"))
	}

	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
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

	var mainQuests []models.Quest
	var mediumQuests []models.Quest
	var quickQuests []models.Quest

	for _, q := range quests {
		switch classifyTodayQuest(q) {
		case "main":
			mainQuests = append(mainQuests, q)
		case "medium":
			mediumQuests = append(mediumQuests, q)
		default:
			quickQuests = append(quickQuests, q)
		}
	}

	sortTodayQuestSection(mainQuests)
	sortTodayQuestSection(mediumQuests)
	sortTodayQuestSection(quickQuests)

	accordion := widget.NewAccordion(
		widget.NewAccordionItem(
			fmt.Sprintf("Главные (%d)", len(mainQuests)),
			buildTodayQuestSection(ctx, mainQuests),
		),
		widget.NewAccordionItem(
			fmt.Sprintf("Средние (%d)", len(mediumQuests)),
			buildTodayQuestSection(ctx, mediumQuests),
		),
		widget.NewAccordionItem(
			fmt.Sprintf("Быстрые (%d)", len(quickQuests)),
			buildTodayQuestSection(ctx, quickQuests),
		),
	)
	accordion.Open(0)
	ctx.CharacterPanel.Add(accordion)
}

func buildTodayQuestSection(ctx *Context, quests []models.Quest) fyne.CanvasObject {
	if len(quests) == 0 {
		return components.MakeEmptyState("Нет задач в этой группе.")
	}
	section := container.NewVBox()
	for _, q := range quests {
		section.Add(buildTodayQuestCard(ctx, q))
	}
	return section
}

func buildTodayQuestCard(ctx *Context, q models.Quest) *fyne.Container {
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
	return components.MakeCard(container.NewVBox(topRow, statText))
}

func classifyTodayQuest(q models.Quest) string {
	if q.Rank == models.RankA || q.Rank == models.RankB || q.Exp > 30 {
		return "main"
	}
	if q.Rank == models.RankC || q.Rank == models.RankD || (q.Exp >= 15 && q.Exp <= 30) {
		return "medium"
	}
	return "quick"
}

func sortTodayQuestSection(quests []models.Quest) {
	sort.Slice(quests, func(i, j int) bool {
		leftRank := questRankSortWeight(quests[i].Rank)
		rightRank := questRankSortWeight(quests[j].Rank)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if quests[i].Exp != quests[j].Exp {
			return quests[i].Exp > quests[j].Exp
		}
		return strings.ToLower(quests[i].Title) < strings.ToLower(quests[j].Title)
	})
}

func questRankSortWeight(rank models.QuestRank) int {
	switch rank {
	case models.RankS:
		return 6
	case models.RankA:
		return 5
	case models.RankB:
		return 4
	case models.RankC:
		return 3
	case models.RankD:
		return 2
	default:
		return 1
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
