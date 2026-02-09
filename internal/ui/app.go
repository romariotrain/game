package ui

import (
	"fmt"
	"image/color"
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
)

type App struct {
	engine *game.Engine
	window fyne.Window
	app    fyne.App

	characterPanel *fyne.Container
	questsPanel    *fyne.Container
	historyPanel   *fyne.Container
	skillsPanel    *fyne.Container
	statsPanel     *fyne.Container
	dungeonsPanel  *fyne.Container
	arenaPanel     *fyne.Container
	gachaPanel     *fyne.Container
	inventoryPanel *fyne.Container

	// Battle state
	currentBattle *models.BattleState
	battlePanel   *fyne.Container
}

func NewApp(fyneApp fyne.App, engine *game.Engine) *App {
	return &App{
		engine: engine,
		app:    fyneApp,
	}
}

func (a *App) Run() {
	a.window = a.app.NewWindow("SOLO LEVELING — Система Охотника")
	a.window.Resize(fyne.NewSize(1100, 800))
	a.window.CenterOnScreen()

	content := a.buildMainLayout()
	a.window.SetContent(content)
	a.window.ShowAndRun()
}

func (a *App) buildMainLayout() fyne.CanvasObject {
	header := a.buildHeader()

	charTab := container.NewTabItem("Охотник", a.buildCharacterTab())
	questsTab := container.NewTabItem("Задания", a.buildQuestsTab())
	dungeonsTab := container.NewTabItem("Данжи", a.buildDungeonsTab())
	arenaTab := container.NewTabItem("Арена", a.buildArenaTab())
	gachaTab := container.NewTabItem("Призыв", a.buildGachaTab())
	inventoryTab := container.NewTabItem("Инвентарь", a.buildInventoryTab())
	skillsTab := container.NewTabItem("Скилы", a.buildSkillsTab())
	statsTab := container.NewTabItem("Статистика", a.buildStatsTab())
	historyTab := container.NewTabItem("История", a.buildHistoryTab())

	tabs := container.NewAppTabs(charTab, questsTab, dungeonsTab, arenaTab, gachaTab, inventoryTab, skillsTab, statsTab, historyTab)
	tabs.SetTabLocation(container.TabLocationTop)

	return container.NewBorder(header, nil, nil, nil, tabs)
}

func (a *App) buildHeader() *fyne.Container {
	bg := canvas.NewRectangle(color.NRGBA{R: 15, G: 12, B: 30, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 56))

	title := canvas.NewText("S O L O   L E V E L I N G", ColorAccentBright)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Система Пробуждения Охотника", ColorTextDim)
	subtitle.TextSize = 12
	subtitle.Alignment = fyne.TextAlignCenter

	headerContent := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
	)

	return container.NewStack(bg, container.NewPadded(headerContent))
}

// ================================================================
// Character Tab
// ================================================================

func (a *App) buildCharacterTab() fyne.CanvasObject {
	a.characterPanel = container.NewVBox()
	a.refreshCharacterPanel()
	return container.NewVScroll(container.NewPadded(a.characterPanel))
}

func (a *App) refreshCharacterPanel() {
	a.characterPanel.Objects = nil

	stats, err := a.engine.GetStatLevels()
	if err != nil {
		a.characterPanel.Add(MakeLabel("Ошибка загрузки: "+err.Error(), ColorRed))
		return
	}

	overallLevel, _ := a.engine.GetOverallLevel()
	rankTitle := game.HunterRank(overallLevel)

	completedDungeons, _ := a.engine.DB.GetCompletedDungeons(a.engine.Character.ID)

	charCard := a.buildCharacterCard(overallLevel, rankTitle, stats, completedDungeons)
	a.characterPanel.Add(charCard)

	// Resources card
	resCard := a.buildResourcesCard()
	a.characterPanel.Add(resCard)

	// Daily reward card
	dailyCard := a.buildDailyRewardCard()
	a.characterPanel.Add(dailyCard)

	a.characterPanel.Add(widget.NewSeparator())
	a.characterPanel.Add(MakeSectionHeader("Характеристики"))
	for _, stat := range stats {
		row := MakeStatRow(stat)
		card := MakeCard(row)
		a.characterPanel.Add(card)
	}

	// Equipped items
	equipped, err := a.engine.DB.GetEquippedItems(a.engine.Character.ID)
	if err == nil && len(equipped) > 0 {
		a.characterPanel.Add(widget.NewSeparator())
		a.characterPanel.Add(MakeSectionHeader("Экипировка"))
		for _, eq := range equipped {
			card := a.buildEquipmentMiniCard(eq)
			a.characterPanel.Add(card)
		}
	}

	a.characterPanel.Refresh()
}

func (a *App) buildCharacterCard(level int, rank string, stats []models.StatLevel, completedDungeons []models.CompletedDungeon) *fyne.Container {
	nameText := MakeTitle(a.engine.Character.Name, ColorGold, 24)
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		entry := widget.NewEntry()
		entry.SetText(a.engine.Character.Name)
		dialog.ShowForm("Имя Охотника", "Сохранить", "Отмена",
			[]*widget.FormItem{widget.NewFormItem("Имя", entry)},
			func(ok bool) {
				if ok && strings.TrimSpace(entry.Text) != "" {
					a.engine.RenameCharacter(strings.TrimSpace(entry.Text))
					a.refreshCharacterPanel()
				}
			}, a.window)
	})

	nameRow := container.NewHBox(nameText, editBtn)

	rankColor := parseHexColor(game.HunterRankColor(level))
	rankText := canvas.NewText(rank, rankColor)
	rankText.TextSize = 16
	rankText.TextStyle = fyne.TextStyle{Bold: true}

	levelText := MakeTitle(fmt.Sprintf("Общий уровень: %d", level), ColorText, 16)

	var statSummary []fyne.CanvasObject
	for _, s := range stats {
		txt := MakeLabel(fmt.Sprintf("%s %s: %d", s.StatType.Icon(), s.StatType.DisplayName(), s.Level), ColorTextDim)
		statSummary = append(statSummary, txt)
	}

	top := container.NewVBox(nameRow, rankText, levelText)
	statsRow := container.NewHBox(statSummary...)

	contentItems := []fyne.CanvasObject{top, widget.NewSeparator(), statsRow}

	if len(completedDungeons) > 0 {
		var titles []fyne.CanvasObject
		titles = append(titles, MakeTitle("Титулы:", ColorGold, 13))
		for _, cd := range completedDungeons {
			titles = append(titles, MakeLabel("  "+cd.EarnedTitle, ColorPurple))
		}
		contentItems = append(contentItems, widget.NewSeparator())
		contentItems = append(contentItems, container.NewHBox(titles...))
	}

	content := container.NewVBox(contentItems...)
	return MakeCard(content)
}

func (a *App) buildResourcesCard() *fyne.Container {
	res, err := a.engine.GetResources()
	if err != nil {
		return MakeCard(MakeLabel("Ошибка ресурсов", ColorRed))
	}

	header := MakeTitle("Ресурсы", ColorAccentBright, 16)
	crystals := MakeLabel(fmt.Sprintf("Кристаллы: %d", res.Crystals), ColorGold)
	matCommon := MakeLabel(fmt.Sprintf("Обычные материалы: %d", res.MaterialCommon), parseHexColor(models.MaterialCommon.Color()))
	matRare := MakeLabel(fmt.Sprintf("Редкие материалы: %d", res.MaterialRare), parseHexColor(models.MaterialRare.Color()))
	matEpic := MakeLabel(fmt.Sprintf("Эпические материалы: %d", res.MaterialEpic), parseHexColor(models.MaterialEpic.Color()))

	content := container.NewVBox(header, widget.NewSeparator(), crystals, matCommon, matRare, matEpic)
	return MakeCard(content)
}

