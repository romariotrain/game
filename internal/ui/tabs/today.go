package tabs

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
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
	// Use Stack (not VBox!) so the single child Border layout stretches to fill all space.
	ctx.CharacterPanel = container.NewStack()
	RefreshToday(ctx)
	return ctx.CharacterPanel
}

func RefreshToday(ctx *Context) {
	if ctx.CharacterPanel == nil {
		return
	}
	ctx.CharacterPanel.Objects = nil

	stats, err := ctx.Engine.GetStatLevels()
	if err != nil {
		ctx.CharacterPanel.Objects = []fyne.CanvasObject{
			components.MakeLabel("Ошибка загрузки: "+err.Error(), components.T().Danger),
		}
		ctx.CharacterPanel.Refresh()
		return
	}

	overallLevel, _ := ctx.Engine.GetOverallLevel()
	rankTitle := game.HunterRank(overallLevel)

	// --- Top zone: character card + enemy card side by side, fixed height ---
	charCard := buildCharacterCard(ctx, overallLevel, rankTitle, stats)
	enemyCard := buildEnemyDayCard(ctx)
	topRow := container.NewGridWithColumns(2, charCard, enemyCard)

	// --- Middle zone: compact streak line ---
	streakLine := buildStreakLine(ctx)

	// --- Top block = cards + streak ---
	topBlock := container.NewVBox(topRow, streakLine)

	// --- Bottom zone: quests fill all remaining space ---
	questsHeader := components.MakeSectionHeader("Задания на сегодня")
	questsContent := buildTodayQuestsWidget(ctx)
	questsScroll := container.NewVScroll(container.NewPadded(questsContent))

	questsBlock := container.NewBorder(questsHeader, nil, nil, nil, questsScroll)

	// --- Assemble: top fixed, quests stretch ---
	root := container.NewBorder(
		container.NewPadded(topBlock),    // top — fixed
		nil,                              // bottom
		nil,                              // left
		nil,                              // right
		container.NewPadded(questsBlock), // center — stretches
	)

	ctx.CharacterPanel.Objects = []fyne.CanvasObject{root}
	ctx.CharacterPanel.Refresh()
}

// =============================================================================
// Character Card
// =============================================================================

func buildCharacterCard(ctx *Context, level int, rank string, stats []models.StatLevel) *fyne.Container {
	t := components.T()

	// --- Portrait ---
	const portraitW float32 = 256
	const portraitH float32 = 316
	portraitBg := canvas.NewRectangle(t.BGPanel)
	portraitBg.CornerRadius = components.RadiusLG
	portraitBg.SetMinSize(fyne.NewSize(portraitW, portraitH))
	portraitBg.StrokeWidth = components.BorderThin
	portraitBg.StrokeColor = t.Border

	var portraitBox fyne.CanvasObject
	if avatarPath := resolveAvatarPath(); avatarPath != "" {
		avatar := canvas.NewImageFromFile(avatarPath)
		if trimmed := loadTrimmedAvatar(avatarPath); trimmed != nil {
			avatar = canvas.NewImageFromImage(trimmed)
		}
		avatar.FillMode = canvas.ImageFillContain
		avatar.SetMinSize(fyne.NewSize(portraitW, portraitH))
		portraitBox = container.NewStack(portraitBg, avatar)
	} else {
		placeholder := canvas.NewImageFromResource(theme.AccountIcon())
		placeholder.FillMode = canvas.ImageFillContain
		placeholder.SetMinSize(fyne.NewSize(182, 182))
		portraitBox = container.NewStack(portraitBg, container.NewCenter(placeholder))
	}

	// Rank + level + EXP under portrait
	rankColor := components.ParseHexColor(game.HunterRankColor(level))
	rankLabel := canvas.NewText(rank, rankColor)
	rankLabel.TextSize = components.TextHeadingMD
	rankLabel.TextStyle = fyne.TextStyle{Bold: true}

	levelLabel := canvas.NewText(fmt.Sprintf("Lv.%d", level), t.TextSecondary)
	levelLabel.TextSize = components.TextBodyMD

	totalEXP := 0
	for _, stat := range stats {
		totalEXP += stat.TotalEXP
	}
	expLabel := canvas.NewText(fmt.Sprintf("EXP %d", totalEXP), t.TextMuted)
	expLabel.TextSize = components.TextBodySM

	metaRow := container.NewHBox(layout.NewSpacer(), rankLabel, levelLabel, expLabel, layout.NewSpacer())

	// Active title selector
	var titleRow fyne.CanvasObject
	allTitles, _ := ctx.Engine.GetAllTitles()
	if len(allTitles) > 0 {
		sel := widget.NewSelect(allTitles, func(chosen string) {
			_ = ctx.Engine.SetActiveTitle(chosen)
		})
		sel.PlaceHolder = "Нет титула"
		if ctx.Engine.Character.ActiveTitle != "" {
			sel.SetSelected(ctx.Engine.Character.ActiveTitle)
		}
		titleRow = sel
	} else {
		placeholder := canvas.NewText("Нет титулов", t.TextMuted)
		placeholder.TextSize = components.TextBodySM
		placeholder.Alignment = fyne.TextAlignCenter
		titleRow = container.NewHBox(layout.NewSpacer(), placeholder, layout.NewSpacer())
	}

	leftCol := container.NewVBox(portraitBox, metaRow, titleRow)

	// --- Right column: stats with colored left bars ---
	statsBlock := buildStatBlockWithBars(stats)
	statsHeader := components.MakeSystemHeaderCompact("Характеристики")
	rightCol := container.NewVBox(statsHeader, statsBlock)
	rightColPadded := container.New(layout.NewCustomPaddedLayout(0, 0, 10, 0), rightCol)

	row := container.NewBorder(nil, nil, leftCol, nil, rightColPadded)
	return components.MakeHUDPanelSized(row, fyne.NewSize(0, 388))
}

