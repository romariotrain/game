package ui

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/config"
	"solo-leveling/internal/game"
	"solo-leveling/internal/game/combat/boss"
	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
	"solo-leveling/internal/ui/tabs"
)

type App struct {
	engine   *game.Engine
	window   fyne.Window
	app      fyne.App
	features config.Features
	tabsCtx  *tabs.Context

	characterPanel *fyne.Container
	questsPanel    *fyne.Container
	historyPanel   *fyne.Container
	statsPanel     *fyne.Container
	dungeonsPanel  *fyne.Container
	arenaPanel     *fyne.Container

	// Battle state
	currentBattle *models.BattleState
	battlePanel   *fyne.Container
	currentBoss   *boss.State
}

func NewApp(fyneApp fyne.App, engine *game.Engine, features config.Features) *App {
	return &App{
		engine:   engine,
		app:      fyneApp,
		features: features,
	}
}

func (a *App) Run() {
	a.window = a.app.NewWindow("SOLO LEVELING — Система Охотника")
	a.window.Resize(fyne.NewSize(1100, 800))
	a.window.CenterOnScreen()

	a.tabsCtx = &tabs.Context{
		Engine:   a.engine,
		Window:   a.window,
		App:      a.app,
		Features: a.features,
		RefreshAll: func() {
			a.refreshAll()
		},
		RefreshCharacter: func() {
			a.refreshCharacterPanel()
		},
		RefreshQuests: func() {
			a.refreshQuestsPanel()
		},
		RefreshStats: func() {
			a.refreshStatsPanel()
		},
		RefreshDungeons: func() {
			a.refreshDungeonsPanel()
		},
		RefreshHistory: func() {
			a.refreshHistoryPanel()
		},
	}

	content := a.buildMainLayout()
	a.window.SetContent(content)
	a.window.ShowAndRun()
}