func (a *App) buildDailyRewardCard() *fyne.Container {
	canClaim, day, crystals, err := a.engine.GetDailyRewardInfo()
	if err != nil {
		return MakeCard(MakeLabel("Ошибка ежедневной награды", ColorRed))
	}

	header := MakeTitle("Ежедневная награда", ColorAccentBright, 16)

	// Show 7-day streak
	var dayBlocks []fyne.CanvasObject
	for i := 1; i <= 7; i++ {
		reward := models.DailyRewardCrystals(i)
		blockColor := color.NRGBA{R: 30, G: 30, B: 50, A: 255}
		textColor := ColorTextDim
		if i < day || (i == day && !canClaim) {
			blockColor = color.NRGBA{R: 50, G: 180, B: 80, A: 255}
			textColor = ColorText
		} else if i == day && canClaim {
			blockColor = ColorGold
			textColor = ColorBG
		}

		bg := canvas.NewRectangle(blockColor)
		bg.SetMinSize(fyne.NewSize(60, 50))
		bg.CornerRadius = 6

		dayLabel := canvas.NewText(fmt.Sprintf("Д%d", i), textColor)
		dayLabel.TextSize = 11
		dayLabel.Alignment = fyne.TextAlignCenter

		rewardLabel := canvas.NewText(fmt.Sprintf("%d", reward), textColor)
		rewardLabel.TextSize = 12
		rewardLabel.TextStyle = fyne.TextStyle{Bold: true}
		rewardLabel.Alignment = fyne.TextAlignCenter

		block := container.NewStack(bg, container.NewVBox(
			container.NewCenter(dayLabel),
			container.NewCenter(rewardLabel),
		))
		dayBlocks = append(dayBlocks, block)
	}

	streakRow := container.NewHBox(dayBlocks...)

	var actionObj fyne.CanvasObject
	if canClaim {
		claimBtn := widget.NewButtonWithIcon(
			fmt.Sprintf("Забрать %d кристаллов", crystals),
			theme.ConfirmIcon(),
			func() {
				reward, err := a.engine.ClaimDailyReward()
				if err != nil {
					dialog.ShowError(err, a.window)
					return
				}
				dialog.ShowInformation("Награда получена!",
					fmt.Sprintf("Получено %d кристаллов!\nДень серии: %d", reward.Crystals, reward.Day),
					a.window)
				a.refreshCharacterPanel()
			},
		)
		claimBtn.Importance = widget.HighImportance
		actionObj = claimBtn
	} else {
		actionObj = MakeLabel("Награда уже получена сегодня", ColorGreen)
	}

	content := container.NewVBox(header, widget.NewSeparator(), streakRow, actionObj)
	return MakeCard(content)
}

func (a *App) buildEquipmentMiniCard(eq models.Equipment) *fyne.Container {
	rarColor := parseHexColor(eq.Rarity.Color())
	nameText := MakeTitle(eq.Name, rarColor, 14)
	slotText := MakeLabel(eq.Slot.DisplayName(), ColorTextDim)

	var bonusText string
	switch eq.Slot {
	case models.SlotWeapon:
		bonusText = fmt.Sprintf("ATK +%d", eq.BonusAttack)
	case models.SlotArmor:
		bonusText = fmt.Sprintf("HP +%d", eq.BonusHP)
	case models.SlotAccessory:
		bonusText = fmt.Sprintf("TIME +%.1fs", eq.BonusTime)
	}
	bonus := MakeLabel(bonusText, ColorGreen)
	lvl := MakeLabel(fmt.Sprintf("Ур.%d", eq.Level), ColorAccentBright)

	row := container.NewHBox(nameText, slotText, layout.NewSpacer(), bonus, lvl)
	return MakeCard(row)
}

// ================================================================
// Quests Tab
// ================================================================

func (a *App) buildQuestsTab() fyne.CanvasObject {
	a.questsPanel = container.NewVBox()

	addBtn := MakeStyledButton("Новое задание", theme.ContentAddIcon(), a.showCreateQuestDialog)

	topBar := container.NewHBox(MakeSectionHeader("Активные Задания"), layout.NewSpacer(), addBtn)

	a.refreshQuestsPanel()

	return container.NewVScroll(container.NewPadded(
		container.NewVBox(topBar, a.questsPanel),
	))
}

func (a *App) refreshQuestsPanel() {
	a.questsPanel.Objects = nil

	quests, err := a.engine.DB.GetActiveQuests(a.engine.Character.ID)
	if err != nil {
		a.questsPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.questsPanel.Refresh()
		return
	}

	if len(quests) == 0 {
		a.questsPanel.Add(MakeEmptyState("Нет активных заданий. Создайте новое!"))
		a.questsPanel.Refresh()
		return
	}

	for _, q := range quests {
		quest := q
		card := a.buildQuestCard(quest)
		a.questsPanel.Add(card)
	}
	a.questsPanel.Refresh()
}

func (a *App) buildQuestCard(q models.Quest) *fyne.Container {
	rankBadge := MakeRankBadge(q.Rank)
	titleText := MakeTitle(q.Title, ColorText, 15)

	var dailyIndicator fyne.CanvasObject
	if q.IsDaily {
		dailyLabel := MakeLabel("Ежедневное", ColorBlue)
		dailyLabel.TextSize = 11
		dailyIndicator = dailyLabel
	} else if q.DungeonID != nil {
		dungeonLabel := MakeLabel("Данж", ColorPurple)
		dungeonLabel.TextSize = 11
		dailyIndicator = dungeonLabel
	} else {
		dailyIndicator = layout.NewSpacer()
	}

	statText := MakeLabel(
		fmt.Sprintf("%s %s  •  +%d EXP", q.TargetStat.Icon(), q.TargetStat.DisplayName(), q.Rank.BaseEXP()),
		ColorTextDim,
	)

	completeBtn := widget.NewButtonWithIcon("Выполнить", theme.ConfirmIcon(), func() {
		a.completeQuest(q)
	})
	completeBtn.Importance = widget.HighImportance

	failBtn := widget.NewButtonWithIcon("Провал", theme.CancelIcon(), func() {
		dialog.ShowConfirm("Провалить задание?",
			fmt.Sprintf("Провалить \"%s\"? EXP не будет начислен.", q.Title),
			func(ok bool) {
				if ok {
					a.engine.FailQuest(q.ID)
					a.refreshQuestsPanel()
					a.refreshStatsPanel()
				}
			}, a.window)
	})

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		msg := fmt.Sprintf("Удалить \"%s\"?", q.Title)
		if q.IsDaily {
			msg += "\nЕжедневный шаблон тоже будет деактивирован."
		}
		dialog.ShowConfirm("Удалить задание?", msg, func(ok bool) {
			if ok {
				a.engine.DeleteQuest(q.ID)
				a.refreshQuestsPanel()
			}
		}, a.window)
	})

	var descObj fyne.CanvasObject
	if q.Description != "" {
		descLabel := MakeLabel(q.Description, ColorTextDim)
		descObj = descLabel
	} else {
		descObj = layout.NewSpacer()
	}

	topRow := container.NewHBox(rankBadge, titleText, dailyIndicator, layout.NewSpacer(), completeBtn, failBtn, deleteBtn)
	content := container.NewVBox(topRow, statText, descObj)
	return MakeCard(content)
}

func (a *App) completeQuest(q models.Quest) {
	result, err := a.engine.CompleteQuest(q.ID)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	msg := fmt.Sprintf("Задание выполнено!\n\n+%d EXP к %s %s",
		result.EXPAwarded, result.StatType.Icon(), result.StatType.DisplayName())

	if result.LeveledUp {
		msg += fmt.Sprintf("\n\nУРОВЕНЬ ПОВЫШЕН! %s: %d -> %d",
			result.StatType.DisplayName(), result.OldLevel, result.NewLevel)

		for lvl := result.OldLevel + 1; lvl <= result.NewLevel; lvl++ {
			options := game.GetSkillOptions(result.StatType, lvl)
			if len(options) > 0 {
				msg += fmt.Sprintf("\n\nНовый скил доступен на уровне %d!", lvl)
				a.showSkillUnlockDialog(result.StatType, lvl)
			}
		}
	}

	if q.DungeonID != nil {
		done, err := a.engine.CheckDungeonCompletion(*q.DungeonID)
		if err == nil && done {
			if err := a.engine.CompleteDungeon(*q.DungeonID); err == nil {
				msg += "\n\nДАНЖ ПРОЙДЕН! Получена награда!"
			}
		}
	}

	dialog.ShowInformation("Задание выполнено!", msg, a.window)
	a.refreshAll()
}