// =============================================================================
// Stat Block with colored progress bars
// =============================================================================

func statBarColor(statType models.StatType) color.Color {
	t := components.T()
	switch statType {
	case models.StatStrength:
		return t.StatSTR
	case models.StatAgility:
		return t.StatAGI
	case models.StatIntellect:
		return t.StatINT
	case models.StatEndurance:
		return t.StatSTA
	default:
		return t.AccentDim
	}
}

func statIcon(statType models.StatType) string {
	switch statType {
	case models.StatStrength:
		return "⚔"
	case models.StatAgility:
		return "⚡"
	case models.StatIntellect:
		return "🧠"
	case models.StatEndurance:
		return "❤"
	default:
		return "?"
	}
}

func buildStatBlockWithBars(stats []models.StatLevel) fyne.CanvasObject {
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

	rows := container.NewVBox()
	for i, statType := range statOrder {
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

		barColor := statBarColor(statType)

		icon := canvas.NewText(statIcon(statType), barColor)
		icon.TextSize = 14

		code := canvas.NewText(fmt.Sprintf("%s %d", statShortCode(statType), stat.Level), components.T().Text)
		code.TextSize = 13
		code.TextStyle = fyne.TextStyle{Bold: true}

		expLabel := canvas.NewText(fmt.Sprintf("%d/%d", stat.CurrentEXP, expNeeded), components.T().TextSecondary)
		expLabel.TextSize = 11

		headerRow := container.NewHBox(icon, code, layout.NewSpacer(), expLabel)
		bar := makeColoredProgressBar(stat.CurrentEXP, expNeeded, barColor)

		rows.Add(container.NewVBox(headerRow, bar))

		// Add spacing between stats (not after the last one)
		if i < len(statOrder)-1 {
			spacer := canvas.NewRectangle(color.Transparent)
			spacer.SetMinSize(fyne.NewSize(0, 6))
			rows.Add(spacer)
		}
	}

	return rows
}

func makeColoredProgressBar(current, max int, barColor color.Color) fyne.CanvasObject {
	t := components.T()
	if max <= 0 {
		max = 1
	}
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}

	ratio := float64(current) / float64(max)

	bg := canvas.NewRectangle(t.BGPanel)
	bg.CornerRadius = components.RadiusSM
	bg.SetMinSize(fyne.NewSize(0, 10))
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.Border

	fill := canvas.NewRectangle(barColor)
	fill.CornerRadius = components.RadiusSM

	return container.NewStack(
		bg,
		container.New(&progressBarLayout{ratio: ratio}, fill),
	)
}

type progressBarLayout struct {
	ratio float64
}

func (p *progressBarLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 10)
}

func (p *progressBarLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, obj := range objects {
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(fyne.NewSize(containerSize.Width*float32(p.ratio), containerSize.Height))
	}
}

// =============================================================================
// Enemy Day Card
// =============================================================================

