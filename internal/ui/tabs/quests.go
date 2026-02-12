package tabs

import (
	"encoding/json"
	"fmt"
	"image/color"
	"sort"
	"strconv"
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

const (
	questThemeClassic = "Classic"
	questThemeSystem  = "System"
	questContentMaxW  = float32(1240)
)

func currentQuestThemeMode(ctx *Context) string {
	mode := strings.TrimSpace(ctx.QuestThemeMode)
	switch mode {
	case questThemeClassic, questThemeSystem:
		return mode
	default:
		if components.T().HeaderUppercase {
			return questThemeSystem
		}
		return questThemeClassic
	}
}

func isSystemQuestTheme(ctx *Context) bool {
	return currentQuestThemeMode(ctx) == questThemeSystem
}

func BuildQuests(ctx *Context) fyne.CanvasObject {
	ctx.QuestsPanel = container.NewVBox()
	RefreshQuests(ctx)

	listScroll := container.NewVScroll(ctx.QuestsPanel)
	content := container.NewBorder(nil, nil, nil, nil, listScroll)
	centered := container.New(&maxWidthCenterLayout{maxWidth: questContentMaxW}, content)
	padded := container.New(layout.NewCustomPaddedLayout(components.SpaceLG, components.SpaceLG, components.SpaceXL, components.SpaceLG), centered)

	bg := canvas.NewRectangle(darkenColor(components.T().BG, 8))
	return container.NewMax(bg, padded)
}

func RefreshQuests(ctx *Context) {
	if ctx.QuestsPanel == nil {
		return
	}
	ctx.QuestsPanel.Objects = nil

	t := components.T()
	quests, err := ctx.Engine.DB.GetActiveQuests(ctx.Engine.Character.ID)
	if err != nil {
		ctx.QuestsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), t.Danger))
		ctx.QuestsPanel.Refresh()
		return
	}

	if isSystemQuestTheme(ctx) {
		ctx.QuestsPanel.Add(buildSystemQuestDashboard(ctx, quests))
		ctx.QuestsPanel.Refresh()
		return
	}

	ctx.QuestsPanel.Add(buildQuestsHeader(ctx))

	if len(quests) == 0 {
		ctx.QuestsPanel.Add(components.MakeEmptyState("Нет активных заданий. Создайте новое!"))
		ctx.QuestsPanel.Refresh()
		return
	}

	for i, quest := range quests {
		if i > 0 {
			gap := canvas.NewRectangle(color.Transparent)
			gap.SetMinSize(fyne.NewSize(0, 18))
			ctx.QuestsPanel.Add(gap)
		}
		card := buildQuestCardClassic(ctx, quest)
		ctx.QuestsPanel.Add(card)
	}

	ctx.QuestsPanel.Refresh()
}

func buildQuestsHeader(ctx *Context) fyne.CanvasObject {
	t := components.T()
	title := components.MakeTitle("Активные задания", t.Accent, components.TextHeadingLG)

	importBtn := widget.NewButtonWithIcon("Импорт JSON", theme.FolderOpenIcon(), func() {
		showImportJSONDialog(ctx)
	})
	importBtn.Importance = widget.MediumImportance

	addBtn := widget.NewButtonWithIcon("+ Новое", theme.ContentAddIcon(), func() {
		showCreateQuestDialog(ctx)
	})
	addBtn.Importance = widget.HighImportance

	rightControls := container.NewHBox(importBtn, addBtn)
	centeredTitle := container.NewCenter(title)
	controlsRow := container.NewHBox(layout.NewSpacer(), rightControls)
	row := container.NewStack(centeredTitle, controlsRow)

	divider := canvas.NewRectangle(color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 96})
	divider.SetMinSize(fyne.NewSize(0, 1))

	panel := container.NewVBox(row, divider)
	return components.MakeCard(panel)
}

type questDashboardData struct {
	DailyQuests            []models.Quest
	MainQuests             []models.Quest
	LongTermQuests         []models.Quest
	DailyCompletedToday    int
	MainCompletedToday     int
	LongTermCompletedToday int
	ExpToday               int
	Streak                 int
	ActiveMission          *models.Quest
}

func buildSystemQuestDashboard(ctx *Context, quests []models.Quest) fyne.CanvasObject {
	data := collectQuestDashboardData(ctx, quests)
	left := buildQuestJournalColumn(ctx, data)
	right := buildHunterLogPanel(data)

	return container.New(&questDashboardLayout{
		leftRatio: 0.65,
		gap:       components.SpaceLG,
	}, left, right)
}