func (a *App) refreshAll() {
	a.refreshQuestsPanel()
	a.refreshCharacterPanel()
	a.refreshHistoryPanel()
	a.refreshSkillsPanel()
	a.refreshStatsPanel()
	a.refreshDungeonsPanel()
}

func (a *App) showCreateQuestDialog() {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название задания...")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Описание (необязательно)...")
	descEntry.SetMinRowsVisible(2)

	rankSelect := widget.NewSelect([]string{"E", "D", "C", "B", "A", "S"}, nil)
	rankSelect.SetSelected("E")

	statNames := []string{"Сила", "Ловкость", "Интеллект", "Выносливость"}
	statSelect := widget.NewSelect(statNames, nil)
	statSelect.SetSelected("Сила")

	dailyCheck := widget.NewCheck("Ежедневное задание", nil)

	expLabel := MakeLabel("EXP: +20", ColorAccentBright)
	rankSelect.OnChanged = func(s string) {
		rank := models.QuestRank(s)
		expLabel.Text = fmt.Sprintf("EXP: +%d", rank.BaseEXP())
		expLabel.Refresh()
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Задание", titleEntry),
		widget.NewFormItem("Описание", descEntry),
		widget.NewFormItem("Ранг", container.NewHBox(rankSelect, expLabel)),
		widget.NewFormItem("Стат", statSelect),
		widget.NewFormItem("Тип", dailyCheck),
	}

	dialog.ShowForm("Новое Задание", "Создать", "Отмена", formItems, func(ok bool) {
		if !ok || strings.TrimSpace(titleEntry.Text) == "" {
			return
		}

		rank := models.QuestRank(rankSelect.Selected)
		statMap := map[string]models.StatType{
			"Сила": models.StatStrength, "Ловкость": models.StatAgility,
			"Интеллект": models.StatIntellect, "Выносливость": models.StatEndurance,
		}
		stat := statMap[statSelect.Selected]

		_, err := a.engine.CreateQuest(
			strings.TrimSpace(titleEntry.Text),
			strings.TrimSpace(descEntry.Text),
			rank, stat, dailyCheck.Checked,
		)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.refreshQuestsPanel()
	}, a.window)
}

func (a *App) showSkillUnlockDialog(stat models.StatType, level int) {
	options := game.GetSkillOptions(stat, level)
	if len(options) == 0 {
		return
	}

	opt := options[0]
	msg := fmt.Sprintf("Новый скил для %s на уровне %d!\n\n%s\n%s\nМультипликатор: x%.2f",
		stat.DisplayName(), level, opt.Name, opt.Description, opt.Multiplier)

	dialog.ShowConfirm("Скил разблокирован!", msg, func(ok bool) {
		if ok {
			a.engine.UnlockSkill(stat, level, 0)
			a.refreshSkillsPanel()
		}
	}, a.window)
}

// ================================================================
// Skills Tab
// ================================================================

func (a *App) buildSkillsTab() fyne.CanvasObject {
	a.skillsPanel = container.NewVBox()
	a.refreshSkillsPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Скилы Охотника"), a.skillsPanel),
	))
}

func (a *App) refreshSkillsPanel() {
	if a.skillsPanel == nil {
		return
	}
	a.skillsPanel.Objects = nil

	skills, err := a.engine.GetSkills()
	if err != nil {
		a.skillsPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.skillsPanel.Refresh()
		return
	}

	if len(skills) == 0 {
		a.skillsPanel.Add(MakeEmptyState("Скилы откроются при повышении уровней статов (Ур. 3, 5, 8, 10, 15)"))
		a.skillsPanel.Refresh()
		return
	}

	grouped := map[models.StatType][]models.Skill{}
	for _, s := range skills {
		grouped[s.StatType] = append(grouped[s.StatType], s)
	}

	for _, statType := range models.AllStats {
		statSkills, ok := grouped[statType]
		if !ok {
			continue
		}
		header := MakeTitle(fmt.Sprintf("%s %s", statType.Icon(), statType.DisplayName()), ColorAccentBright, 16)
		a.skillsPanel.Add(header)

		for _, sk := range statSkills {
			skill := sk
			card := a.buildSkillCard(skill)
			a.skillsPanel.Add(card)
		}
	}
	a.skillsPanel.Refresh()
}

func (a *App) buildSkillCard(skill models.Skill) *fyne.Container {
	nameText := MakeTitle(skill.Name, ColorGold, 14)
	descText := MakeLabel(skill.Description, ColorTextDim)
	multText := MakeLabel(fmt.Sprintf("x%.2f EXP", skill.Multiplier), ColorGreen)
	levelText := MakeLabel(fmt.Sprintf("Открыт на Ур. %d", skill.UnlockedAt), ColorTextDim)

	statusText := "ВКЛ"
	statusColor := ColorGreen
	if !skill.Active {
		statusText = "ВЫКЛ"
		statusColor = ColorRed
	}
	statusLabel := MakeLabel(statusText, statusColor)

	toggleBtn := widget.NewButton("Переключить", func() {
		a.engine.ToggleSkill(skill.ID, !skill.Active)
		a.refreshSkillsPanel()
	})

	topRow := container.NewHBox(nameText, layout.NewSpacer(), multText, statusLabel, toggleBtn)
	content := container.NewVBox(topRow, descText, levelText)
	return MakeCard(content)
}

// ================================================================
// History Tab
// ================================================================

func (a *App) buildHistoryTab() fyne.CanvasObject {
	a.historyPanel = container.NewVBox()
	a.refreshHistoryPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("История Заданий"), a.historyPanel),
	))
}

func (a *App) refreshHistoryPanel() {
	if a.historyPanel == nil {
		return
	}
	a.historyPanel.Objects = nil

	quests, err := a.engine.DB.GetCompletedQuests(a.engine.Character.ID, 50)
	if err != nil {
		a.historyPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.historyPanel.Refresh()
		return
	}

	if len(quests) == 0 {
		a.historyPanel.Add(MakeEmptyState("История пуста. Выполняйте задания!"))
		a.historyPanel.Refresh()
		return
	}

	for _, q := range quests {
		card := a.buildHistoryCard(q)
		a.historyPanel.Add(card)
	}
	a.historyPanel.Refresh()
}

func (a *App) buildHistoryCard(q models.Quest) *fyne.Container {
	rankBadge := MakeRankBadge(q.Rank)
	titleText := MakeTitle(q.Title, ColorText, 14)

	completedStr := ""
	if q.CompletedAt != nil {
		completedStr = q.CompletedAt.Format("02.01.2006 15:04")
	}

	var typeIndicator fyne.CanvasObject
	if q.IsDaily {
		lbl := MakeLabel("Ежедневное", ColorBlue)
		lbl.TextSize = 11
		typeIndicator = lbl
	} else if q.DungeonID != nil {
		lbl := MakeLabel("Данж", ColorPurple)
		lbl.TextSize = 11
		typeIndicator = lbl
	} else {
		typeIndicator = layout.NewSpacer()
	}

	dateText := MakeLabel(completedStr, ColorTextDim)
	expText := MakeLabel(
		fmt.Sprintf("+%d EXP -> %s %s", q.Rank.BaseEXP(), q.TargetStat.Icon(), q.TargetStat.DisplayName()),
		ColorGreen,
	)

	topRow := container.NewHBox(rankBadge, titleText, typeIndicator, layout.NewSpacer(), dateText)
	content := container.NewVBox(topRow, expText)
	return MakeCard(content)
}

// ================================================================
// Statistics Tab
// ================================================================

func (a *App) buildStatsTab() fyne.CanvasObject {
	a.statsPanel = container.NewVBox()
	a.refreshStatsPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Статистика Охотника"), a.statsPanel),
	))
}