func (a *App) buildMainLayout() fyne.CanvasObject {
	header := a.buildHeader()

	var appTabs *container.AppTabs
	if a.features.MinimalMode {
		todayTab := container.NewTabItem("Сегодня", tabs.BuildToday(a.tabsCtx))
		questsTab := container.NewTabItem("Задания", tabs.BuildQuests(a.tabsCtx))
		progressTab := container.NewTabItem("Прогресс", tabs.BuildProgress(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		appTabs = container.NewAppTabs(todayTab, questsTab, progressTab, dungeonsTab)
	} else {
		charTab := container.NewTabItem("Охотник", tabs.BuildToday(a.tabsCtx))
		questsTab := container.NewTabItem("Задания", tabs.BuildQuests(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		statsTab := container.NewTabItem("Статистика", tabs.BuildProgress(a.tabsCtx))

		tabItems := []*container.TabItem{charTab, questsTab, dungeonsTab, statsTab}
		if a.features.Combat {
			tabItems = append(tabItems, container.NewTabItem("Арена", a.buildArenaTab()))
		}
		tabItems = append(tabItems,
			container.NewTabItem("История", a.buildHistoryTab()),
		)
		appTabs = container.NewAppTabs(tabItems...)
	}
	appTabs.SetTabLocation(container.TabLocationTop)

	return container.NewBorder(header, nil, nil, nil, appTabs)
}

func (a *App) buildHeader() *fyne.Container {
	bg := canvas.NewRectangle(color.NRGBA{R: 15, G: 12, B: 30, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 56))

	title := canvas.NewText("S O L O   L E V E L I N G", components.ColorAccentBright)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Система Пробуждения Охотника", components.ColorTextDim)
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
	a.characterPanel = a.tabsCtx.CharacterPanel
	return tabs.BuildToday(a.tabsCtx)
}

func (a *App) refreshCharacterPanel() {
	tabs.RefreshToday(a.tabsCtx)
}

func (a *App) buildCharacterCard(level int, rank string, stats []models.StatLevel, completedDungeons []models.CompletedDungeon) *fyne.Container {
	nameText := components.MakeTitle(a.engine.Character.Name, components.ColorGold, 24)
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

	if len(completedDungeons) > 0 {
		var titles []fyne.CanvasObject
		titles = append(titles, components.MakeTitle("Титулы:", components.ColorGold, 13))
		for _, cd := range completedDungeons {
			titles = append(titles, components.MakeLabel("  "+cd.EarnedTitle, components.ColorPurple))
		}
		contentItems = append(contentItems, widget.NewSeparator())
		contentItems = append(contentItems, container.NewHBox(titles...))
	}

	content := container.NewVBox(contentItems...)
	return components.MakeCard(content)
}

// ================================================================
// Quests Tab
// ================================================================

func (a *App) buildQuestsTab() fyne.CanvasObject {
	a.questsPanel = a.tabsCtx.QuestsPanel
	return tabs.BuildQuests(a.tabsCtx)
}

func (a *App) refreshQuestsPanel() {
	tabs.RefreshQuests(a.tabsCtx)
}

func (a *App) refreshAll() {
	a.refreshQuestsPanel()
	a.refreshCharacterPanel()
	a.refreshHistoryPanel()
	a.refreshStatsPanel()
	a.refreshDungeonsPanel()
}

// ================================================================
// History Tab
// ================================================================

func (a *App) buildHistoryTab() fyne.CanvasObject {
	a.historyPanel = container.NewVBox()
	a.refreshHistoryPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(components.MakeSectionHeader("История Заданий"), a.historyPanel),
	))
}

func (a *App) refreshHistoryPanel() {
	if a.historyPanel == nil {
		return
	}
	a.historyPanel.Objects = nil

	quests, err := a.engine.DB.GetCompletedQuests(a.engine.Character.ID, 50)
	if err != nil {
		a.historyPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		a.historyPanel.Refresh()
		return
	}

	if len(quests) == 0 {
		a.historyPanel.Add(components.MakeEmptyState("История пуста. Выполняйте задания!"))
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
	rankBadge := components.MakeRankBadge(q.Rank)
	titleText := components.MakeTitle(q.Title, components.ColorText, 14)

	completedStr := ""
	if q.CompletedAt != nil {
		completedStr = q.CompletedAt.Format("02.01.2006 15:04")
	}

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

	dateText := components.MakeLabel(completedStr, components.ColorTextDim)
	expText := components.MakeLabel(
		fmt.Sprintf("+%d EXP -> %s %s", q.Rank.BaseEXP(), q.TargetStat.Icon(), q.TargetStat.DisplayName()),
		components.ColorGreen,
	)

	topRow := container.NewHBox(rankBadge, titleText, typeIndicator, layout.NewSpacer(), dateText)
	content := container.NewVBox(topRow, expText)
	return components.MakeCard(content)
}

// ================================================================
// Statistics Tab
// ================================================================

func (a *App) buildStatsTab() fyne.CanvasObject {
	a.statsPanel = a.tabsCtx.StatsPanel
	return tabs.BuildProgress(a.tabsCtx)
}

func (a *App) refreshStatsPanel() {
	tabs.RefreshProgress(a.tabsCtx)
}

// ================================================================
// Dungeons Tab
// ================================================================

func (a *App) buildDungeonsTab() fyne.CanvasObject {
	a.dungeonsPanel = a.tabsCtx.DungeonsPanel
	return tabs.BuildDungeons(a.tabsCtx)
}

func (a *App) refreshDungeonsPanel() {
	tabs.RefreshDungeons(a.tabsCtx)
}

// ================================================================
// Arena Tab
// ================================================================

func (a *App) buildArenaTab() fyne.CanvasObject {
	a.arenaPanel = container.NewVBox()
	a.refreshArenaPanel()
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(
			components.MakeSectionHeader("Башня — Выбор противника"),
			components.MakeLabel("Бои не дают EXP. 1 бой = 1 попытка. Попытки зарабатываются заданиями.", components.ColorTextDim),
			a.arenaPanel,
		),
	))
}

func (a *App) refreshArenaPanel() {
	if a.arenaPanel == nil {
		return
	}
	a.arenaPanel.Objects = nil

	// Show attempts
	attempts := a.engine.GetAttempts()
	attemptsBar := components.MakeEXPBar(attempts, models.MaxAttempts, components.ColorAccent)
	attemptsLabel := components.MakeTitle(
		fmt.Sprintf("Попытки боя: %d / %d", attempts, models.MaxAttempts),
		components.ColorText, 16,
	)
	a.arenaPanel.Add(components.MakeCard(container.NewVBox(attemptsLabel, attemptsBar)))

	enemies, err := a.engine.GetEnemies()
	if err != nil {
		a.arenaPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		a.arenaPanel.Refresh()
		return
	}

	// Group enemies by floor
	floorMap := make(map[int][]models.Enemy)
	for _, e := range enemies {
		floorMap[e.Floor] = append(floorMap[e.Floor], e)
	}

	// Collect and sort floors
	var floors []int
	for f := range floorMap {
		floors = append(floors, f)
	}
	sortInts(floors)

	for _, floor := range floors {
		floorEnemies := floorMap[floor]
		floorLabel := game.FloorName(floor)
		a.arenaPanel.Add(widget.NewSeparator())
		a.arenaPanel.Add(components.MakeTitle(fmt.Sprintf("Этаж %d — %s", floor, floorLabel), components.ColorAccentBright, 16))

		for _, e := range floorEnemies {
			enemy := e
			card := a.buildEnemyCard(enemy)
			a.arenaPanel.Add(card)
		}
	}

	// Battle history
	battles, err := a.engine.GetBattleHistory(10)
	if err == nil && len(battles) > 0 {
		a.arenaPanel.Add(widget.NewSeparator())
		a.arenaPanel.Add(components.MakeTitle("Последние бои", components.ColorAccentBright, 16))
		for _, b := range battles {
			card := a.buildBattleHistoryCard(b)
			a.arenaPanel.Add(card)
		}
	}

	a.arenaPanel.Refresh()
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}

func (a *App) buildEnemyCard(e models.Enemy) *fyne.Container {
	rankBadge := components.MakeRankBadge(e.Rank)

	var typeLabel *canvas.Text
	if e.Type == models.EnemyBoss {
		typeLabel = components.MakeLabel("БОСС", components.ColorRed)
		typeLabel.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		typeLabel = components.MakeLabel("", components.ColorTextDim)
	}

	nameText := components.MakeTitle(e.Name, components.ColorText, 15)
	descText := components.MakeLabel(e.Description, components.ColorTextDim)

	statsText := components.MakeLabel(
		fmt.Sprintf("HP: %d  ATK: %d", e.HP, e.Attack),
		components.ColorTextDim,
	)

	rewardText := components.MakeLabel(
		"Награда: титул, бейдж, открытие следующего этажа",
		components.ColorGold,
	)

	fightBtn := widget.NewButtonWithIcon("Сражаться!", theme.MediaPlayIcon(), func() {
		a.startBattle(e)
	})
	fightBtn.Importance = widget.HighImportance

	topRow := container.NewHBox(rankBadge, nameText, typeLabel, layout.NewSpacer(), fightBtn)
	content := container.NewVBox(topRow, descText, statsText, rewardText)
	return components.MakeCard(content)
}

func (a *App) startBattle(enemy models.Enemy) {
	if enemy.Type == models.EnemyBoss {
		state, err := a.engine.StartBossBattle(enemy.ID)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		a.currentBoss = state
		a.showBossScreen()
		return
	}
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
	var cellButtons []*widget.Button
	var gridContainer *fyne.Container
	var inputSequence []int
	var inputErrors int
	var progressLabel *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		contentRef.Objects = nil

		// HP bars
		playerHPBar := components.MakeEXPBar(state.PlayerHP, state.PlayerMaxHP, components.ColorGreen)
		enemyHPBar := components.MakeEXPBar(state.EnemyHP, state.EnemyMaxHP, components.ColorRed)

		playerLabel := components.MakeTitle(fmt.Sprintf("Охотник HP: %d/%d", state.PlayerHP, state.PlayerMaxHP), components.ColorGreen, 14)
		enemyLabel := components.MakeTitle(fmt.Sprintf("%s HP: %d/%d", state.Enemy.Name, state.EnemyHP, state.EnemyMaxHP), components.ColorRed, 14)

		roundLabel := components.MakeTitle(fmt.Sprintf("Раунд %d", state.Round), components.ColorAccentBright, 16)

		contentRef.Add(container.NewHBox(roundLabel))
		contentRef.Add(container.NewVBox(playerLabel, playerHPBar))
		contentRef.Add(container.NewVBox(enemyLabel, enemyHPBar))
		contentRef.Add(widget.NewSeparator())

		if state.BattleOver {
			var resultText string
			var resultColor color.Color
			if state.Result == models.BattleWin {
				resultText = "ПОБЕДА!"
				resultColor = components.ColorGold
			} else {
				resultText = "ПОРАЖЕНИЕ"
				resultColor = components.ColorRed
			}

			resultLabel := components.MakeTitle(resultText, resultColor, 24)
			contentRef.Add(container.NewCenter(resultLabel))
			contentRef.Add(container.NewCenter(components.MakeLabel("Бои не дают EXP. Это испытание силы.", components.ColorTextDim)))

			// Finish battle and show rewards
			record, err := a.engine.FinishBattle(state)
			if err == nil && state.Result == models.BattleWin {
				if record.RewardTitle != "" {
					contentRef.Add(components.MakeLabel("Титул открыт: "+record.RewardTitle, components.ColorGold))
				}
				if record.RewardBadge != "" {
					contentRef.Add(components.MakeLabel("Бейдж получен: "+record.RewardBadge, components.ColorGold))
				}
				if record.UnlockedEnemyName != "" {
					contentRef.Add(components.MakeLabel("Открыт новый враг: "+record.UnlockedEnemyName, components.ColorAccentBright))
				}
			}

			contentRef.Add(components.MakeLabel(
				fmt.Sprintf("Точность: %.1f%% | Ошибки: %d",
					record.Accuracy, state.TotalMisses),
				components.ColorTextDim,
			))

			if hint := battleStatHint(state, record); hint != "" {
				contentRef.Add(components.MakeLabel("Подсказка: "+hint, components.ColorAccentBright))
			}

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
		infoLabel := components.MakeLabel("Запомните последовательность!", components.ColorGold)
		metaLabel := components.MakeLabel(
			fmt.Sprintf("Сетка: %dx%d | Паттерн: %d | Ошибки: %d",
				state.GridSize, state.GridSize, state.PatternLength, state.AllowedErrors),
			components.ColorTextDim,
		)
		contentRef.Add(container.NewCenter(infoLabel))
		contentRef.Add(container.NewCenter(metaLabel))

		inputSequence = nil
		inputErrors = 0
		progressLabel = components.MakeLabel(
			fmt.Sprintf("Ввод: 0/%d | Ошибки: 0/%d", state.PatternLength, state.AllowedErrors),
			components.ColorTextDim,
		)
		contentRef.Add(container.NewCenter(progressLabel))

		// Build grid
		cellCount := state.GridSize * state.GridSize
		cellButtons = make([]*widget.Button, cellCount)
		var gridCells []fyne.CanvasObject
		for i := 0; i < cellCount; i++ {
			btn := widget.NewButton("", nil)
			btn.Importance = widget.LowImportance
			btn.Disable()
			cellButtons[i] = btn
			gridCells = append(gridCells, btn)
		}

		cellSize := cellSizeForGrid(state.GridSize)
		gridContainer = container.New(layout.NewGridWrapLayout(fyne.NewSize(cellSize, cellSize)), gridCells...)
		contentRef.Add(gridContainer)

		confirmBtn = widget.NewButtonWithIcon("Подтвердить ход", theme.ConfirmIcon(), func() {
			err := a.engine.ProcessRound(state, inputSequence)
			if err != nil {
				dialog.ShowError(err, battleWindow)
				return
			}
			rebuildScreen()
		})
		confirmBtn.Importance = widget.HighImportance
		confirmBtn.Disable()

		resetBtn = widget.NewButtonWithIcon("Сбросить ввод", theme.ViewRefreshIcon(), func() {
			inputSequence = nil
			inputErrors = 0
			progressLabel.Text = fmt.Sprintf("Ввод: 0/%d | Ошибки: 0/%d", state.PatternLength, state.AllowedErrors)
			progressLabel.Refresh()
			confirmBtn.Disable()
		})
		resetBtn.Importance = widget.MediumImportance
		contentRef.Add(container.NewHBox(confirmBtn, resetBtn))
		contentRef.Refresh()

		// Show pattern, then enable input
		showTimeMs, _ := a.engine.GetShowTimeMs(state.ShowTimeMs)
		perStep := showTimeMs / len(state.Pattern)
		if perStep < 150 {
			perStep = 150
		}
		go func() {
			for _, idx := range state.Pattern {
				cellIdx := idx
				runOnMain(func() {
					cellButtons[cellIdx].Importance = widget.HighImportance
					cellButtons[cellIdx].Refresh()
				})
				time.Sleep(time.Duration(perStep) * time.Millisecond)
				runOnMain(func() {
					cellButtons[cellIdx].Importance = widget.LowImportance
					cellButtons[cellIdx].Refresh()
				})
			}

			runOnMain(func() {
				infoLabel.Text = fmt.Sprintf("Повторите %d шагов. Ошибки допустимы: %d", state.PatternLength, state.AllowedErrors)
				infoLabel.Refresh()

				for i := 0; i < cellCount; i++ {
					idx := i
					cellButtons[idx].Enable()
					cellButtons[idx].OnTapped = func() {
						if len(inputSequence) >= state.PatternLength || inputErrors > state.AllowedErrors {
							return
						}
						inputSequence = append(inputSequence, idx)
						correct := idx == state.Pattern[len(inputSequence)-1]
						if !correct {
							inputErrors++
						}

						if correct {
							cellButtons[idx].Importance = widget.MediumImportance
						} else {
							cellButtons[idx].Importance = widget.LowImportance
						}
						cellButtons[idx].Refresh()

						progressLabel.Text = fmt.Sprintf("Ввод: %d/%d | Ошибки: %d/%d", len(inputSequence), state.PatternLength, inputErrors, state.AllowedErrors)
						progressLabel.Refresh()

						if len(inputSequence) >= state.PatternLength || inputErrors > state.AllowedErrors {
							confirmBtn.Enable()
							if inputErrors > state.AllowedErrors {
								infoLabel.Text = "Лимит ошибок превышен — подтвердите ход"
								infoLabel.Refresh()
							}
						}
					}
				}
			})
		}()
	}

	contentRef = container.NewVBox()
	rebuildScreen()

	battleWindow.SetContent(container.NewVScroll(container.NewPadded(contentRef)))
	battleWindow.Show()
}

func (a *App) showBossScreen() {
	state := a.currentBoss

	battleWindow := a.app.NewWindow(fmt.Sprintf("Босс: %s", state.Enemy.Name))
	battleWindow.Resize(fyne.NewSize(740, 700))
	battleWindow.CenterOnScreen()

	var contentRef *fyne.Container
	var cellButtons []*widget.Button
	var gridContainer *fyne.Container
	var inputSequence []int
	var inputErrors int
	var progressLabel *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button
	var timerLabel *canvas.Text

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		contentRef.Objects = nil

		playerHPBar := components.MakeEXPBar(state.PlayerHP, state.PlayerMaxHP, components.ColorGreen)
		enemyHPBar := components.MakeEXPBar(state.EnemyHP, state.EnemyMaxHP, components.ColorRed)

		playerLabel := components.MakeTitle(fmt.Sprintf("Охотник HP: %d/%d", state.PlayerHP, state.PlayerMaxHP), components.ColorGreen, 14)
		enemyLabel := components.MakeTitle(fmt.Sprintf("%s HP: %d/%d", state.Enemy.Name, state.EnemyHP, state.EnemyMaxHP), components.ColorRed, 14)

		phaseLabel := components.MakeTitle(fmt.Sprintf("Фаза: %s", phaseDisplay(state.Phase)), components.ColorAccentBright, 16)

		contentRef.Add(container.NewHBox(phaseLabel))
		contentRef.Add(container.NewVBox(playerLabel, playerHPBar))
		contentRef.Add(container.NewVBox(enemyLabel, enemyHPBar))
		contentRef.Add(widget.NewSeparator())

		if state.Phase == boss.PhaseWin || state.Phase == boss.PhaseLose {
			var resultText string
			var resultColor color.Color
			if state.Phase == boss.PhaseWin {
				resultText = "ПОБЕДА НАД БОССОМ!"
				resultColor = components.ColorGold
			} else {
				resultText = "ПОРАЖЕНИЕ"
				resultColor = components.ColorRed
			}

			resultLabel := components.MakeTitle(resultText, resultColor, 24)
			contentRef.Add(container.NewCenter(resultLabel))
			contentRef.Add(container.NewCenter(components.MakeLabel("Бои не дают EXP. Это испытание силы.", components.ColorTextDim)))

			var record *models.BattleRecord
			var err error
			if state.Phase == boss.PhaseWin {
				record, err = a.engine.FinishBoss(state)
			} else {
				record, err = a.engine.FailBoss(state)
			}
			if err == nil && state.Phase == boss.PhaseWin {
				if record.RewardTitle != "" {
					contentRef.Add(components.MakeLabel("Титул открыт: "+record.RewardTitle, components.ColorGold))
				}
				if record.RewardBadge != "" {
					contentRef.Add(components.MakeLabel("Бейдж получен: "+record.RewardBadge, components.ColorGold))
				}
				if record.UnlockedEnemyName != "" {
					contentRef.Add(components.MakeLabel("Открыт новый враг: "+record.UnlockedEnemyName, components.ColorAccentBright))
				}
			}

			contentRef.Add(components.MakeLabel(
				fmt.Sprintf("Точность: %.1f%% | Ошибки: %d",
					record.Accuracy, state.TotalMisses),
				components.ColorTextDim,
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

		if state.Phase == boss.PhaseMemory {
			infoLabel := components.MakeLabel("Фаза 1: Tactical Memory", components.ColorGold)
			metaLabel := components.MakeLabel(
				fmt.Sprintf("Сетка: %dx%d | Паттерн: %d | Ошибки: %d",
					state.Memory.GridSize, state.Memory.GridSize, state.Memory.PatternLength, state.Memory.AllowedErrors),
				components.ColorTextDim,
			)
			contentRef.Add(container.NewCenter(infoLabel))
			contentRef.Add(container.NewCenter(metaLabel))

			inputSequence = nil
			inputErrors = 0
			progressLabel = components.MakeLabel(
				fmt.Sprintf("Ввод: 0/%d | Ошибки: 0/%d", state.Memory.PatternLength, state.Memory.AllowedErrors),
				components.ColorTextDim,
			)
			contentRef.Add(container.NewCenter(progressLabel))

			cellCount := state.Memory.GridSize * state.Memory.GridSize
			cellButtons = make([]*widget.Button, cellCount)
			var gridCells []fyne.CanvasObject
			for i := 0; i < cellCount; i++ {
				btn := widget.NewButton("", nil)
				btn.Importance = widget.LowImportance
				btn.Disable()
				cellButtons[i] = btn
				gridCells = append(gridCells, btn)
			}

			cellSize := cellSizeForGrid(state.Memory.GridSize)
			gridContainer = container.New(layout.NewGridWrapLayout(fyne.NewSize(cellSize, cellSize)), gridCells...)
			contentRef.Add(gridContainer)

			confirmBtn = widget.NewButtonWithIcon("Подтвердить ход", theme.ConfirmIcon(), func() {
				err := a.engine.ProcessBossMemory(state, inputSequence)
				if err != nil {
					dialog.ShowError(err, battleWindow)
					return
				}
				rebuildScreen()
			})
			confirmBtn.Importance = widget.HighImportance
			confirmBtn.Disable()

			resetBtn = widget.NewButtonWithIcon("Сбросить ввод", theme.ViewRefreshIcon(), func() {
				inputSequence = nil
				inputErrors = 0
				progressLabel.Text = fmt.Sprintf("Ввод: 0/%d | Ошибки: 0/%d", state.Memory.PatternLength, state.Memory.AllowedErrors)
				progressLabel.Refresh()
				confirmBtn.Disable()
			})
			resetBtn.Importance = widget.MediumImportance
			contentRef.Add(container.NewHBox(confirmBtn, resetBtn))
			contentRef.Refresh()

			showTimeMs, _ := a.engine.GetShowTimeMs(state.Memory.ShowTimeMs)
			perStep := showTimeMs / len(state.Memory.Pattern)
			if perStep < 150 {
				perStep = 150
			}
			go func() {
				for _, idx := range state.Memory.Pattern {
					cellIdx := idx
					runOnMain(func() {
						cellButtons[cellIdx].Importance = widget.HighImportance
						cellButtons[cellIdx].Refresh()
					})
					time.Sleep(time.Duration(perStep) * time.Millisecond)
					runOnMain(func() {
						cellButtons[cellIdx].Importance = widget.LowImportance
						cellButtons[cellIdx].Refresh()
					})
				}

				runOnMain(func() {
					infoLabel.Text = fmt.Sprintf("Повторите %d шагов. Ошибки допустимы: %d", state.Memory.PatternLength, state.Memory.AllowedErrors)
					infoLabel.Refresh()

					for i := 0; i < cellCount; i++ {
						idx := i
						cellButtons[idx].Enable()
						cellButtons[idx].OnTapped = func() {
							if len(inputSequence) >= state.Memory.PatternLength || inputErrors > state.Memory.AllowedErrors {
								return
							}
							inputSequence = append(inputSequence, idx)
							correct := idx == state.Memory.Pattern[len(inputSequence)-1]
							if !correct {
								inputErrors++
							}

							if correct {
								cellButtons[idx].Importance = widget.MediumImportance
							} else {
								cellButtons[idx].Importance = widget.LowImportance
							}
							cellButtons[idx].Refresh()

							progressLabel.Text = fmt.Sprintf("Ввод: %d/%d | Ошибки: %d/%d", len(inputSequence), state.Memory.PatternLength, inputErrors, state.Memory.AllowedErrors)
							progressLabel.Refresh()

							if len(inputSequence) >= state.Memory.PatternLength || inputErrors > state.Memory.AllowedErrors {
								confirmBtn.Enable()
								if inputErrors > state.Memory.AllowedErrors {
									infoLabel.Text = "Лимит ошибок превышен — подтвердите ход"
									infoLabel.Refresh()
								}
							}
						}
					}
				})
			}()
			return
		}

		if state.Phase == boss.PhasePressure {
			infoLabel := components.MakeLabel("Фаза 2: Pressure Puzzle", components.ColorGold)
			contentRef.Add(container.NewCenter(infoLabel))

			timerLabel = components.MakeLabel("", components.ColorRed)
			contentRef.Add(container.NewCenter(timerLabel))

			stepsLabel := components.MakeLabel(
				fmt.Sprintf("Шаги: %d | Ошибки: %d | Попытки: %d", state.Puzzle.Steps, state.Puzzle.AllowedErrors, state.Puzzle.AttemptsLeft),
				components.ColorTextDim,
			)
			contentRef.Add(container.NewCenter(stepsLabel))

			numbers := make([]int, state.Puzzle.Steps)
			for i := 0; i < state.Puzzle.Steps; i++ {
				numbers[i] = i + 1
			}
			rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })

			var buttons []fyne.CanvasObject
			for _, v := range numbers {
				val := v
				btn := widget.NewButton(fmt.Sprintf("%d", val), func() {
					err := a.engine.ProcessBossPuzzleInput(state, val)
					if err != nil {
						dialog.ShowError(err, battleWindow)
						return
					}
					stepsLabel.Text = fmt.Sprintf("Шаги: %d | Ошибки: %d | Попытки: %d", state.Puzzle.Steps, state.Puzzle.AllowedErrors, state.Puzzle.AttemptsLeft)
					stepsLabel.Refresh()
					if state.Phase == boss.PhaseWin || state.Phase == boss.PhaseLose {
						rebuildScreen()
					}
				})
				btn.Importance = widget.HighImportance
				buttons = append(buttons, btn)
			}

			gridSize := 3
			if state.Puzzle.Steps >= 7 {
				gridSize = 4
			}
			cellSize := float32(70)
			grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(cellSize, cellSize)), buttons...)
			contentRef.Add(container.NewCenter(grid))

			contentRef.Refresh()

			go func() {
				deadline := time.Now().Add(time.Duration(state.Puzzle.TimeLimitMs) * time.Millisecond)
				for {
					if state.Phase != boss.PhasePressure {
						return
					}
					remaining := time.Until(deadline)
					if remaining <= 0 {
						boss.PressureTimedOut(state)
						runOnMain(rebuildScreen)
						return
					}
					runOnMain(func() {
						timerLabel.Text = fmt.Sprintf("Осталось: %.1fс", remaining.Seconds())
						timerLabel.Refresh()
					})
					time.Sleep(100 * time.Millisecond)
				}
			}()
			_ = gridSize
			return
		}
	}

	contentRef = container.NewVBox()
	rebuildScreen()

	battleWindow.SetContent(container.NewVScroll(container.NewPadded(contentRef)))
	battleWindow.Show()
}