func buildEnemyDayCard(ctx *Context) *fyne.Container {
	enemyName := "Неизвестный враг"
	enemyRankText := "?"
	enemyDesc := ""
	var enemyAtk int
	var enemyZone int
	var enemy *models.Enemy

	if ctx.Features.Combat {
		e, err := ctx.Engine.GetNextEnemyForPlayer()
		if err == nil && e != nil {
			enemy = e
			enemyName = e.Name
			enemyRankText = string(e.Rank)
			enemyAtk = e.Attack
			enemyDesc = e.Description
			enemyZone = e.Zone
		} else if err == nil && e == nil {
			enemyName = "Все враги побеждены"
			enemyRankText = "-"
		}
	}

	quests, _ := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	activities, _ := ctx.Engine.DB.GetDailyActivityLast30(ctx.Engine.Character.ID)
	today := time.Now().Format("2006-01-02")
	completedToday := 0
	for _, a := range activities {
		if a.Date == today {
			completedToday = a.QuestsComplete
			break
		}
	}

	attempts := ctx.Engine.GetAttempts()

	// Enemy title
	enemyTitle := components.MakeSystemHeaderCompact("Следующий враг")

	// Enemy image — 200x200
	const enemyImgSize float32 = 200
	tk := components.T()
	enemyIconBg := canvas.NewRectangle(tk.BGPanel)
	enemyIconBg.CornerRadius = components.RadiusLG
	enemyIconBg.SetMinSize(fyne.NewSize(enemyImgSize, enemyImgSize))
	enemyIconBg.StrokeWidth = components.BorderThin
	enemyIconBg.StrokeColor = tk.Border

	var enemyIconBox fyne.CanvasObject
	if enemy != nil {
		if imgPath := resolveEnemyImagePath(*enemy); imgPath != "" {
			img := canvas.NewImageFromFile(imgPath)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(enemyImgSize, enemyImgSize))
			enemyIconBox = container.NewStack(enemyIconBg, img)
		}
	}
	if enemyIconBox == nil {
		fallback := canvas.NewImageFromResource(theme.VisibilityIcon())
		fallback.FillMode = canvas.ImageFillContain
		fallback.SetMinSize(fyne.NewSize(80, 80))
		enemyIconBox = container.NewStack(enemyIconBg, container.NewCenter(fallback))
	}

	// Enemy name + rank badge
	enemyNameLabel := components.MakeTitle(enemyName, components.T().Text, 20)
	enemyRankBadge := components.MakeRankBadge(models.QuestRank(enemyRankText))
	nameItems := []fyne.CanvasObject{enemyRankBadge, enemyNameLabel}
	if enemy != nil && (enemy.IsBoss || enemy.Type == models.EnemyBoss) {
		bossLabel := canvas.NewText("BOSS", components.T().Danger)
		bossLabel.TextSize = 11
		bossLabel.TextStyle = fyne.TextStyle{Bold: true}
		nameItems = append(nameItems, bossLabel)
	}
	nameRankRow := container.NewHBox(nameItems...)

	// Description + attack
	var infoItems []fyne.CanvasObject
	if enemy != nil {
		zoneLabel := components.MakeLabel(
			fmt.Sprintf("Zone %d · %s", enemyZone, zoneBiomeName(enemyZone)),
			components.T().Accent,
		)
		zoneLabel.TextSize = 12
		infoItems = append(infoItems, zoneLabel)
	}
	if enemyDesc != "" {
		descLabel := components.MakeLabel(enemyDesc, components.T().TextSecondary)
		descLabel.TextSize = 12
		infoItems = append(infoItems, descLabel)
	}
	if enemy != nil {
		statsLabel := components.MakeLabel(
			fmt.Sprintf("HP %d  ATK %d  Ранг %s", enemy.HP, enemyAtk, enemyRankText),
			components.T().TextSecondary,
		)
		statsLabel.TextSize = 12
		infoItems = append(infoItems, statsLabel)
	}

	// Difficulty bar
	if enemy != nil {
		diffSection := buildDifficultySection(ctx, enemy)
		infoItems = append(infoItems, diffSection)
	}

	// First-win reward
	if enemy != nil {
		rewardSection := buildFirstWinReward(ctx, enemy)
		infoItems = append(infoItems, rewardSection)
	}

	enemyInfo := container.NewVBox(append([]fyne.CanvasObject{nameRankRow}, infoItems...)...)
	enemyHeader := container.NewHBox(enemyIconBox, enemyInfo)

	// Attempts visual blocks
	attemptsSection := buildAttemptsBlocks(attempts)

	// Requirements / fight button
	activeTotal := len(quests)
	minQuestsForFight := 1
	canFight := ctx.Features.Combat && attempts > 0 && enemy != nil

	var ctaSection fyne.CanvasObject
	if !ctx.Features.Combat {
		hint := components.MakeLabel("Бой отключён", components.T().TextSecondary)
		hint.TextSize = 12
		ctaSection = hint
	} else if enemy == nil {
		hint := components.MakeLabel("Все враги побеждены", components.T().Gold)
		hint.TextSize = 13
		ctaSection = hint
	} else if attempts <= 0 {
		hint := components.MakeLabel("Нет попыток — выполни квест", components.T().Danger)
		hint.TextSize = 13
		ctaSection = hint
	} else if completedToday < minQuestsForFight && activeTotal > 0 {
		hint := components.MakeLabel(
			fmt.Sprintf("Нужно выполнить %d заданий", minQuestsForFight),
			components.T().Orange,
		)
		hint.TextSize = 13
		ctaSection = hint
	} else {
		ctaBtn := widget.NewButtonWithIcon("⚔ Вступить в бой", theme.MediaPlayIcon(), func() {
			if ctx.StartBattle != nil && enemy != nil {
				ctx.StartBattle(*enemy)
				return
			}
			dialog.ShowInformation("Бой", "Не удалось запустить бой.", ctx.Window)
		})
		ctaBtn.Importance = widget.HighImportance
		if !canFight {
			ctaBtn.Disable()
		}
		ctaSection = ctaBtn
	}

	content := container.NewVBox(
		enemyTitle,
		enemyHeader,
		widget.NewSeparator(),
		attemptsSection,
		layout.NewSpacer(),
		ctaSection,
	)

	// Enemy card: HUD panel with accent border for boss
	if enemy != nil && (enemy.IsBoss || enemy.Type == models.EnemyBoss) {
		return components.MakeHUDPanelAccent(content, tk.Danger)
	}
	return components.MakeHUDPanelSized(content, fyne.NewSize(0, 320))
}