func (a *App) refreshStatsPanel() {
	if a.statsPanel == nil {
		return
	}
	a.statsPanel.Objects = nil

	stats, err := a.engine.GetStatistics()
	if err != nil {
		a.statsPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.statsPanel.Refresh()
		return
	}

	overallCard := a.buildOverallStatsCard(stats)
	a.statsPanel.Add(overallCard)

	rankCard := a.buildRankStatsCard(stats)
	a.statsPanel.Add(rankCard)

	// Battle stats
	battleStats, err := a.engine.GetBattleStats()
	if err == nil && battleStats.TotalBattles > 0 {
		bCard := a.buildBattleStatsCard(battleStats)
		a.statsPanel.Add(bCard)
	}

	// Gacha stats
	gachaStats, err := a.engine.GetGachaStats()
	if err == nil && gachaStats.TotalPulls > 0 {
		gCard := a.buildGachaStatsCard(gachaStats)
		a.statsPanel.Add(gCard)
	}

	chartCard := a.buildActivityChart()
	a.statsPanel.Add(chartCard)

	a.statsPanel.Refresh()
}

func (a *App) buildOverallStatsCard(stats *models.Statistics) *fyne.Container {
	header := MakeTitle("Общая статистика", ColorAccentBright, 16)

	totalCompleted := MakeLabel(fmt.Sprintf("Выполнено заданий: %d", stats.TotalQuestsCompleted), ColorText)
	totalFailed := MakeLabel(fmt.Sprintf("Провалено заданий: %d", stats.TotalQuestsFailed), ColorText)
	successRate := MakeLabel(fmt.Sprintf("Процент успеха: %.1f%%", stats.SuccessRate), ColorGreen)
	totalEXP := MakeLabel(fmt.Sprintf("Всего EXP получено: %d", stats.TotalEXPEarned), ColorGold)
	streak := MakeLabel(fmt.Sprintf("Текущий Streak: %d дней подряд", stats.CurrentStreak), ColorAccentBright)
	bestStat := MakeLabel(
		fmt.Sprintf("Лучший стат: %s %s (Ур. %d)", stats.BestStat.Icon(), stats.BestStat.DisplayName(), stats.BestStatLevel),
		ColorText,
	)

	content := container.NewVBox(header, widget.NewSeparator(), totalCompleted, totalFailed, successRate, totalEXP, streak, bestStat)
	return MakeCard(content)
}

func (a *App) buildRankStatsCard(stats *models.Statistics) *fyne.Container {
	header := MakeTitle("Задания по рангам", ColorAccentBright, 16)

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())

	for _, rank := range models.AllRanks {
		count := stats.QuestsByRank[rank]
		clr := parseHexColor(rank.Color())
		label := MakeLabel(fmt.Sprintf("Ранг %s: %d заданий (+%d EXP каждое)", string(rank), count, rank.BaseEXP()), clr)
		rows = append(rows, label)
	}

	content := container.NewVBox(rows...)
	return MakeCard(content)
}

func (a *App) buildBattleStatsCard(stats *models.BattleStatistics) *fyne.Container {
	header := MakeTitle("Боевая статистика", ColorAccentBright, 16)

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())
	rows = append(rows, MakeLabel(fmt.Sprintf("Всего боёв: %d", stats.TotalBattles), ColorText))
	rows = append(rows, MakeLabel(fmt.Sprintf("Побед: %d / Поражений: %d", stats.Wins, stats.Losses), ColorText))
	rows = append(rows, MakeLabel(fmt.Sprintf("Winrate: %.1f%%", stats.WinRate), ColorGreen))
	rows = append(rows, MakeLabel(fmt.Sprintf("Общий урон: %d", stats.TotalDamage), ColorRed))
	rows = append(rows, MakeLabel(fmt.Sprintf("Критов: %d / Уклонений: %d", stats.TotalCrits, stats.TotalDodges), ColorAccentBright))

	if len(stats.EnemiesDefeated) > 0 {
		rows = append(rows, widget.NewSeparator())
		rows = append(rows, MakeTitle("Побеждённые враги:", ColorText, 14))
		for name, count := range stats.EnemiesDefeated {
			rows = append(rows, MakeLabel(fmt.Sprintf("  %s: %d раз", name, count), ColorTextDim))
		}
	}

	content := container.NewVBox(rows...)
	return MakeCard(content)
}

func (a *App) buildGachaStatsCard(stats *models.GachaStatistics) *fyne.Container {
	header := MakeTitle("Статистика призывов", ColorAccentBright, 16)

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())
	rows = append(rows, MakeLabel(fmt.Sprintf("Всего призывов: %d", stats.TotalPulls), ColorText))

	for _, rarity := range models.AllRarities {
		count := stats.PullsByRarity[rarity]
		if count > 0 {
			clr := parseHexColor(rarity.Color())
			rows = append(rows, MakeLabel(fmt.Sprintf("  %s: %d", rarity.DisplayName(), count), clr))
		}
	}

	content := container.NewVBox(rows...)
	return MakeCard(content)
}

func (a *App) buildActivityChart() *fyne.Container {
	header := MakeTitle("Активность за 30 дней", ColorAccentBright, 16)

	activities, err := a.engine.DB.GetDailyActivityLast30(a.engine.Character.ID)
	if err != nil || len(activities) == 0 {
		noData := MakeLabel("Нет данных об активности", ColorTextDim)
		content := container.NewVBox(header, widget.NewSeparator(), noData)
		return MakeCard(content)
	}

	activityMap := make(map[string]models.DailyActivity)
	for _, act := range activities {
		activityMap[act.Date] = act
	}

	var chartRows []fyne.CanvasObject
	chartRows = append(chartRows, header, widget.NewSeparator())

	var blocks []fyne.CanvasObject
	now := time.Now()
	for i := 29; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")

		blockColor := color.NRGBA{R: 25, G: 25, B: 40, A: 255}
		if act, ok := activityMap[dateStr]; ok {
			if act.QuestsComplete > 0 {
				intensity := act.QuestsComplete
				if intensity > 5 {
					intensity = 5
				}
				g := uint8(80 + intensity*30)
				blockColor = color.NRGBA{R: 30, G: g, B: 60, A: 255}
			}
		}

		block := canvas.NewRectangle(blockColor)
		block.SetMinSize(fyne.NewSize(18, 18))
		block.CornerRadius = 3
		blocks = append(blocks, block)
	}

	gridContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(20, 20)), blocks...)
	chartRows = append(chartRows, gridContainer)

	legendRow := container.NewHBox(
		MakeLabel("Меньше", ColorTextDim),
		makeSmallBlock(color.NRGBA{R: 25, G: 25, B: 40, A: 255}),
		makeSmallBlock(color.NRGBA{R: 30, G: 110, B: 60, A: 255}),
		makeSmallBlock(color.NRGBA{R: 30, G: 170, B: 60, A: 255}),
		makeSmallBlock(color.NRGBA{R: 30, G: 230, B: 60, A: 255}),
		MakeLabel("Больше", ColorTextDim),
	)
	chartRows = append(chartRows, legendRow)

	chartRows = append(chartRows, widget.NewSeparator())
	chartRows = append(chartRows, MakeTitle("Последние 7 дней:", ColorText, 14))
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		displayDate := day.Format("02.01")

		if act, ok := activityMap[dateStr]; ok {
			line := MakeLabel(
				fmt.Sprintf("  %s: %d заданий, +%d EXP", displayDate, act.QuestsComplete, act.EXPEarned),
				ColorText,
			)
			chartRows = append(chartRows, line)
		} else {
			line := MakeLabel(fmt.Sprintf("  %s: нет активности", displayDate), ColorTextDim)
			chartRows = append(chartRows, line)
		}
	}

	content := container.NewVBox(chartRows...)
	return MakeCard(content)
}

func makeSmallBlock(c color.Color) *canvas.Rectangle {
	block := canvas.NewRectangle(c)
	block.SetMinSize(fyne.NewSize(14, 14))
	block.CornerRadius = 2
	return block
}

// ================================================================
// Dungeons Tab
// ================================================================