func phaseDisplay(p boss.Phase) string {
	switch p {
	case boss.PhaseMemory:
		return "Tactical Memory"
	case boss.PhasePressure:
		return "Pressure Puzzle"
	case boss.PhaseWin:
		return "Победа"
	case boss.PhaseLose:
		return "Поражение"
	default:
		return string(p)
	}
}

func (a *App) buildBattleHistoryCard(b models.BattleRecord) *fyne.Container {
	var resultText string
	var resultColor color.Color
	if b.Result == models.BattleWin {
		resultText = "Победа"
		resultColor = components.ColorGreen
	} else {
		resultText = "Поражение"
		resultColor = components.ColorRed
	}

	nameText := components.MakeTitle(b.EnemyName, components.ColorText, 14)
	result := components.MakeLabel(resultText, resultColor)
	result.TextStyle = fyne.TextStyle{Bold: true}
	dateText := components.MakeLabel(b.FoughtAt.Format("02.01.2006 15:04"), components.ColorTextDim)

	statsText := components.MakeLabel(
		fmt.Sprintf("Урон: %d | Точность: %.0f%% | Криты: %d", b.DamageDealt, b.Accuracy, b.CriticalHits),
		components.ColorTextDim,
	)

	topRow := container.NewHBox(nameText, result, layout.NewSpacer(), dateText)
	content := container.NewVBox(topRow, statsText)
	return components.MakeCard(content)
}

func battleStatHint(state *models.BattleState, record *models.BattleRecord) string {
	var hints []string
	if record.Accuracy < 60 {
		hints = append(hints, "Интеллект (больше времени и короче паттерн)")
	}
	if state.TotalMisses > state.AllowedErrors {
		hints = append(hints, "Ловкость (доп. ошибка)")
	}
	if state.DamageDealt < state.Enemy.HP/2 {
		hints = append(hints, "Сила (меньше шагов до урона)")
	}
	if state.PlayerHP == 0 {
		hints = append(hints, "Выносливость (больше HP)")
	}
	if len(hints) == 0 {
		return ""
	}
	if len(hints) > 2 {
		hints = hints[:2]
	}
	return strings.Join(hints, ", ")
}

func cellSizeForGrid(grid int) float32 {
	switch grid {
	case 3:
		return 90
	case 4:
		return 70
	case 5:
		return 58
	case 6:
		return 50
	default:
		return 55
	}
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