// buildDifficultySection compares player stats to enemy floor to show difficulty.
func buildDifficultySection(ctx *Context, enemy *models.Enemy) fyne.CanvasObject {
	overallLevel, _ := ctx.Engine.GetOverallLevel()

	// Map enemy floor to an expected level:
	// Floor 1-5 → level ~1-5, Floor 6-10 → level ~8-15, Floor 11-15 → level ~20-40
	expectedLevel := enemy.Floor * 3
	diff := overallLevel - expectedLevel

	dt := components.T()
	var diffRatio float64
	var diffColor color.Color
	var diffText string
	switch {
	case diff >= 3:
		diffRatio = 0.2
		diffColor = dt.SuccessDim
		diffText = "Легко"
	case diff >= -2:
		diffRatio = 0.5
		diffColor = dt.Warning
		diffText = "В самый раз"
	default:
		diffRatio = 0.85
		diffColor = dt.DangerDim
		diffText = "Сложно"
	}

	label := components.MakeLabel("Сложность", components.T().TextSecondary)
	label.TextSize = 11

	diffLabel := canvas.NewText(diffText, diffColor)
	diffLabel.TextSize = 11
	diffLabel.TextStyle = fyne.TextStyle{Bold: true}

	barBg := canvas.NewRectangle(dt.BGPanel)
	barBg.CornerRadius = components.RadiusSM
	barBg.SetMinSize(fyne.NewSize(0, 8))
	barFill := canvas.NewRectangle(diffColor)
	barFill.CornerRadius = 3
	bar := container.NewStack(barBg, container.New(&progressBarLayout{ratio: diffRatio}, barFill))

	headerRow := container.NewHBox(label, layout.NewSpacer(), diffLabel)
	return container.NewVBox(headerRow, bar)
}