func (a *App) buildDungeonsTab() fyne.CanvasObject {
	a.dungeonsPanel = container.NewVBox()
	a.refreshDungeonsPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Данжи"), a.dungeonsPanel),
	))
}

func (a *App) refreshDungeonsPanel() {
	if a.dungeonsPanel == nil {
		return
	}
	a.dungeonsPanel.Objects = nil

	a.engine.RefreshDungeonStatuses()

	dungeons, err := a.engine.DB.GetAllDungeons()
	if err != nil {
		a.dungeonsPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.dungeonsPanel.Refresh()
		return
	}

	if len(dungeons) == 0 {
		a.dungeonsPanel.Add(MakeEmptyState("Данжи не найдены"))
		a.dungeonsPanel.Refresh()
		return
	}

	completedDungeons, _ := a.engine.DB.GetCompletedDungeons(a.engine.Character.ID)
	if len(completedDungeons) > 0 {
		titlesHeader := MakeTitle("Полученные титулы:", ColorGold, 15)
		a.dungeonsPanel.Add(titlesHeader)
		for _, cd := range completedDungeons {
			titleLabel := MakeLabel(fmt.Sprintf("  %s (получен %s)", cd.EarnedTitle, cd.CompletedAt.Format("02.01.2006")), ColorPurple)
			a.dungeonsPanel.Add(titleLabel)
		}
		a.dungeonsPanel.Add(widget.NewSeparator())
	}

	for _, d := range dungeons {
		dungeon := d
		card := a.buildDungeonCard(dungeon)
		a.dungeonsPanel.Add(card)
	}
	a.dungeonsPanel.Refresh()
}

func (a *App) buildDungeonCard(d models.Dungeon) *fyne.Container {
	var statusIcon string
	var statusColor color.Color
	switch d.Status {
	case models.DungeonLocked:
		statusIcon = "LOCKED"
		statusColor = ColorTextDim
	case models.DungeonAvailable:
		statusIcon = "AVAILABLE"
		statusColor = ColorGreen
	case models.DungeonInProgress:
		statusIcon = "IN PROGRESS"
		statusColor = ColorBlue
	case models.DungeonCompleted:
		statusIcon = "COMPLETED"
		statusColor = ColorGold
	}

	nameText := MakeTitle(d.Name, ColorText, 16)
	statusBadge := MakeLabel(statusIcon, statusColor)
	statusBadge.TextStyle = fyne.TextStyle{Bold: true}

	descText := MakeLabel(d.Description, ColorTextDim)

	var reqParts []string
	for _, req := range d.Requirements {
		reqParts = append(reqParts, fmt.Sprintf("%s %s >= %d", req.StatType.Icon(), req.StatType.DisplayName(), req.MinLevel))
	}
	reqText := MakeLabel("Требования: "+strings.Join(reqParts, ", "), ColorTextDim)

	rewardText := MakeLabel(
		fmt.Sprintf("Награда: титул \"%s\" + %d EXP ко всем статам", d.RewardTitle, d.RewardEXP),
		ColorGold,
	)

	topRow := container.NewHBox(nameText, layout.NewSpacer(), statusBadge)

	var contentItems []fyne.CanvasObject
	contentItems = append(contentItems, topRow, descText, reqText, rewardText)

	if d.Status == models.DungeonInProgress {
		completed, total, err := a.engine.GetDungeonProgress(d.ID)
		if err == nil {
			progressText := MakeLabel(fmt.Sprintf("Прогресс: %d / %d заданий", completed, total), ColorAccentBright)
			progressBar := MakeEXPBar(completed, total, ColorAccentBright)
			contentItems = append(contentItems, progressText, progressBar)
		}

		allQuests, err := a.engine.DB.GetDungeonAllQuests(a.engine.Character.ID, d.ID)
		if err == nil && len(allQuests) > 0 {
			for _, q := range allQuests {
				var qStatus string
				var qColor color.Color
				switch q.Status {
				case models.QuestCompleted:
					qStatus = "[+]"
					qColor = ColorGreen
				case models.QuestActive:
					qStatus = "[ ]"
					qColor = ColorText
				default:
					qStatus = "[X]"
					qColor = ColorRed
				}
				ql := MakeLabel(fmt.Sprintf("  %s %s (%s)", qStatus, q.Title, string(q.Rank)), qColor)
				contentItems = append(contentItems, ql)
			}
		}
	}

	if d.Status == models.DungeonLocked || d.Status == models.DungeonAvailable {
		contentItems = append(contentItems, MakeLabel(fmt.Sprintf("Заданий в данже: %d", len(d.QuestDefinitions)), ColorTextDim))
		for _, qd := range d.QuestDefinitions {
			ql := MakeLabel(
				fmt.Sprintf("  - %s (Ранг %s, %s)", qd.Title, string(qd.Rank), qd.TargetStat.DisplayName()),
				ColorTextDim,
			)
			contentItems = append(contentItems, ql)
		}
	}

	if d.Status == models.DungeonAvailable {
		enterBtn := widget.NewButtonWithIcon("Войти в данж", theme.MediaPlayIcon(), func() {
			dialog.ShowConfirm("Войти в данж?",
				fmt.Sprintf("Войти в \"%s\"?\nБудет создано %d заданий.", d.Name, len(d.QuestDefinitions)),
				func(ok bool) {
					if ok {
						if err := a.engine.EnterDungeon(d.ID); err != nil {
							dialog.ShowError(err, a.window)
							return
						}
						a.refreshDungeonsPanel()
						a.refreshQuestsPanel()
					}
				}, a.window)
		})
		enterBtn.Importance = widget.HighImportance
		contentItems = append(contentItems, enterBtn)
	}

	if d.Status == models.DungeonCompleted {
		contentItems = append(contentItems, MakeLabel("Данж пройден!", ColorGold))
	}

	content := container.NewVBox(contentItems...)
	return MakeCard(content)
}

// ================================================================
// Arena Tab
// ================================================================

func (a *App) buildArenaTab() fyne.CanvasObject {
	a.arenaPanel = container.NewVBox()
	a.refreshArenaPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Арена — Выбор противника"), a.arenaPanel),
	))
}

func (a *App) refreshArenaPanel() {
	if a.arenaPanel == nil {
		return
	}
	a.arenaPanel.Objects = nil

	enemies, err := a.engine.GetEnemies()
	if err != nil {
		a.arenaPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.arenaPanel.Refresh()
		return
	}

	// Regular enemies
	a.arenaPanel.Add(MakeTitle("Обычные враги", ColorText, 16))
	for _, e := range enemies {
		if e.Type == models.EnemyRegular {
			enemy := e
			card := a.buildEnemyCard(enemy)
			a.arenaPanel.Add(card)
		}
	}

	a.arenaPanel.Add(widget.NewSeparator())
	a.arenaPanel.Add(MakeTitle("Боссы", ColorRed, 16))
	for _, e := range enemies {
		if e.Type == models.EnemyBoss {
			enemy := e
			card := a.buildEnemyCard(enemy)
			a.arenaPanel.Add(card)
		}
	}

	// Battle history
	battles, err := a.engine.GetBattleHistory(10)
	if err == nil && len(battles) > 0 {
		a.arenaPanel.Add(widget.NewSeparator())
		a.arenaPanel.Add(MakeTitle("Последние бои", ColorAccentBright, 16))
		for _, b := range battles {
			card := a.buildBattleHistoryCard(b)
			a.arenaPanel.Add(card)
		}
	}

	a.arenaPanel.Refresh()
}

