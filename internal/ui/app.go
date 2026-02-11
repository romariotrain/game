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
		RefreshAchievements: func() {
			a.refreshAchievementsPanel()
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
		achievementsTab := container.NewTabItem("Achievements", tabs.BuildAchievements(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		tabItems := []*container.TabItem{todayTab, questsTab, progressTab, achievementsTab, dungeonsTab}
		if a.features.Combat {
			tabItems = append(tabItems, container.NewTabItem("Tower", a.buildArenaTab()))
		}
		appTabs = container.NewAppTabs(tabItems...)
	} else {
		charTab := container.NewTabItem("Охотник", tabs.BuildToday(a.tabsCtx))
		questsTab := container.NewTabItem("Задания", tabs.BuildQuests(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		statsTab := container.NewTabItem("Статистика", tabs.BuildProgress(a.tabsCtx))
		achievementsTab := container.NewTabItem("Achievements", tabs.BuildAchievements(a.tabsCtx))

		tabItems := []*container.TabItem{charTab, questsTab, dungeonsTab, statsTab, achievementsTab}
		if a.features.Combat {
			tabItems = append(tabItems, container.NewTabItem("Tower", a.buildArenaTab()))
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
	a.refreshAchievementsPanel()
	a.refreshDungeonsPanel()
	a.refreshArenaPanel()
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
		fmt.Sprintf("+%d EXP -> %s %s | Ранг: %s", q.Exp, q.TargetStat.Icon(), q.TargetStat.DisplayName(), q.Rank),
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

func (a *App) refreshAchievementsPanel() {
	tabs.RefreshAchievements(a.tabsCtx)
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
			components.MakeSectionHeader("Tower"),
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

	currentEnemy, err := a.engine.GetCurrentEnemy()
	if err != nil {
		a.arenaPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		a.arenaPanel.Refresh()
		return
	}
	a.arenaPanel.Add(widget.NewSeparator())
	if currentEnemy == nil {
		a.arenaPanel.Add(components.MakeCard(container.NewVBox(
			components.MakeTitle("Tower: зачищена", components.ColorGold, 18),
			components.MakeLabel("Статус: cleared", components.ColorAccentBright),
			components.MakeLabel("Все текущие враги побеждены. EXP за бои не начисляется.", components.ColorTextDim),
		)))
		a.arenaPanel.Refresh()
		return
	}

	section := components.MakeTitle(
		fmt.Sprintf("Текущий сектор: Этаж %d — %s", currentEnemy.Floor, game.FloorName(currentEnemy.Floor)),
		components.ColorAccentBright, 16,
	)
	a.arenaPanel.Add(section)
	a.arenaPanel.Add(a.buildEnemyCard(*currentEnemy))

	a.arenaPanel.Refresh()
}

func (a *App) buildEnemyCard(e models.Enemy) *fyne.Container {
	rankBadge := components.MakeRankBadge(e.Rank)

	var typeLabel *canvas.Text
	if e.Type == models.EnemyBoss {
		typeLabel = components.MakeLabel("boss", components.ColorRed)
		typeLabel.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		typeLabel = components.MakeLabel("regular", components.ColorTextDim)
	}

	nameText := components.MakeTitle(e.Name, components.ColorText, 15)
	descText := components.MakeLabel(e.Description, components.ColorTextDim)
	statusText := components.MakeLabel("Статус: unlocked/current", components.ColorAccentBright)

	statsText := components.MakeLabel(
		fmt.Sprintf("HP: %d  ATK: %d", e.HP, e.Attack),
		components.ColorTextDim,
	)

	rewardText := components.MakeLabel(
		"Награда: титул, бейдж, открытие следующего этажа",
		components.ColorGold,
	)

	canFight, canFightErr := a.engine.CanFightCurrentEnemy(e.ID)
	fightBtn := widget.NewButtonWithIcon("В бой", theme.MediaPlayIcon(), func() {
		a.startBattle(e)
	})
	fightBtn.Importance = widget.HighImportance
	if !canFight || canFightErr != nil {
		fightBtn.Disable()
	}

	var hintText fyne.CanvasObject = layout.NewSpacer()
	if canFightErr != nil {
		hintText = components.MakeLabel("Ошибка проверки боя: "+canFightErr.Error(), components.ColorRed)
	} else if a.engine.GetAttempts() <= 0 {
		hintText = components.MakeLabel("Нет попыток — выполни квест, чтобы получить попытки.", components.ColorRed)
	}

	topRow := container.NewHBox(rankBadge, nameText, typeLabel, layout.NewSpacer(), fightBtn)
	content := container.NewVBox(topRow, descText, statsText, statusText, rewardText, hintText)
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
	battleWindow.Resize(fyne.NewSize(760, 760))
	battleWindow.CenterOnScreen()

	topRef := container.NewVBox()
	rightRef := container.NewVBox()
	centerRef := container.NewCenter()
	bottomRef := container.NewVBox()

	var cells []*battleCell
	var inputSequence []int
	var inputErrors int
	var primaryStatus *canvas.Text
	var secondaryStatus *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button
	var resolved bool
	var resolvedRecord *models.BattleRecord
	var resolvedErr error

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		topRef.Objects = nil
		rightRef.Objects = nil
		centerRef.Objects = nil
		bottomRef.Objects = nil

		header := components.MakeTitle(
			fmt.Sprintf("Бой: %s • Раунд %d", state.Enemy.Name, state.Round),
			components.ColorAccentBright, 18,
		)
		topRef.Add(header)
		topRef.Add(container.NewGridWithColumns(
			3,
			buildBattleHPRow("Охотник", state.PlayerHP, state.PlayerMaxHP, components.ColorGreen),
			buildBattleHPRow(state.Enemy.Name, state.EnemyHP, state.EnemyMaxHP, components.ColorRed),
			buildBattleAttemptsBox(a.engine.GetAttempts()),
		))

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

			if !resolved {
				resolvedRecord, resolvedErr = a.engine.FinishBattle(state)
				resolved = true
			}

			resultCard := components.MakeCard(container.NewVBox(
				container.NewCenter(components.MakeTitle(resultText, resultColor, 24)),
				container.NewCenter(components.MakeLabel("Бои не дают EXP. Это испытание силы.", components.ColorTextDim)),
			))
			centerRef.Add(container.NewCenter(resultCard))

			sideItems := []fyne.CanvasObject{
				components.MakeTitle("Итог боя", components.ColorAccentBright, 15),
			}
			if resolvedErr != nil {
				sideItems = append(sideItems, components.MakeLabel("Ошибка завершения боя: "+resolvedErr.Error(), components.ColorRed))
			} else if state.Result == models.BattleWin && resolvedRecord != nil {
				if resolvedRecord.RewardTitle != "" {
					sideItems = append(sideItems, components.MakeLabel("Титул: "+resolvedRecord.RewardTitle, components.ColorGold))
				}
				if resolvedRecord.RewardBadge != "" {
					sideItems = append(sideItems, components.MakeLabel("Бейдж: "+resolvedRecord.RewardBadge, components.ColorGold))
				}
				if resolvedRecord.UnlockedEnemyName != "" {
					sideItems = append(sideItems, components.MakeLabel("Открыт: "+resolvedRecord.UnlockedEnemyName, components.ColorAccentBright))
				}
			} else {
				sideItems = append(sideItems, components.MakeLabel("Поражение не наказывается. Попробуйте позже после квестов.", components.ColorTextDim))
			}
			if resolvedRecord != nil {
				sideItems = append(sideItems,
					components.MakeLabel(
						fmt.Sprintf("Точность: %.1f%%", resolvedRecord.Accuracy),
						components.ColorTextDim,
					),
					components.MakeLabel(
						fmt.Sprintf("Ошибки: %d", state.TotalMisses),
						components.ColorTextDim,
					),
				)
				if hint := battleStatHint(state, resolvedRecord); hint != "" {
					sideItems = append(sideItems, components.MakeLabel("Подсказка: "+hint, components.ColorAccentBright))
				}
			}
			rightRef.Add(components.MakeCard(container.NewVBox(sideItems...)))

			nextLabel := "Закрыть"
			nextIcon := theme.CancelIcon()
			if state.Result == models.BattleWin {
				nextLabel = "Дальше"
				nextIcon = theme.NavigateNextIcon()
			}
			closeBtn := widget.NewButtonWithIcon(nextLabel, nextIcon, func() {
				battleWindow.Close()
				a.refreshArenaPanel()
				a.refreshCharacterPanel()
				a.refreshStatsPanel()
			})
			closeBtn.Importance = widget.HighImportance
			bottomRef.Add(container.NewHBox(layout.NewSpacer(), closeBtn))

			topRef.Refresh()
			rightRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()
			return
		}

		inputSequence = nil
		inputErrors = 0
		primaryStatus = components.MakeLabel("Запомни последовательность", components.ColorGold)
		secondaryStatus = components.MakeLabel(
			fmt.Sprintf("Сетка %dx%d • Ввод 0/%d", state.GridSize, state.GridSize, state.PatternLength),
			components.ColorTextDim,
		)
		secondaryStatus.TextSize = 12

		cellCount := state.GridSize * state.GridSize
		cells = make([]*battleCell, cellCount)
		var gridCells []fyne.CanvasObject
		cellSize := cellSizeForGrid(state.GridSize)
		for i := 0; i < cellCount; i++ {
			cell := newBattleCell(cellSize)
			cell.Disable()
			cell.SetState(battleCellStateIdle)
			cells[i] = cell
			gridCells = append(gridCells, cell)
		}

		gridContainer := container.NewGridWithColumns(state.GridSize, gridCells...)
		fieldCard := components.MakeCard(container.NewPadded(gridContainer))
		centerContent := container.NewVBox(
			container.NewCenter(fieldCard),
			container.NewCenter(primaryStatus),
			container.NewCenter(secondaryStatus),
		)
		centerRef.Add(container.NewCenter(centerContent))

		rightRef.Add(components.MakeCard(container.NewVBox(
			components.MakeTitle("Параметры", components.ColorAccentBright, 15),
			components.MakeLabel(fmt.Sprintf("Враг: %s", state.Enemy.Name), components.ColorText),
			components.MakeLabel(fmt.Sprintf("Ранг: %s", state.Enemy.Rank), components.ColorTextDim),
			components.MakeLabel(fmt.Sprintf("Сетка: %dx%d", state.GridSize, state.GridSize), components.ColorTextDim),
			components.MakeLabel(fmt.Sprintf("Паттерн: %d", state.PatternLength), components.ColorTextDim),
			components.MakeLabel(fmt.Sprintf("Ошибок можно: %d", state.AllowedErrors), components.ColorTextDim),
			components.MakeLabel("Бои не дают EXP.", components.ColorTextDim),
		)))

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
			primaryStatus.Text = fmt.Sprintf("Повтори %d шагов • Ошибок можно: %d", state.PatternLength, state.AllowedErrors)
			primaryStatus.Color = components.ColorText
			primaryStatus.Refresh()
			secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Ввод 0/%d", state.GridSize, state.GridSize, state.PatternLength)
			secondaryStatus.Refresh()
			for _, c := range cells {
				c.SetState(battleCellStateIdle)
			}
			confirmBtn.Disable()
		})
		resetBtn.Importance = widget.MediumImportance
		bottomRef.Add(container.NewHBox(confirmBtn, resetBtn))

		topRef.Refresh()
		rightRef.Refresh()
		centerRef.Refresh()
		bottomRef.Refresh()

		showTimeMs, _ := a.engine.GetShowTimeMs(state.ShowTimeMs)
		perStep := showTimeMs / len(state.Pattern)
		if perStep < 150 {
			perStep = 150
		}
		go func() {
			for _, idx := range state.Pattern {
				cellIdx := idx
				runOnMain(func() {
					cells[cellIdx].SetState(battleCellStateShowing)
				})
				time.Sleep(time.Duration(perStep) * time.Millisecond)
				runOnMain(func() {
					cells[cellIdx].SetState(battleCellStateIdle)
				})
			}

			runOnMain(func() {
				primaryStatus.Text = fmt.Sprintf("Повтори %d шагов • Ошибок можно: %d", state.PatternLength, state.AllowedErrors)
				primaryStatus.Color = components.ColorText
				primaryStatus.Refresh()

				for i := 0; i < cellCount; i++ {
					idx := i
					cells[idx].Enable()
					cells[idx].SetOnTapped(func() {
						if len(inputSequence) >= state.PatternLength || inputErrors > state.AllowedErrors {
							return
						}
						inputSequence = append(inputSequence, idx)
						correct := idx == state.Pattern[len(inputSequence)-1]
						if !correct {
							inputErrors++
							cells[idx].SetState(battleCellStateError)
						} else {
							cells[idx].SetState(battleCellStateSelected)
						}

						secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Ввод %d/%d", state.GridSize, state.GridSize, len(inputSequence), state.PatternLength)
						secondaryStatus.Refresh()

						if len(inputSequence) >= state.PatternLength || inputErrors > state.AllowedErrors {
							confirmBtn.Enable()
							if inputErrors > state.AllowedErrors {
								primaryStatus.Text = "Лимит ошибок превышен • Нажми Подтвердить"
								primaryStatus.Color = components.ColorRed
								primaryStatus.Refresh()
							}
						}
					})
				}
			})
		}()
	}

	root := container.NewBorder(topRef, bottomRef, nil, rightRef, centerRef)
	battleWindow.SetContent(container.NewPadded(root))
	rebuildScreen()
	battleWindow.Show()
}

func (a *App) showBossScreen() {
	state := a.currentBoss

	battleWindow := a.app.NewWindow(fmt.Sprintf("Босс: %s", state.Enemy.Name))
	battleWindow.Resize(fyne.NewSize(760, 760))
	battleWindow.CenterOnScreen()

	topRef := container.NewVBox()
	rightRef := container.NewVBox()
	centerRef := container.NewCenter()
	bottomRef := container.NewVBox()

	var cells []*battleCell
	var inputSequence []int
	var inputErrors int
	var primaryStatus *canvas.Text
	var secondaryStatus *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button
	var timerLabel *canvas.Text
	var resolved bool
	var resolvedRecord *models.BattleRecord
	var resolvedErr error

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		topRef.Objects = nil
		rightRef.Objects = nil
		centerRef.Objects = nil
		bottomRef.Objects = nil

		header := components.MakeTitle(
			fmt.Sprintf("Бой: %s • %s", state.Enemy.Name, phaseDisplay(state.Phase)),
			components.ColorAccentBright, 18,
		)
		topRef.Add(header)
		topRef.Add(container.NewGridWithColumns(
			3,
			buildBattleHPRow("Охотник", state.PlayerHP, state.PlayerMaxHP, components.ColorGreen),
			buildBattleHPRow(state.Enemy.Name, state.EnemyHP, state.EnemyMaxHP, components.ColorRed),
			buildBattleAttemptsBox(a.engine.GetAttempts()),
		))

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

			if !resolved {
				if state.Phase == boss.PhaseWin {
					resolvedRecord, resolvedErr = a.engine.FinishBoss(state)
				} else {
					resolvedRecord, resolvedErr = a.engine.FailBoss(state)
				}
				resolved = true
			}

			centerRef.Add(container.NewCenter(components.MakeCard(container.NewVBox(
				container.NewCenter(components.MakeTitle(resultText, resultColor, 24)),
				container.NewCenter(components.MakeLabel("Бои не дают EXP. Это испытание силы.", components.ColorTextDim)),
			))))

			sideItems := []fyne.CanvasObject{
				components.MakeTitle("Итог боя", components.ColorAccentBright, 15),
			}
			if resolvedErr != nil {
				sideItems = append(sideItems, components.MakeLabel("Ошибка завершения боя: "+resolvedErr.Error(), components.ColorRed))
			} else if state.Phase == boss.PhaseWin && resolvedRecord != nil {
				if resolvedRecord.RewardTitle != "" {
					sideItems = append(sideItems, components.MakeLabel("Титул: "+resolvedRecord.RewardTitle, components.ColorGold))
				}
				if resolvedRecord.RewardBadge != "" {
					sideItems = append(sideItems, components.MakeLabel("Бейдж: "+resolvedRecord.RewardBadge, components.ColorGold))
				}
				if resolvedRecord.UnlockedEnemyName != "" {
					sideItems = append(sideItems, components.MakeLabel("Открыт: "+resolvedRecord.UnlockedEnemyName, components.ColorAccentBright))
				}
			} else {
				sideItems = append(sideItems, components.MakeLabel("Поражение не наказывается. Попробуйте позже после квестов.", components.ColorTextDim))
			}
			if resolvedRecord != nil {
				sideItems = append(sideItems,
					components.MakeLabel(fmt.Sprintf("Точность: %.1f%%", resolvedRecord.Accuracy), components.ColorTextDim),
					components.MakeLabel(fmt.Sprintf("Ошибки: %d", state.TotalMisses), components.ColorTextDim),
				)
			}
			rightRef.Add(components.MakeCard(container.NewVBox(sideItems...)))

			nextLabel := "Закрыть"
			nextIcon := theme.CancelIcon()
			if state.Phase == boss.PhaseWin {
				nextLabel = "Дальше"
				nextIcon = theme.NavigateNextIcon()
			}
			closeBtn := widget.NewButtonWithIcon(nextLabel, nextIcon, func() {
				battleWindow.Close()
				a.refreshArenaPanel()
				a.refreshCharacterPanel()
				a.refreshStatsPanel()
			})
			closeBtn.Importance = widget.HighImportance
			bottomRef.Add(container.NewHBox(layout.NewSpacer(), closeBtn))

			topRef.Refresh()
			rightRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()
			return
		}

		if state.Phase == boss.PhaseMemory {
			inputSequence = nil
			inputErrors = 0
			primaryStatus = components.MakeLabel(
				fmt.Sprintf("Повтори %d шагов • Ошибок можно: %d", state.Memory.PatternLength, state.Memory.AllowedErrors),
				components.ColorText,
			)
			secondaryStatus = components.MakeLabel(
				fmt.Sprintf("Сетка %dx%d • Ввод 0/%d", state.Memory.GridSize, state.Memory.GridSize, state.Memory.PatternLength),
				components.ColorTextDim,
			)
			secondaryStatus.TextSize = 12

			cellCount := state.Memory.GridSize * state.Memory.GridSize
			cells = make([]*battleCell, cellCount)
			var gridCells []fyne.CanvasObject
			cellSize := cellSizeForGrid(state.Memory.GridSize)
			for i := 0; i < cellCount; i++ {
				cell := newBattleCell(cellSize)
				cell.Disable()
				cell.SetState(battleCellStateIdle)
				cells[i] = cell
				gridCells = append(gridCells, cell)
			}

			gridContainer := container.NewGridWithColumns(state.Memory.GridSize, gridCells...)
			fieldCard := components.MakeCard(container.NewPadded(gridContainer))
			centerRef.Add(container.NewCenter(container.NewVBox(
				container.NewCenter(fieldCard),
				container.NewCenter(primaryStatus),
				container.NewCenter(secondaryStatus),
			)))

			rightRef.Add(components.MakeCard(container.NewVBox(
				components.MakeTitle("Фаза 1", components.ColorAccentBright, 15),
				components.MakeLabel("Tactical Memory", components.ColorGold),
				components.MakeLabel(fmt.Sprintf("Сетка: %dx%d", state.Memory.GridSize, state.Memory.GridSize), components.ColorTextDim),
				components.MakeLabel(fmt.Sprintf("Паттерн: %d", state.Memory.PatternLength), components.ColorTextDim),
				components.MakeLabel(fmt.Sprintf("Ошибок можно: %d", state.Memory.AllowedErrors), components.ColorTextDim),
			)))

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
				primaryStatus.Text = fmt.Sprintf("Повтори %d шагов • Ошибок можно: %d", state.Memory.PatternLength, state.Memory.AllowedErrors)
				primaryStatus.Color = components.ColorText
				primaryStatus.Refresh()
				secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Ввод 0/%d", state.Memory.GridSize, state.Memory.GridSize, state.Memory.PatternLength)
				secondaryStatus.Refresh()
				for _, c := range cells {
					c.SetState(battleCellStateIdle)
				}
				confirmBtn.Disable()
			})
			resetBtn.Importance = widget.MediumImportance
			bottomRef.Add(container.NewHBox(confirmBtn, resetBtn))

			topRef.Refresh()
			rightRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()

			showTimeMs, _ := a.engine.GetShowTimeMs(state.Memory.ShowTimeMs)
			perStep := showTimeMs / len(state.Memory.Pattern)
			if perStep < 150 {
				perStep = 150
			}
			go func() {
				for _, idx := range state.Memory.Pattern {
					cellIdx := idx
					runOnMain(func() {
						cells[cellIdx].SetState(battleCellStateShowing)
					})
					time.Sleep(time.Duration(perStep) * time.Millisecond)
					runOnMain(func() {
						cells[cellIdx].SetState(battleCellStateIdle)
					})
				}

				runOnMain(func() {
					primaryStatus.Text = fmt.Sprintf("Повтори %d шагов • Ошибок можно: %d", state.Memory.PatternLength, state.Memory.AllowedErrors)
					primaryStatus.Color = components.ColorText
					primaryStatus.Refresh()

					for i := 0; i < cellCount; i++ {
						idx := i
						cells[idx].Enable()
						cells[idx].SetOnTapped(func() {
							if len(inputSequence) >= state.Memory.PatternLength || inputErrors > state.Memory.AllowedErrors {
								return
							}
							inputSequence = append(inputSequence, idx)
							correct := idx == state.Memory.Pattern[len(inputSequence)-1]
							if !correct {
								inputErrors++
								cells[idx].SetState(battleCellStateError)
							} else {
								cells[idx].SetState(battleCellStateSelected)
							}

							secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Ввод %d/%d", state.Memory.GridSize, state.Memory.GridSize, len(inputSequence), state.Memory.PatternLength)
							secondaryStatus.Refresh()

							if len(inputSequence) >= state.Memory.PatternLength || inputErrors > state.Memory.AllowedErrors {
								confirmBtn.Enable()
								if inputErrors > state.Memory.AllowedErrors {
									primaryStatus.Text = "Лимит ошибок превышен • Нажми Подтвердить"
									primaryStatus.Color = components.ColorRed
									primaryStatus.Refresh()
								}
							}
						})
					}
				})
			}()
			return
		}

		if state.Phase == boss.PhasePressure {
			timerLabel = components.MakeLabel("", components.ColorRed)

			stepsLabel := components.MakeLabel(
				fmt.Sprintf("Шаги: %d | Ошибки: %d | Попытки: %d", state.Puzzle.Steps, state.Puzzle.AllowedErrors, state.Puzzle.AttemptsLeft),
				components.ColorTextDim,
			)

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

			cols := 3
			if state.Puzzle.Steps > 9 {
				cols = 4
			}
			grid := container.NewGridWithColumns(cols, buttons...)
			centerRef.Add(container.NewCenter(components.MakeCard(container.NewPadded(container.NewVBox(
				container.NewCenter(components.MakeLabel("Фаза 2: Pressure Puzzle", components.ColorGold)),
				container.NewCenter(timerLabel),
				container.NewCenter(stepsLabel),
				container.NewCenter(grid),
			)))))

			rightRef.Add(components.MakeCard(container.NewVBox(
				components.MakeTitle("Фаза 2", components.ColorAccentBright, 15),
				components.MakeLabel("Pressure Puzzle", components.ColorGold),
				components.MakeLabel("Быстро нажимайте числа по порядку.", components.ColorTextDim),
			)))

			topRef.Refresh()
			rightRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()

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
			return
		}
	}

	root := container.NewBorder(topRef, bottomRef, nil, rightRef, centerRef)
	battleWindow.SetContent(container.NewPadded(root))
	rebuildScreen()
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
		return 96
	case 4:
		return 78
	case 5:
		return 64
	case 6:
		return 54
	default:
		return 60
	}
}