// buildFirstWinReward shows first-win rewards, or ✓ if already defeated.
func buildFirstWinReward(ctx *Context, enemy *models.Enemy) fyne.CanvasObject {
	reward, err := ctx.Engine.DB.GetBattleReward(ctx.Engine.Character.ID, enemy.ID)
	if err == nil && reward != nil {
		// Already defeated
		defeated := canvas.NewText("✓ Побеждён", components.T().Success)
		defeated.TextSize = 12
		defeated.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewHBox(defeated)
	}

	// Not yet defeated — show expected rewards
	title := fmt.Sprintf("Покоритель: %s", enemy.Name)
	badge := fmt.Sprintf("Знак: %s", enemy.Name)
	icon := canvas.NewText("🏆", components.T().Gold)
	icon.TextSize = 12
	label := components.MakeLabel("Первая победа:", components.T().TextSecondary)
	label.TextSize = 11
	titleLabel := canvas.NewText(title, components.T().Gold)
	titleLabel.TextSize = 11
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	badgeLabel := components.MakeLabel(badge, components.T().Gold)
	badgeLabel.TextSize = 11
	unlockLabel := components.MakeLabel("Продвигает прогресс зоны", components.T().Accent)
	unlockLabel.TextSize = 11

	return container.NewVBox(
		container.NewHBox(icon, label),
		titleLabel,
		badgeLabel,
		unlockLabel,
	)
}

func zoneBiomeName(zone int) string {
	switch zone {
	case 1:
		return "Forgotten Ruins"
	case 2:
		return "Abyss Corridors"
	case 3:
		return "Monarch Domain"
	default:
		return "Unknown Zone"
	}
}

// =============================================================================
// Attempts visual blocks (8 blocks in a row)
// =============================================================================

func buildAttemptsBlocks(attempts int) fyne.CanvasObject {
	t := components.T()
	label := components.MakeLabel(
		fmt.Sprintf("Попытки %d/%d", attempts, models.MaxAttempts),
		t.TextSecondary,
	)
	label.TextSize = components.TextBodySM

	var blocks []fyne.CanvasObject
	for i := 0; i < models.MaxAttempts; i++ {
		block := canvas.NewRectangle(t.BGPanel)
		if i < attempts {
			block.FillColor = t.AccentDim
		}
		block.CornerRadius = components.RadiusSM
		block.StrokeWidth = components.BorderThin
		block.StrokeColor = t.Border
		block.SetMinSize(fyne.NewSize(24, 14))
		blocks = append(blocks, block)
	}

	row := container.NewHBox(blocks...)
	return container.NewVBox(label, row)
}

// =============================================================================
// Streak + Attempts Card
// =============================================================================

func buildStreakLine(ctx *Context) fyne.CanvasObject {
	t := components.T()
	streak, _ := ctx.Engine.DB.GetStreak(ctx.Engine.Character.ID)

	streakLabel := components.MakeTitle(fmt.Sprintf("🔥 Streak: %d дней", streak), t.Accent, 14)

	var milestoneLabel *canvas.Text
	title := models.StreakTitle(streak)
	if title != "" {
		milestoneLabel = canvas.NewText(title, t.Gold)
	} else if streak == 0 {
		milestoneLabel = canvas.NewText("Начни сегодня!", t.Accent)
	} else {
		nextMilestone := 7
		for _, m := range models.AllStreakMilestones() {
			if streak < m.Days {
				nextMilestone = m.Days
				break
			}
		}
		milestoneLabel = canvas.NewText(
			fmt.Sprintf("До награды: %d дн.", nextMilestone-streak),
			t.TextSecondary,
		)
	}
	milestoneLabel.TextSize = components.TextBodyMD

	sep := canvas.NewText("|", t.TextMuted)
	sep.TextSize = components.TextBodyMD

	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = components.RadiusMD
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.Border
	row := container.NewHBox(streakLabel, sep, milestoneLabel)
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(6, 6, 10, 10), row))
}

// =============================================================================
// Today's Quests
// =============================================================================

func buildTodayQuestsWidget(ctx *Context) fyne.CanvasObject {
	quests, err := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	if err != nil {
		return components.MakeLabel("Ошибка: "+err.Error(), components.T().Danger)
	}

	if len(quests) == 0 {
		return components.MakeEmptyState("Нет активных заданий. Создайте новое!")
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
	return accordion
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
	titleText := components.MakeTitle(q.Title, components.T().Text, 14)

	var typeIndicator fyne.CanvasObject
	if q.IsDaily {
		lbl := components.MakeLabel("Ежедневное", components.T().Blue)
		lbl.TextSize = 11
		typeIndicator = lbl
	} else if q.DungeonID != nil {
		lbl := components.MakeLabel("Данж", components.T().Purple)
		lbl.TextSize = 11
		typeIndicator = lbl
	} else {
		typeIndicator = layout.NewSpacer()
	}

	statText := components.MakeLabel(
		fmt.Sprintf("+%d EXP -> %s | Ранг: %s | +%d попыток",
			q.Exp, q.TargetStat.DisplayName(), q.Rank, models.AttemptsForQuestEXP(q.Exp)),
		components.T().TextSecondary,
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

// =============================================================================
// Helpers
// =============================================================================

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
	t := components.T()
	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = components.RadiusLG
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.Border
	bg.SetMinSize(minSize)
	inset := container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), content)
	return container.NewStack(bg, inset)
}