func (a *App) buildEnemyCard(e models.Enemy) *fyne.Container {
	rankBadge := MakeRankBadge(e.Rank)

	var typeLabel *canvas.Text
	if e.Type == models.EnemyBoss {
		typeLabel = MakeLabel("БОСС", ColorRed)
		typeLabel.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		typeLabel = MakeLabel("", ColorTextDim)
	}

	nameText := MakeTitle(e.Name, ColorText, 15)
	descText := MakeLabel(e.Description, ColorTextDim)

	statsText := MakeLabel(
		fmt.Sprintf("HP: %d  ATK: %d  Паттерн: %d клеток  Время: %.1fs",
			e.HP, e.Attack, e.PatternSize, e.ShowTime),
		ColorTextDim,
	)

	rewardText := MakeLabel(
		fmt.Sprintf("Награда: +%d EXP, +%d кристаллов", e.RewardEXP, e.RewardCrystals),
		ColorGold,
	)

	fightBtn := widget.NewButtonWithIcon("Сражаться!", theme.MediaPlayIcon(), func() {
		a.startBattle(e)
	})
	fightBtn.Importance = widget.HighImportance

	topRow := container.NewHBox(rankBadge, nameText, typeLabel, layout.NewSpacer(), fightBtn)
	content := container.NewVBox(topRow, descText, statsText, rewardText)
	return MakeCard(content)
}

func (a *App) startBattle(enemy models.Enemy) {
	state, err := a.engine.StartBattle(enemy.ID)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.currentBattle = state
	a.showBattleScreen()
}

func (a *App) showBattleScreen() {
	state := a.currentBattle

	battleWindow := a.app.NewWindow(fmt.Sprintf("Бой: %s", state.Enemy.Name))
	battleWindow.Resize(fyne.NewSize(700, 650))
	battleWindow.CenterOnScreen()

	var contentRef *fyne.Container
	var selectedCells map[int]bool
	var cellButtons []*widget.Button
	var gridContainer *fyne.Container

	var rebuildScreen func()
	rebuildScreen = func() {
		contentRef.Objects = nil

		// HP bars
		playerHPBar := MakeEXPBar(state.PlayerHP, state.PlayerMaxHP, ColorGreen)
		enemyHPBar := MakeEXPBar(state.EnemyHP, state.EnemyMaxHP, ColorRed)

		playerLabel := MakeTitle(fmt.Sprintf("Охотник HP: %d/%d", state.PlayerHP, state.PlayerMaxHP), ColorGreen, 14)
		enemyLabel := MakeTitle(fmt.Sprintf("%s HP: %d/%d", state.Enemy.Name, state.EnemyHP, state.EnemyMaxHP), ColorRed, 14)

		roundLabel := MakeTitle(fmt.Sprintf("Раунд %d", state.Round), ColorAccentBright, 16)

		contentRef.Add(container.NewHBox(roundLabel))
		contentRef.Add(container.NewVBox(playerLabel, playerHPBar))
		contentRef.Add(container.NewVBox(enemyLabel, enemyHPBar))
		contentRef.Add(widget.NewSeparator())

		if state.BattleOver {
			var resultText string
			var resultColor color.Color
			if state.Result == models.BattleWin {
				resultText = "ПОБЕДА!"
				resultColor = ColorGold
			} else {
				resultText = "ПОРАЖЕНИЕ"
				resultColor = ColorRed
			}

			resultLabel := MakeTitle(resultText, resultColor, 24)
			contentRef.Add(container.NewCenter(resultLabel))

			// Finish battle and show rewards
			record, err := a.engine.FinishBattle(state)
			if err == nil && state.Result == models.BattleWin {
				rewardMsg := fmt.Sprintf("+%d EXP, +%d кристаллов", record.RewardEXP, record.RewardCrystals)
				if record.MaterialDrop != "" {
					rewardMsg += fmt.Sprintf(", +1 %s материал", models.MaterialTier(record.MaterialDrop).DisplayName())
				}
				contentRef.Add(MakeLabel(rewardMsg, ColorGold))
			}

			contentRef.Add(MakeLabel(
				fmt.Sprintf("Точность: %.1f%% | Криты: %d | Уклонения: %d",
					record.Accuracy, record.CriticalHits, record.Dodges),
				ColorTextDim,
			))

			closeBtn := widget.NewButtonWithIcon("Закрыть", theme.CancelIcon(), func() {
				battleWindow.Close()
				a.refreshArenaPanel()
				a.refreshCharacterPanel()
				a.refreshStatsPanel()
			})
			closeBtn.Importance = widget.HighImportance
			contentRef.Add(closeBtn)

			contentRef.Refresh()
			return
		}

		// Show pattern phase
		infoLabel := MakeLabel("Запомните подсвеченные клетки!", ColorGold)
		contentRef.Add(container.NewCenter(infoLabel))

		// Build 8x8 grid
		selectedCells = make(map[int]bool)
		cellButtons = make([]*widget.Button, 64)
		patternSet := make(map[int]bool)
		for _, p := range state.Pattern {
			patternSet[p] = true
		}

		var gridCells []fyne.CanvasObject
		for i := 0; i < 64; i++ {
			idx := i
			btn := widget.NewButton("", nil)
			btn.Importance = widget.LowImportance
			if patternSet[idx] {
				btn.Importance = widget.HighImportance
			}
			btn.Disable()
			cellButtons[idx] = btn
			gridCells = append(gridCells, btn)
		}

		gridContainer = container.New(layout.NewGridWrapLayout(fyne.NewSize(55, 55)), gridCells...)
		contentRef.Add(gridContainer)
		contentRef.Refresh()

		// After show time, hide pattern and enable clicking
		showTime, _ := a.engine.GetShowTime(state.Enemy.ShowTime)
		go func() {
			time.Sleep(time.Duration(showTime*1000) * time.Millisecond)

			// Hide pattern, enable buttons
			for i := 0; i < 64; i++ {
				idx := i
				cellButtons[idx].Importance = widget.LowImportance
				cellButtons[idx].Enable()
				cellButtons[idx].OnTapped = func() {
					if selectedCells[idx] {
						selectedCells[idx] = false
						cellButtons[idx].Importance = widget.LowImportance
					} else {
						selectedCells[idx] = true
						cellButtons[idx].Importance = widget.MediumImportance
					}
					cellButtons[idx].Refresh()
				}
				cellButtons[idx].Refresh()
			}

			infoLabel.Text = fmt.Sprintf("Выберите %d клеток и нажмите Атаковать!", len(state.Pattern))
			infoLabel.Refresh()

			// Add attack button
			attackBtn := widget.NewButtonWithIcon("Атаковать!", theme.ConfirmIcon(), func() {
				var guesses []int
				for idx, selected := range selectedCells {
					if selected {
						guesses = append(guesses, idx)
					}
				}

				err := a.engine.ProcessRound(state, guesses)
				if err != nil {
					dialog.ShowError(err, battleWindow)
					return
				}

				rebuildScreen()
			})
			attackBtn.Importance = widget.HighImportance
			contentRef.Add(attackBtn)
			contentRef.Refresh()
		}()
	}

	contentRef = container.NewVBox()
	rebuildScreen()

	battleWindow.SetContent(container.NewVScroll(container.NewPadded(contentRef)))
	battleWindow.Show()
}

func (a *App) buildBattleHistoryCard(b models.BattleRecord) *fyne.Container {
	var resultText string
	var resultColor color.Color
	if b.Result == models.BattleWin {
		resultText = "Победа"
		resultColor = ColorGreen
	} else {
		resultText = "Поражение"
		resultColor = ColorRed
	}

	nameText := MakeTitle(b.EnemyName, ColorText, 14)
	result := MakeLabel(resultText, resultColor)
	result.TextStyle = fyne.TextStyle{Bold: true}
	dateText := MakeLabel(b.FoughtAt.Format("02.01.2006 15:04"), ColorTextDim)

	statsText := MakeLabel(
		fmt.Sprintf("Урон: %d | Точность: %.0f%% | Криты: %d", b.DamageDealt, b.Accuracy, b.CriticalHits),
		ColorTextDim,
	)

	topRow := container.NewHBox(nameText, result, layout.NewSpacer(), dateText)
	content := container.NewVBox(topRow, statsText)
	return MakeCard(content)
}

// ================================================================
// Gacha Tab
// ================================================================

func (a *App) buildGachaTab() fyne.CanvasObject {
	a.gachaPanel = container.NewVBox()
	a.refreshGachaPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Призыв экипировки"), a.gachaPanel),
	))
}