func buildBattleHPRow(name string, current, max int, fillColor color.Color) fyne.CanvasObject {
	label := components.MakeLabel(fmt.Sprintf("%s %d/%d", name, current, max), components.ColorText)
	label.TextSize = 13
	bar := makeBattleMiniHPBar(current, max, fillColor)
	return container.NewVBox(label, bar)
}

func buildBattleAttemptsBox(attempts int) fyne.CanvasObject {
	label := components.MakeLabel(fmt.Sprintf("Попытки %d/%d", attempts, models.MaxAttempts), components.ColorText)
	label.TextSize = 13
	bar := makeBattleMiniHPBar(attempts, models.MaxAttempts, components.ColorAccentBright)
	return container.NewVBox(label, bar)
}

func makeBattleMiniHPBar(current, max int, fillColor color.Color) *fyne.Container {
	ratio := 0.0
	if max > 0 {
		ratio = float64(current) / float64(max)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	bg := canvas.NewRectangle(color.NRGBA{R: 24, G: 22, B: 40, A: 255})
	bg.CornerRadius = 6
	bg.SetMinSize(fyne.NewSize(180, 12))
	bg.StrokeWidth = 1
	bg.StrokeColor = color.NRGBA{R: 70, G: 64, B: 102, A: 180}

	fill := canvas.NewRectangle(fillColor)
	fill.CornerRadius = 6

	return container.NewStack(
		bg,
		container.New(&battleBarLayout{ratio: ratio}, fill),
	)
}

type battleBarLayout struct {
	ratio float64
}

func (p *battleBarLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(180, 12)
}

func (p *battleBarLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, obj := range objects {
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(fyne.NewSize(containerSize.Width*float32(p.ratio), containerSize.Height))
	}
}

type battleCellState int

const (
	battleCellStateIdle battleCellState = iota
	battleCellStateShowing
	battleCellStateSelected
	battleCellStateError
)

type battleCell struct {
	widget.BaseWidget

	minSize  fyne.Size
	state    battleCellState
	disabled bool
	onTapped func()
}

func newBattleCell(side float32) *battleCell {
	if side < 44 {
		side = 44
	}
	c := &battleCell{
		minSize:  fyne.NewSize(side, side),
		state:    battleCellStateIdle,
		disabled: true,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *battleCell) CreateRenderer() fyne.WidgetRenderer {
	glow := canvas.NewRectangle(color.Transparent)
	glow.CornerRadius = 10

	fill := canvas.NewRectangle(color.Transparent)
	fill.CornerRadius = 8

	r := &battleCellRenderer{
		cell:    c,
		glow:    glow,
		fill:    fill,
		objects: []fyne.CanvasObject{glow, fill},
	}
	r.Refresh()
	return r
}

func (c *battleCell) SetOnTapped(fn func()) {
	c.onTapped = fn
}

func (c *battleCell) SetState(state battleCellState) {
	c.state = state
	c.Refresh()
}

func (c *battleCell) Enable() {
	c.disabled = false
	c.Refresh()
}

func (c *battleCell) Disable() {
	c.disabled = true
	c.Refresh()
}

func (c *battleCell) Tapped(_ *fyne.PointEvent) {
	if c.disabled || c.onTapped == nil {
		return
	}
	c.onTapped()
}

func (c *battleCell) TappedSecondary(_ *fyne.PointEvent) {}

type battleCellRenderer struct {
	cell    *battleCell
	glow    *canvas.Rectangle
	fill    *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *battleCellRenderer) Layout(size fyne.Size) {
	const gap = float32(4)
	const glowGap = float32(2)

	glowSize := fyne.NewSize(maxFloat32(size.Width-2*glowGap, 0), maxFloat32(size.Height-2*glowGap, 0))
	r.glow.Move(fyne.NewPos(glowGap, glowGap))
	r.glow.Resize(glowSize)

	fillSize := fyne.NewSize(maxFloat32(size.Width-2*gap, 0), maxFloat32(size.Height-2*gap, 0))
	r.fill.Move(fyne.NewPos(gap, gap))
	r.fill.Resize(fillSize)
}

func (r *battleCellRenderer) MinSize() fyne.Size {
	return r.cell.minSize
}

func (r *battleCellRenderer) Refresh() {
	fillClr, borderClr, glowClr := battleCellPalette(r.cell.state, r.cell.disabled)

	r.fill.FillColor = fillClr
	r.fill.StrokeColor = borderClr
	r.fill.StrokeWidth = 1.4

	r.glow.FillColor = color.Transparent
	r.glow.StrokeColor = glowClr
	if glowClr.A > 0 {
		r.glow.StrokeWidth = 2.0
	} else {
		r.glow.StrokeWidth = 0
	}

	r.fill.Refresh()
	r.glow.Refresh()
}

func (r *battleCellRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *battleCellRenderer) Destroy() {}

func (r *battleCellRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func battleCellPalette(state battleCellState, disabled bool) (color.NRGBA, color.NRGBA, color.NRGBA) {
	type palette struct {
		fill   color.NRGBA
		border color.NRGBA
		glow   color.NRGBA
	}

	idle := palette{
		fill:   color.NRGBA{R: 38, G: 35, B: 58, A: 220},
		border: color.NRGBA{R: 88, G: 82, B: 128, A: 220},
		glow:   color.NRGBA{R: 0, G: 0, B: 0, A: 0},
	}
	showing := palette{
		fill:   color.NRGBA{R: 118, G: 94, B: 255, A: 235},
		border: color.NRGBA{R: 170, G: 149, B: 255, A: 255},
		glow:   color.NRGBA{R: 152, G: 128, B: 255, A: 170},
	}
	selected := palette{
		fill:   color.NRGBA{R: 82, G: 126, B: 228, A: 235},
		border: color.NRGBA{R: 126, G: 165, B: 255, A: 255},
		glow:   color.NRGBA{R: 110, G: 148, B: 255, A: 140},
	}
	err := palette{
		fill:   color.NRGBA{R: 138, G: 46, B: 56, A: 245},
		border: color.NRGBA{R: 214, G: 84, B: 96, A: 255},
		glow:   color.NRGBA{R: 214, G: 84, B: 96, A: 160},
	}

	current := idle
	switch state {
	case battleCellStateShowing:
		current = showing
	case battleCellStateSelected:
		current = selected
	case battleCellStateError:
		current = err
	}

	if disabled {
		current.fill.A = dimAlpha(current.fill.A, 45)
		current.border.A = dimAlpha(current.border.A, 65)
		current.glow.A = dimAlpha(current.glow.A, 85)
	}

	return current.fill, current.border, current.glow
}

func dimAlpha(src uint8, delta uint8) uint8 {
	if src <= delta {
		return 0
	}
	return src - delta
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
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