func buildQuestJournalColumn(ctx *Context, data questDashboardData) fyne.CanvasObject {
	t := components.T()

	title := components.MakeSystemLabel("[ ACTIVE QUESTS ]", t.Accent, components.TextHeadingMD)

	importBtn := widget.NewButtonWithIcon("Импорт JSON", theme.FolderOpenIcon(), func() {
		showImportJSONDialog(ctx)
	})
	importBtn.Importance = widget.MediumImportance

	addBtn := widget.NewButtonWithIcon("+ Новое", theme.ContentAddIcon(), func() {
		showCreateQuestDialog(ctx)
	})
	addBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, container.NewHBox(importBtn, addBtn), title)
	divider := canvas.NewRectangle(color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 96})
	divider.SetMinSize(fyne.NewSize(0, 1))

	accordion := widget.NewAccordion(
		widget.NewAccordionItem(
			fmt.Sprintf("ЕЖЕДНЕВНЫЕ %d/%d", data.DailyCompletedToday, len(data.DailyQuests)),
			buildQuestCategorySection(ctx, data.DailyQuests),
		),
		widget.NewAccordionItem(
			fmt.Sprintf("ОСНОВНЫЕ %d/%d", data.MainCompletedToday, len(data.MainQuests)),
			buildQuestCategorySection(ctx, data.MainQuests),
		),
		widget.NewAccordionItem(
			fmt.Sprintf("ДОЛГОСРОЧНЫЕ %d", len(data.LongTermQuests)),
			buildQuestCategorySection(ctx, data.LongTermQuests),
		),
	)
	accordion.MultiOpen = true
	accordion.Open(0)
	accordion.Open(1)

	content := container.NewVBox(
		header,
		divider,
		makeVerticalGap(components.SpaceSM),
		accordion,
	)
	return makeQuestDashboardPanel(content, t.BGPanel, t.Border, components.RadiusXL)
}

func buildQuestCategorySection(ctx *Context, quests []models.Quest) fyne.CanvasObject {
	if len(quests) == 0 {
		return components.MakeEmptyState("Нет задач в этой категории.")
	}
	section := container.NewVBox()
	for i, q := range quests {
		if i > 0 {
			section.Add(makeVerticalGap(16))
		}
		section.Add(buildQuestCardSystem(ctx, q))
	}
	return section
}

func buildHunterLogPanel(data questDashboardData) fyne.CanvasObject {
	t := components.T()

	title := components.MakeSystemLabel("[ HUNTER LOG ]", t.AccentDim, components.TextHeadingMD)
	divider := canvas.NewRectangle(color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 96})
	divider.SetMinSize(fyne.NewSize(0, 1))

	progressTitle := components.MakeSystemLabel("СЕГОДНЯ", t.TextSecondary, components.TextHeadingSM)
	expToday := components.MakeTitle(fmt.Sprintf("+%d EXP", data.ExpToday), t.Accent, components.TextHeadingLG)

	statsTitle := components.MakeSystemLabel("СТАТИСТИКА ЗАДАЧ", t.TextSecondary, components.TextHeadingSM)
	statsBox := container.NewVBox(
		makeHunterLogLine("Daily", fmt.Sprintf("%d/%d", data.DailyCompletedToday, len(data.DailyQuests)), t.Text),
		makeHunterLogLine("Main", fmt.Sprintf("%d/%d", data.MainCompletedToday, len(data.MainQuests)), t.Text),
		makeHunterLogLine("Long-term", fmt.Sprintf("%d", len(data.LongTermQuests)), t.Text),
	)

	activeTitle := components.MakeSystemLabel("АКТИВНАЯ МИССИЯ", t.TextSecondary, components.TextHeadingSM)
	activeMission := makeActiveMissionBlock(data.ActiveMission)

	streakTitle := components.MakeSystemLabel("STREAK", t.TextSecondary, components.TextHeadingSM)
	streakValue := components.MakeTitle(fmt.Sprintf("%d дней", data.Streak), t.Accent, components.TextHeadingMD)

	content := container.NewVBox(
		title,
		divider,
		makeVerticalGap(components.SpaceSM),
		progressTitle,
		expToday,
		makeVerticalGap(16),
		statsTitle,
		statsBox,
		makeVerticalGap(16),
		activeTitle,
		activeMission,
		makeVerticalGap(16),
		streakTitle,
		streakValue,
	)

	secondaryBg := darkenColor(t.BGPanel, 6)
	secondaryBorder := color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 180}
	return makeQuestDashboardPanel(content, secondaryBg, secondaryBorder, components.RadiusXL)
}