func resolveEnemyImagePath(enemy models.Enemy) string {
	slugName := enemyImageSlug(enemy.Name)
	candidates := []string{}
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		candidates = append(candidates,
			filepath.Join("assets", "enemies", fmt.Sprintf("enemy_%d%s", enemy.ID, ext)),
			filepath.Join("assets", "enemies", slugName+ext),
			filepath.Join("assets", "enemies", fmt.Sprintf("floor_%d%s", enemy.Floor, ext)),
		)
	}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		for _, ext := range []string{".jpg", ".jpeg", ".png"} {
			candidates = append(candidates,
				filepath.Join(exeDir, "assets", "enemies", fmt.Sprintf("enemy_%d%s", enemy.ID, ext)),
				filepath.Join(exeDir, "assets", "enemies", slugName+ext),
				filepath.Join(exeDir, "assets", "enemies", fmt.Sprintf("floor_%d%s", enemy.Floor, ext)),
			)
		}
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func enemyImageSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "_")
	slug = strings.ReplaceAll(slug, "—", "_")
	slug = strings.ReplaceAll(slug, "-", "_")
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}
	return slug
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

func loadTrimmedAvatar(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	src, err := png.Decode(file)
	if err != nil {
		return nil
	}
	return trimTransparentBounds(src)
}

func trimTransparentBounds(src image.Image) image.Image {
	bounds := src.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	found := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a > 0 {
				found = true
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if !found {
		return src
	}

	const pad = 8
	if minX > bounds.Min.X+pad {
		minX -= pad
	} else {
		minX = bounds.Min.X
	}
	if minY > bounds.Min.Y+pad {
		minY -= pad
	} else {
		minY = bounds.Min.Y
	}
	if maxX+pad < bounds.Max.X-1 {
		maxX += pad
	} else {
		maxX = bounds.Max.X - 1
	}
	if maxY+pad < bounds.Max.Y-1 {
		maxY += pad
	} else {
		maxY = bounds.Max.Y - 1
	}

	trimmedRect := image.Rect(minX, minY, maxX+1, maxY+1)
	if !trimmedRect.Eq(bounds) {
		sub, ok := src.(interface {
			SubImage(r image.Rectangle) image.Image
		})
		if !ok {
			return src
		}
		src = sub.SubImage(trimmedRect)
		bounds = src.Bounds()
	}

	zoomedRect := zoomRect(bounds, 20)
	if zoomedRect.Eq(bounds) {
		return src
	}
	sub, ok := src.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return src
	}
	return sub.SubImage(zoomedRect)
}

// zoomRect shrinks rect by percent to make subject appear larger after fit.
func zoomRect(r image.Rectangle, percent int) image.Rectangle {
	if percent <= 0 || r.Dx() < 2 || r.Dy() < 2 {
		return r
	}

	scale := 100 + percent
	newW := r.Dx() * 100 / scale
	newH := r.Dy() * 100 / scale
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	cx := r.Min.X + r.Dx()/2
	cy := r.Min.Y + r.Dy()/2
	minX := cx - newW/2
	minY := cy - newH/2
	maxX := minX + newW
	maxY := minY + newH

	if minX < r.Min.X {
		minX = r.Min.X
		maxX = minX + newW
	}
	if minY < r.Min.Y {
		minY = r.Min.Y
		maxY = minY + newH
	}
	if maxX > r.Max.X {
		maxX = r.Max.X
		minX = maxX - newW
	}
	if maxY > r.Max.Y {
		maxY = r.Max.Y
		minY = maxY - newH
	}

	if minX < r.Min.X || minY < r.Min.Y || maxX > r.Max.X || maxY > r.Max.Y {
		return r
	}
	return image.Rect(minX, minY, maxX, maxY)
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