func (a *App) refreshGachaPanel() {
	if a.gachaPanel == nil {
		return
	}
	a.gachaPanel.Objects = nil

	res, err := a.engine.GetResources()
	if err != nil {
		a.gachaPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.gachaPanel.Refresh()
		return
	}

	crystalLabel := MakeTitle(fmt.Sprintf("Кристаллы: %d", res.Crystals), ColorGold, 18)
	a.gachaPanel.Add(container.NewCenter(crystalLabel))
	a.gachaPanel.Add(widget.NewSeparator())

	// Pity counters
	pity, _ := a.engine.GetGachaPity()
	if pity != nil {
		pityInfo := MakeLabel(
			fmt.Sprintf("Гарант обычный: %d/30 до редкого+ | Гарант продвинутый: %d/50 до эпического+",
				pity.NormalPity, pity.AdvancedPity),
			ColorTextDim,
		)
		a.gachaPanel.Add(pityInfo)
	}
	a.gachaPanel.Add(widget.NewSeparator())

	// Normal banner
	normalCard := a.buildBannerCard(models.BannerNormal, res.Crystals)
	a.gachaPanel.Add(normalCard)

	// Advanced banner
	advancedCard := a.buildBannerCard(models.BannerAdvanced, res.Crystals)
	a.gachaPanel.Add(advancedCard)

	// Drop rates
	a.gachaPanel.Add(widget.NewSeparator())
	ratesCard := a.buildDropRatesCard()
	a.gachaPanel.Add(ratesCard)

	// History
	history, err := a.engine.GetGachaHistory(20)
	if err == nil && len(history) > 0 {
		a.gachaPanel.Add(widget.NewSeparator())
		a.gachaPanel.Add(MakeTitle("История призывов", ColorAccentBright, 16))
		for _, h := range history {
			clr := parseHexColor(h.Rarity.Color())
			dateStr := h.PulledAt.Format("02.01 15:04")
			label := MakeLabel(
				fmt.Sprintf("[%s] %s — %s", dateStr, h.Banner.DisplayName(), h.Rarity.DisplayName()),
				clr,
			)
			a.gachaPanel.Add(label)
		}
	}

	a.gachaPanel.Refresh()
}

func (a *App) buildBannerCard(banner models.GachaBanner, crystals int) *fyne.Container {
	header := MakeTitle(banner.DisplayName(), ColorAccentBright, 16)
	costLabel := MakeLabel(fmt.Sprintf("Стоимость: %d кристаллов", banner.Cost()), ColorGold)

	pullBtn := widget.NewButtonWithIcon("Призвать x1", theme.ContentAddIcon(), func() {
		a.doGachaPull(banner, 1)
	})
	pullBtn.Importance = widget.HighImportance
	if crystals < banner.Cost() {
		pullBtn.Disable()
	}

	multiCost := banner.Cost() * 10
	multiBtn := widget.NewButtonWithIcon(
		fmt.Sprintf("Призвать x10 (%d)", multiCost),
		theme.ContentAddIcon(),
		func() {
			a.doGachaPull(banner, 10)
		},
	)
	multiBtn.Importance = widget.HighImportance
	if crystals < multiCost {
		multiBtn.Disable()
	}

	btnRow := container.NewHBox(pullBtn, multiBtn)
	content := container.NewVBox(header, costLabel, btnRow)
	return MakeCard(content)
}

func (a *App) doGachaPull(banner models.GachaBanner, count int) {
	if count == 1 {
		eq, err := a.engine.GachaPull(banner)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.showPullResult([]*models.Equipment{eq})
	} else {
		results, err := a.engine.GachaMultiPull(banner, count)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.showPullResult(results)
	}
}

func (a *App) showPullResult(items []*models.Equipment) {
	var rows []fyne.CanvasObject
	rows = append(rows, MakeTitle("Результат призыва!", ColorGold, 18))
	rows = append(rows, widget.NewSeparator())

	for _, eq := range items {
		clr := parseHexColor(eq.Rarity.Color())
		nameLabel := MakeTitle(eq.Name, clr, 15)
		rarLabel := MakeLabel(eq.Rarity.DisplayName(), clr)
		slotLabel := MakeLabel(eq.Slot.DisplayName(), ColorTextDim)

		var bonusText string
		switch eq.Slot {
		case models.SlotWeapon:
			bonusText = fmt.Sprintf("ATK +%d", eq.BonusAttack)
		case models.SlotArmor:
			bonusText = fmt.Sprintf("HP +%d", eq.BonusHP)
		case models.SlotAccessory:
			bonusText = fmt.Sprintf("TIME +%.1fs", eq.BonusTime)
		}
		bonusLabel := MakeLabel(bonusText, ColorGreen)

		row := container.NewHBox(nameLabel, rarLabel, slotLabel, layout.NewSpacer(), bonusLabel)
		rows = append(rows, MakeCard(row))
	}

	content := container.NewVBox(rows...)
	dialog.ShowCustom("Призыв", "Закрыть", container.NewVScroll(content), a.window)
	a.refreshGachaPanel()
	a.refreshCharacterPanel()
}

func (a *App) buildDropRatesCard() *fyne.Container {
	header := MakeTitle("Шансы выпадения", ColorAccentBright, 14)

	normalHeader := MakeLabel("Обычный баннер:", ColorText)
	normalRates := MakeLabel("Common 60% | Uncommon 25% | Rare 10% | Epic 4% | Legendary 1%", ColorTextDim)
	normalPity := MakeLabel("Гарант: Rare+ каждые 30 призывов", ColorTextDim)

	advHeader := MakeLabel("Продвинутый баннер:", ColorText)
	advRates := MakeLabel("Common 37% | Uncommon 30% | Rare 20% | Epic 10% | Legendary 3%", ColorTextDim)
	advPity := MakeLabel("Гарант: Epic+ каждые 50 призывов", ColorTextDim)

	content := container.NewVBox(header, widget.NewSeparator(),
		normalHeader, normalRates, normalPity,
		widget.NewSeparator(),
		advHeader, advRates, advPity,
	)
	return MakeCard(content)
}

// ================================================================
// Inventory Tab
// ================================================================

func (a *App) buildInventoryTab() fyne.CanvasObject {
	a.inventoryPanel = container.NewVBox()
	a.refreshInventoryPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(MakeSectionHeader("Инвентарь"), a.inventoryPanel),
	))
}

func (a *App) refreshInventoryPanel() {
	if a.inventoryPanel == nil {
		return
	}
	a.inventoryPanel.Objects = nil

	// Resources bar
	res, _ := a.engine.GetResources()
	if res != nil {
		resLabel := MakeLabel(
			fmt.Sprintf("Кристаллы: %d | Обычные: %d | Редкие: %d | Эпические: %d",
				res.Crystals, res.MaterialCommon, res.MaterialRare, res.MaterialEpic),
			ColorGold,
		)
		a.inventoryPanel.Add(resLabel)
		a.inventoryPanel.Add(widget.NewSeparator())
	}

	equipment, err := a.engine.GetEquipment()
	if err != nil {
		a.inventoryPanel.Add(MakeLabel("Ошибка: "+err.Error(), ColorRed))
		a.inventoryPanel.Refresh()
		return
	}

	if len(equipment) == 0 {
		a.inventoryPanel.Add(MakeEmptyState("Инвентарь пуст. Используйте Призыв или Крафт!"))
	} else {
		// Equipped first
		a.inventoryPanel.Add(MakeTitle("Экипировано", ColorGreen, 15))
		hasEquipped := false
		for _, eq := range equipment {
			if eq.Equipped {
				hasEquipped = true
				eqCopy := eq
				card := a.buildEquipmentCard(eqCopy)
				a.inventoryPanel.Add(card)
			}
		}
		if !hasEquipped {
			a.inventoryPanel.Add(MakeLabel("  Ничего не экипировано", ColorTextDim))
		}

		a.inventoryPanel.Add(widget.NewSeparator())
		a.inventoryPanel.Add(MakeTitle("Все предметы", ColorAccentBright, 15))
		for _, eq := range equipment {
			if !eq.Equipped {
				eqCopy := eq
				card := a.buildEquipmentCard(eqCopy)
				a.inventoryPanel.Add(card)
			}
		}
	}

	// Crafting section
	a.inventoryPanel.Add(widget.NewSeparator())
	a.inventoryPanel.Add(MakeSectionHeader("Крафт"))
	recipes, err := a.engine.GetRecipes()
	if err == nil {
		for _, r := range recipes {
			recipe := r
			card := a.buildRecipeCard(recipe)
			a.inventoryPanel.Add(card)
		}
	}

	a.inventoryPanel.Refresh()
}