func makeActiveMissionBlock(mission *models.Quest) fyne.CanvasObject {
	t := components.T()
	if mission == nil {
		empty := components.MakeLabel("Нет активной миссии", t.TextSecondary)
		empty.TextSize = components.TextBodyMD
		return empty
	}

	name := components.MakeTitle(strings.ToUpper(strings.TrimSpace(mission.Title)), t.Text, components.TextBodyMD)
	category := components.MakeLabel(
		fmt.Sprintf("%s • %s", categoryTitleForQuest(mission), questStatCode(mission.TargetStat)),
		t.TextSecondary,
	)
	category.TextSize = components.TextBodySM

	reward := components.MakeLabel(
		fmt.Sprintf("+%d EXP", mission.Exp),
		t.Accent,
	)
	reward.TextSize = components.TextBodySM

	return container.NewVBox(name, category, reward)
}

func makeHunterLogLine(label string, value string, valueColor color.Color) fyne.CanvasObject {
	t := components.T()
	left := components.MakeLabel(label, t.TextSecondary)
	left.TextSize = components.TextBodySM
	right := components.MakeLabel(value, valueColor)
	right.TextSize = components.TextBodySM
	return container.NewBorder(nil, nil, nil, right, left)
}

func makeQuestDashboardPanel(content fyne.CanvasObject, bgColor color.NRGBA, borderColor color.NRGBA, radius float32) fyne.CanvasObject {
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = radius
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = borderColor
	inset := container.New(layout.NewCustomPaddedLayout(20, 20, 22, 22), content)
	return container.NewStack(bg, inset)
}

func makeVerticalGap(height float32) fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(0, height))
	return gap
}

func collectQuestDashboardData(ctx *Context, quests []models.Quest) questDashboardData {
	data := questDashboardData{}

	for _, q := range quests {
		switch classifyQuestJournal(q) {
		case "daily":
			data.DailyQuests = append(data.DailyQuests, q)
		case "longterm":
			data.LongTermQuests = append(data.LongTermQuests, q)
		default:
			data.MainQuests = append(data.MainQuests, q)
		}
	}

	sortTodayQuestSection(data.DailyQuests)
	sortTodayQuestSection(data.MainQuests)
	sortTodayQuestSection(data.LongTermQuests)
	data.ActiveMission = pickActiveMission(quests)

	streak, err := ctx.Engine.DB.GetStreak(ctx.Engine.Character.ID)
	if err == nil {
		data.Streak = streak
	}

	activities, err := ctx.Engine.DB.GetDailyActivityLast30(ctx.Engine.Character.ID)
	if err == nil {
		today := time.Now().Format("2006-01-02")
		for _, a := range activities {
			if a.Date == today {
				data.ExpToday = a.EXPEarned
				break
			}
		}
	}

	completedToday, err := ctx.Engine.DB.GetCompletedQuests(ctx.Engine.Character.ID, 300)
	if err == nil {
		today := time.Now().Format("2006-01-02")
		for _, q := range completedToday {
			if q.CompletedAt == nil || q.CompletedAt.Format("2006-01-02") != today {
				continue
			}
			switch classifyQuestJournal(q) {
			case "daily":
				data.DailyCompletedToday++
			case "longterm":
				data.LongTermCompletedToday++
			default:
				data.MainCompletedToday++
			}
		}
	}

	return data
}

func classifyQuestJournal(q models.Quest) string {
	if q.IsDaily {
		return "daily"
	}
	if q.DungeonID != nil || q.Rank == models.RankA || q.Rank == models.RankS || q.Exp >= 40 {
		return "longterm"
	}
	return "main"
}

func categoryTitleForQuest(q *models.Quest) string {
	if q == nil {
		return "—"
	}
	switch classifyQuestJournal(*q) {
	case "daily":
		return "ЕЖЕДНЕВНАЯ"
	case "longterm":
		return "ДОЛГОСРОЧНАЯ"
	default:
		return "ОСНОВНАЯ"
	}
}

func pickActiveMission(quests []models.Quest) *models.Quest {
	if len(quests) == 0 {
		return nil
	}
	ordered := append([]models.Quest(nil), quests...)
	sort.Slice(ordered, func(i, j int) bool {
		leftRank := questRankSortWeight(ordered[i].Rank)
		rightRank := questRankSortWeight(ordered[j].Rank)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if ordered[i].Exp != ordered[j].Exp {
			return ordered[i].Exp > ordered[j].Exp
		}
		return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
	})
	selected := ordered[0]
	return &selected
}

type questDashboardLayout struct {
	leftRatio float32
	gap       float32
}

func (l *questDashboardLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}
	left := objects[0].MinSize()
	right := objects[1].MinSize()
	return fyne.NewSize(left.Width+right.Width+l.gap, maxFloat32(left.Height, right.Height))
}

func (l *questDashboardLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	ratio := l.leftRatio
	if ratio <= 0 || ratio >= 1 {
		ratio = 0.65
	}
	available := size.Width - l.gap
	if available < 0 {
		available = 0
	}
	leftWidth := available * ratio
	rightWidth := available - leftWidth
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(leftWidth, size.Height))
	objects[1].Move(fyne.NewPos(leftWidth+l.gap, 0))
	objects[1].Resize(fyne.NewSize(rightWidth, size.Height))
}