func (a *App) buildEquipmentCard(eq models.Equipment) *fyne.Container {
	rarColor := parseHexColor(eq.Rarity.Color())
	nameText := MakeTitle(eq.Name, rarColor, 15)
	rarText := MakeLabel(eq.Rarity.DisplayName(), rarColor)
	slotText := MakeLabel(eq.Slot.DisplayName(), ColorTextDim)
	lvlText := MakeLabel(fmt.Sprintf("Ур. %d", eq.Level), ColorAccentBright)

	var bonusText string
	switch eq.Slot {
	case models.SlotWeapon:
		bonusText = fmt.Sprintf("ATK +%d", eq.BonusAttack)
	case models.SlotArmor:
		bonusText = fmt.Sprintf("HP +%d", eq.BonusHP)
	case models.SlotAccessory:
		bonusText = fmt.Sprintf("TIME +%.1fs", eq.BonusTime)
	}
	bonusLabel := MakeLabel(bonusText, ColorGreen)

	// EXP bar
	required := models.EquipmentEXPForLevel(eq.Level)
	expBar := MakeEXPBar(eq.CurrentEXP, required, ColorAccent)

	// Buttons
	var actionBtns []fyne.CanvasObject

	if eq.Equipped {
		unequipBtn := widget.NewButton("Снять", func() {
			a.engine.UnequipItem(eq.ID)
			a.refreshInventoryPanel()
			a.refreshCharacterPanel()
		})
		actionBtns = append(actionBtns, unequipBtn)
	} else {
		equipBtn := widget.NewButtonWithIcon("Надеть", theme.ConfirmIcon(), func() {
			a.engine.EquipItem(eq.ID)
			a.refreshInventoryPanel()
			a.refreshCharacterPanel()
		})
		equipBtn.Importance = widget.HighImportance
		actionBtns = append(actionBtns, equipBtn)

		upgradeBtn := widget.NewButton("Улучшить", func() {
			a.showUpgradeDialog(eq)
		})
		actionBtns = append(actionBtns, upgradeBtn)

		dismantleBtn := widget.NewButton("Разобрать", func() {
			matTier, matCount := eq.Rarity.DismantleMaterial()
			crystals := eq.Rarity.DismantleCrystals() + (eq.Level-1)*5
			msg := fmt.Sprintf("Разобрать \"%s\"?\n\nПолучите:\n%d кристаллов\n%d %s материалов",
				eq.Name, crystals, matCount, matTier.DisplayName())

			dialog.ShowConfirm("Разобрать?", msg, func(ok bool) {
				if ok {
					a.engine.DismantleEquipment(eq.ID)
					a.refreshInventoryPanel()
					a.refreshCharacterPanel()
				}
			}, a.window)
		})
		actionBtns = append(actionBtns, dismantleBtn)

		sellBtn := widget.NewButton("Продать", func() {
			crystals := eq.Rarity.DismantleCrystals()/2 + (eq.Level-1)*2
			if crystals < 1 {
				crystals = 1
			}
			dialog.ShowConfirm("Продать?",
				fmt.Sprintf("Продать \"%s\" за %d кристаллов?", eq.Name, crystals),
				func(ok bool) {
					if ok {
						a.engine.SellEquipment(eq.ID)
						a.refreshInventoryPanel()
						a.refreshCharacterPanel()
					}
				}, a.window)
		})
		actionBtns = append(actionBtns, sellBtn)
	}

	equippedBadge := layout.NewSpacer()
	if eq.Equipped {
		badge := MakeLabel("EQUIPPED", ColorGreen)
		badge.TextStyle = fyne.TextStyle{Bold: true}
		equippedBadge = badge
	}

	topRow := container.NewHBox(nameText, rarText, slotText, equippedBadge, layout.NewSpacer(), lvlText)
	btnRow := container.NewHBox(actionBtns...)
	content := container.NewVBox(topRow, bonusLabel, expBar, btnRow)
	return MakeCard(content)
}

func (a *App) showUpgradeDialog(target models.Equipment) {
	equipment, err := a.engine.GetEquipment()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Filter out the target and equipped items
	var feedOptions []models.Equipment
	var feedLabels []string
	for _, eq := range equipment {
		if eq.ID != target.ID && !eq.Equipped {
			feedOptions = append(feedOptions, eq)
			feedLabels = append(feedLabels, fmt.Sprintf("%s (%s Ур.%d) +%d EXP",
				eq.Name, eq.Rarity.DisplayName(), eq.Level, eq.Rarity.BaseStats()*10+eq.Level*5))
		}
	}

	if len(feedOptions) == 0 {
		dialog.ShowInformation("Нет материала", "Нет экипировки для скармливания.", a.window)
		return
	}

	feedSelect := widget.NewSelect(feedLabels, nil)
	feedSelect.SetSelected(feedLabels[0])

	formItems := []*widget.FormItem{
		widget.NewFormItem("Скормить", feedSelect),
	}

	dialog.ShowForm(
		fmt.Sprintf("Улучшить %s", target.Name),
		"Улучшить", "Отмена", formItems,
		func(ok bool) {
			if !ok {
				return
			}
			idx := 0
			for i, l := range feedLabels {
				if l == feedSelect.Selected {
					idx = i
					break
				}
			}
			feed := feedOptions[idx]
			result, leveledUp, err := a.engine.UpgradeEquipment(target.ID, feed.ID)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			msg := fmt.Sprintf("Улучшено! %s теперь Ур. %d", result.Name, result.Level)
			if leveledUp {
				msg += "\nУровень повышен!"
			}
			dialog.ShowInformation("Улучшение", msg, a.window)
			a.refreshInventoryPanel()
			a.refreshCharacterPanel()
		}, a.window)
}

func (a *App) buildRecipeCard(r models.CraftRecipe) *fyne.Container {
	rarColor := parseHexColor(r.ResultRarity.Color())
	nameText := MakeTitle(r.Name, rarColor, 14)
	slotText := MakeLabel(r.ResultSlot.DisplayName(), ColorTextDim)
	rarText := MakeLabel(r.ResultRarity.DisplayName(), rarColor)

	var costParts []string
	if r.CostCrystals > 0 {
		costParts = append(costParts, fmt.Sprintf("%d кристаллов", r.CostCrystals))
	}
	if r.CostCommon > 0 {
		costParts = append(costParts, fmt.Sprintf("%d обычных", r.CostCommon))
	}
	if r.CostRare > 0 {
		costParts = append(costParts, fmt.Sprintf("%d редких", r.CostRare))
	}
	if r.CostEpic > 0 {
		costParts = append(costParts, fmt.Sprintf("%d эпических", r.CostEpic))
	}
	costText := MakeLabel("Стоимость: "+strings.Join(costParts, ", "), ColorTextDim)

	craftBtn := widget.NewButtonWithIcon("Создать", theme.ContentAddIcon(), func() {
		eq, err := a.engine.CraftItem(r.ID)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		clr := parseHexColor(eq.Rarity.Color())
		_ = clr
		dialog.ShowInformation("Создано!",
			fmt.Sprintf("Создано: %s (%s %s)", eq.Name, eq.Rarity.DisplayName(), eq.Slot.DisplayName()),
			a.window)
		a.refreshInventoryPanel()
		a.refreshCharacterPanel()
	})

	topRow := container.NewHBox(nameText, slotText, rarText, layout.NewSpacer(), craftBtn)
	content := container.NewVBox(topRow, costText)
	return MakeCard(content)
}

// ================================================================
// Utility
// ================================================================

func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%dд %dч назад", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dч назад", hours)
	}
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dм назад", minutes)
	}
	return "только что"
}