func maxFloat32(a float32, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

type maxWidthCenterLayout struct {
	maxWidth float32
}

func (l *maxWidthCenterLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	if l.maxWidth > 0 && min.Width > l.maxWidth {
		min.Width = l.maxWidth
	}
	return min
}

func (l *maxWidthCenterLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	width := size.Width
	if l.maxWidth > 0 && width > l.maxWidth {
		width = l.maxWidth
	}
	x := (size.Width - width) / 2
	objects[0].Move(fyne.NewPos(x, 0))
	objects[0].Resize(fyne.NewSize(width, size.Height))
}

func darkenColor(c color.NRGBA, delta uint8) color.NRGBA {
	return color.NRGBA{
		R: saturatingSub(c.R, delta),
		G: saturatingSub(c.G, delta),
		B: saturatingSub(c.B, delta),
		A: c.A,
	}
}

func saturatingSub(v uint8, d uint8) uint8 {
	if v <= d {
		return 0
	}
	return v - d
}

func questStatCode(stat models.StatType) string {
	switch stat {
	case models.StatStrength:
		return "STR"
	case models.StatAgility:
		return "AGI"
	case models.StatIntellect:
		return "INT"
	case models.StatEndurance:
		return "STA"
	default:
		return strings.ToUpper(stat.DisplayName())
	}
}

func buildQuestCardClassic(ctx *Context, q models.Quest) *fyne.Container {
	t := components.T()
	rankBadge := components.MakeRankBadge(q.Rank)
	titleText := components.MakeTitle(q.Title, t.Text, components.TextBodyLG)

	var dailyIndicator fyne.CanvasObject
	if q.IsDaily {
		dailyLabel := components.MakeLabel("Ежедневное", t.Blue)
		dailyLabel.TextSize = components.TextBodySM
		dailyIndicator = dailyLabel
	} else if q.DungeonID != nil {
		dungeonLabel := components.MakeLabel("Данж", t.Purple)
		dungeonLabel.TextSize = components.TextBodySM
		dailyIndicator = dungeonLabel
	} else {
		dailyIndicator = layout.NewSpacer()
	}

	statText := components.MakeLabel(
		fmt.Sprintf("Цель: %s %s", q.TargetStat.Icon(), q.TargetStat.DisplayName()),
		t.TextSecondary,
	)
	rewardText := components.MakeLabel(
		fmt.Sprintf("Награда: +%d EXP | Ранг: %s", q.Exp, q.Rank),
		t.Accent,
	)

	var descLabel fyne.CanvasObject
	if q.Description != "" {
		descLabel = components.MakeLabel(q.Description, t.TextSecondary)
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

func buildQuestCardSystem(ctx *Context, q models.Quest) *fyne.Container {
	tag := ""
	if q.IsDaily {
		tag = "Ежедневное"
	} else if q.DungeonID != nil {
		tag = "Данж"
	}

	onComplete := func() {
		completeQuest(ctx, q)
	}
	onFail := func() {
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
	}
	onDelete := func() {
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
	}

	data := components.QuestCardSystemData{
		Rank:        q.Rank,
		Title:       q.Title,
		MetaStat:    questStatCode(q.TargetStat),
		EXP:         q.Exp,
		Description: q.Description,
		Tag:         tag,
		Priority:    q.Rank == models.RankA || q.Rank == models.RankS,
	}
	actions := components.QuestCardSystemActions{
		OnComplete: onComplete,
		OnFail:     onFail,
		OnDelete:   onDelete,
	}
	return components.MakeQuestCardSystem(data, actions)
}

func completeQuest(ctx *Context, q models.Quest) {
	result, err := ctx.Engine.CompleteQuest(q.ID)
	if err != nil {
		dialog.ShowError(err, ctx.Window)
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
	t := components.T()

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

	expLabel := components.MakeLabel("", t.Accent)
	updatePreview := func() {
		minutes := parseIntWithDefault(minutesEntry.Text, 20)
		effort := parseIntWithDefault(effortSelect.Selected, 2)
		friction := parseIntWithDefault(frictionSelect.Selected, 1)
		exp := models.CalculateQuestEXP(minutes, effort, friction)
		expLabel.Text = fmt.Sprintf("EXP: +%d | Ранг: %s", exp, models.RankFromEXP(exp))
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
	t := components.T()

	entry := widget.NewMultiLineEntry()
	entry.SetMinRowsVisible(10)
	entry.SetPlaceHolder(`[{"title":"...","desc":"...","minutes":25,"effort":3,"friction":2,"stat":"INT","is_daily":false}]`)

	hint := components.MakeLabel("Вставьте JSON массив заданий", t.TextSecondary)

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
