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

	"solo-leveling/internal/config"
	"solo-leveling/internal/game"
	"solo-leveling/internal/game/combat/boss"
	"solo-leveling/internal/game/combat/memory"
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
		StartBattle: func(enemy models.Enemy) {
			a.startBattle(enemy)
		},
	}

	if components.T().HeaderUppercase {
		a.tabsCtx.QuestThemeMode = "System"
	} else {
		a.tabsCtx.QuestThemeMode = "Classic"
	}

	a.window.SetMainMenu(a.buildMainMenu())

	content := a.buildMainLayout()
	a.window.SetContent(content)
	a.window.ShowAndRun()
}

func (a *App) buildMainMenu() *fyne.MainMenu {
	systemTheme := fyne.NewMenuItem("Theme: System HUD", func() {
		a.applyVisualTheme(true)
	})
	classicTheme := fyne.NewMenuItem("Theme: Classic", func() {
		a.applyVisualTheme(false)
	})
	viewMenu := fyne.NewMenu("Вид", systemTheme, classicTheme)
	return fyne.NewMainMenu(viewMenu)
}

func (a *App) applyVisualTheme(system bool) {
	if system {
		components.SetTheme(&components.SystemTheme)
		if a.tabsCtx != nil {
			a.tabsCtx.QuestThemeMode = "System"
		}
	} else {
		components.SetTheme(&components.ClassicTheme)
		if a.tabsCtx != nil {
			a.tabsCtx.QuestThemeMode = "Classic"
		}
	}
	components.SyncLegacyColors()
	a.app.Settings().SetTheme(&SoloLevelingTheme{})
	if a.window != nil {
		a.window.SetContent(a.buildMainLayout())
	}
}

func (a *App) buildMainLayout() fyne.CanvasObject {
	header := a.buildHeader()

	var appTabs *container.AppTabs
	if a.features.MinimalMode {
		todayTab := container.NewTabItem("Сегодня", tabs.BuildToday(a.tabsCtx))
		questsTab := container.NewTabItem("Задания", tabs.BuildQuests(a.tabsCtx))
		progressTab := container.NewTabItem("Прогресс", tabs.BuildProgress(a.tabsCtx))
		achievementsTab := container.NewTabItem("Достижения", tabs.BuildAchievements(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		tabItems := []*container.TabItem{todayTab, questsTab, progressTab, achievementsTab, dungeonsTab}
		appTabs = container.NewAppTabs(tabItems...)
	} else {
		charTab := container.NewTabItem("Охотник", tabs.BuildToday(a.tabsCtx))
		questsTab := container.NewTabItem("Задания", tabs.BuildQuests(a.tabsCtx))
		dungeonsTab := container.NewTabItem("Данжи", tabs.BuildDungeons(a.tabsCtx))
		statsTab := container.NewTabItem("Статистика", tabs.BuildProgress(a.tabsCtx))
		achievementsTab := container.NewTabItem("Достижения", tabs.BuildAchievements(a.tabsCtx))

		tabItems := []*container.TabItem{charTab, questsTab, dungeonsTab, statsTab, achievementsTab}
		tabItems = append(tabItems,
			container.NewTabItem("История", a.buildHistoryTab()),
		)
		appTabs = container.NewAppTabs(tabItems...)
	}
	appTabs.SetTabLocation(container.TabLocationTop)

	return container.NewBorder(header, nil, nil, nil, appTabs)
}

func (a *App) buildHeader() *fyne.Container {
	t := components.T()
	bg := canvas.NewRectangle(t.BG)
	bg.SetMinSize(fyne.NewSize(0, 52))
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.Border

	// Logo with letter spacing from theme
	logoText := "S" + t.HeaderLetterGap + "O" + t.HeaderLetterGap + "L" + t.HeaderLetterGap + "O" +
		"   " +
		"L" + t.HeaderLetterGap + "E" + t.HeaderLetterGap + "V" + t.HeaderLetterGap + "E" + t.HeaderLetterGap + "L" + t.HeaderLetterGap + "I" + t.HeaderLetterGap + "N" + t.HeaderLetterGap + "G"
	title := canvas.NewText(logoText, t.Accent)
	title.TextSize = components.TextHeadingLG
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Система Пробуждения Охотника", t.TextMuted)
	subtitle.TextSize = components.TextBodySM
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
	t := components.T()
	nameText := components.MakeTitle(a.engine.Character.Name, t.Gold, 24)
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

	levelText := components.MakeTitle(fmt.Sprintf("Общий уровень: %d", level), t.Text, 16)

	var statSummary []fyne.CanvasObject
	for _, s := range stats {
		txt := components.MakeLabel(fmt.Sprintf("%s %s: %d", s.StatType.Icon(), s.StatType.DisplayName(), s.Level), t.TextSecondary)
		statSummary = append(statSummary, txt)
	}

	top := container.NewVBox(nameRow, rankText, levelText)
	statsRow := container.NewHBox(statSummary...)

	contentItems := []fyne.CanvasObject{top, widget.NewSeparator(), statsRow}

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
		t := components.T()
		a.historyPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), t.Danger))
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
	t := components.T()
	rankBadge := components.MakeRankBadge(q.Rank)
	titleText := components.MakeTitle(q.Title, t.Text, 14)

	completedStr := ""
	if q.CompletedAt != nil {
		completedStr = q.CompletedAt.Format("02.01.2006 15:04")
	}

	var typeIndicator fyne.CanvasObject
	if q.IsDaily {
		lbl := components.MakeLabel("Ежедневное", t.Blue)
		lbl.TextSize = components.TextBodySM
		typeIndicator = lbl
	} else if q.DungeonID != nil {
		lbl := components.MakeLabel("Данж", t.Purple)
		lbl.TextSize = components.TextBodySM
		typeIndicator = lbl
	} else {
		typeIndicator = layout.NewSpacer()
	}

	dateText := components.MakeLabel(completedStr, t.TextSecondary)
	expText := components.MakeLabel(
		fmt.Sprintf("+%d EXP -> %s %s | Ранг: %s", q.Exp, q.TargetStat.Icon(), q.TargetStat.DisplayName(), q.Rank),
		t.Success,
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
	battleWindow.Resize(fyne.NewSize(820, 780))
	battleWindow.CenterOnScreen()

	topRef := container.NewVBox()
	centerRef := container.NewVBox()
	bottomRef := container.NewVBox()

	var cells []*battleCell
	var primaryStatus *canvas.Text
	var secondaryStatus *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button
	var surrenderBtn *widget.Button
	var resolved bool
	var resolvedRecord *models.BattleRecord
	var resolvedErr error
	var roundLogBox *fyne.Container
	var dmgOverlay *canvas.Text

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	// --- VS panel builder ---
	buildVSPanel := func() fyne.CanvasObject {
		t := components.T()
		// Player side
		playerIcon := canvas.NewText("⚔️", t.Accent)
		playerIcon.TextSize = 40
		playerIcon.Alignment = fyne.TextAlignCenter
		playerIconBg := canvas.NewRectangle(t.BGPanel)
		playerIconBg.CornerRadius = components.RadiusLG
		playerIconBg.SetMinSize(fyne.NewSize(80, 80))
		playerIconBg.StrokeWidth = components.BorderThin
		playerIconBg.StrokeColor = t.Border
		playerPortrait := container.NewStack(playerIconBg, container.NewCenter(playerIcon))

		playerName := components.MakeTitle("Охотник", t.Text, 14)
		playerHP := makeBattleMiniHPBar(state.PlayerHP, state.PlayerMaxHP, t.Success)
		playerHPLabel := components.MakeLabel(fmt.Sprintf("HP: %d / %d", state.PlayerHP, state.PlayerMaxHP), t.Success)
		playerHPLabel.TextSize = components.TextBodySM
		playerSide := container.NewVBox(
			container.NewCenter(playerPortrait),
			container.NewCenter(playerName),
			playerHP,
			container.NewCenter(playerHPLabel),
		)

		// VS text
		vsText := canvas.NewText("VS", t.Danger)
		vsText.TextSize = 28
		vsText.TextStyle = fyne.TextStyle{Bold: true}
		vsText.Alignment = fyne.TextAlignCenter
		roundText := components.MakeLabel(fmt.Sprintf("Раунд %d", state.Round), t.Gold)
		roundText.TextSize = components.TextHeadingSM
		roundText.Alignment = fyne.TextAlignCenter
		vsSide := container.NewVBox(layout.NewSpacer(), container.NewCenter(vsText), container.NewCenter(roundText), layout.NewSpacer())

		// Enemy side
		enemyIcon := canvas.NewText("👹", t.Danger)
		enemyIcon.TextSize = 40
		enemyIcon.Alignment = fyne.TextAlignCenter
		enemyIconBg := canvas.NewRectangle(withAlpha(t.DangerDim, 40))
		enemyIconBg.CornerRadius = components.RadiusLG
		enemyIconBg.SetMinSize(fyne.NewSize(80, 80))
		enemyIconBg.StrokeWidth = components.BorderThin
		enemyIconBg.StrokeColor = withAlpha(t.Danger, 80)
		enemyPortrait := container.NewStack(enemyIconBg, container.NewCenter(enemyIcon))

		enemyName := components.MakeTitle(state.Enemy.Name, t.Text, 14)
		rankBadge := components.MakeRankBadge(state.Enemy.Rank)
		enemyHP := makeBattleMiniHPBar(state.EnemyHP, state.EnemyMaxHP, t.Danger)
		enemyHPLabel := components.MakeLabel(fmt.Sprintf("HP: %d / %d", state.EnemyHP, state.EnemyMaxHP), t.Danger)
		enemyHPLabel.TextSize = components.TextBodySM
		enemySide := container.NewVBox(
			container.NewCenter(enemyPortrait),
			container.NewHBox(layout.NewSpacer(), enemyName, rankBadge, layout.NewSpacer()),
			enemyHP,
			container.NewCenter(enemyHPLabel),
		)

		// Put VS panel together
		bg := canvas.NewRectangle(t.BGPanel)
		bg.CornerRadius = components.RadiusLG
		bg.StrokeWidth = components.BorderThin
		bg.StrokeColor = t.Border

		vsGrid := container.NewGridWithColumns(3, playerSide, vsSide, enemySide)
		return container.NewStack(bg, container.NewPadded(vsGrid))
	}

	// --- Round log builder ---
	buildRoundLog := func() fyne.CanvasObject {
		t := components.T()
		roundLogBox = container.NewVBox()
		if len(state.RoundLog) > 0 {
			for _, line := range state.RoundLog {
				var clr color.Color = t.TextSecondary
				if strings.Contains(line, "Крит") {
					clr = t.Gold
				} else if strings.Contains(line, "Враг атакует") {
					clr = t.Danger
				}
				l := components.MakeLabel(line, clr)
				l.TextSize = components.TextBodySM
				roundLogBox.Add(l)
			}
		}
		return roundLogBox
	}

	// --- Damage overlay animation ---
	showDamageOverlay := func(parent *fyne.Container, dmg int, isCrit bool) {
		t := components.T()
		txt := fmt.Sprintf("-%d", dmg)
		var clr color.Color = t.Success
		if isCrit {
			txt = fmt.Sprintf("⚡ КРИТ! -%d", dmg)
			clr = t.Gold
		}
		dmgOverlay = canvas.NewText(txt, clr)
		dmgOverlay.TextSize = 20
		dmgOverlay.TextStyle = fyne.TextStyle{Bold: true}
		dmgOverlay.Alignment = fyne.TextAlignCenter
		parent.Add(container.NewCenter(dmgOverlay))
		parent.Refresh()
		go func() {
			time.Sleep(1200 * time.Millisecond)
			runOnMain(func() {
				parent.Remove(container.NewCenter(dmgOverlay))
				parent.Refresh()
			})
		}()
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		topRef.Objects = nil
		centerRef.Objects = nil
		bottomRef.Objects = nil

		// VS panel always on top
		topRef.Add(buildVSPanel())

		if state.BattleOver {
			t := components.T()
			var resultText string
			var resultEmoji string
			var resultColor color.Color
			if state.Result == models.BattleWin {
				resultText = "ПОБЕДА!"
				resultEmoji = "🏆"
				resultColor = t.Gold
			} else {
				resultText = "ПОРАЖЕНИЕ"
				resultEmoji = "💀"
				resultColor = t.Danger
			}

			if !resolved {
				resolvedRecord, resolvedErr = a.engine.FinishBattle(state)
				resolved = true
			}

			// Big result overlay
			bigEmoji := canvas.NewText(resultEmoji, resultColor)
			bigEmoji.TextSize = 56
			bigEmoji.Alignment = fyne.TextAlignCenter
			bigTitle := components.MakeTitle(resultText, resultColor, 28)
			bigTitle.Alignment = fyne.TextAlignCenter
			subtitle := components.MakeLabel("Бои не дают EXP. Это испытание силы.", t.TextSecondary)
			subtitle.Alignment = fyne.TextAlignCenter

			resultContent := container.NewVBox(
				container.NewCenter(bigEmoji),
				container.NewCenter(bigTitle),
				container.NewCenter(subtitle),
			)

			// Stats
			statsItems := []fyne.CanvasObject{}
			if resolvedErr != nil {
				statsItems = append(statsItems, components.MakeLabel("Ошибка: "+resolvedErr.Error(), t.Danger))
			} else if state.Result != models.BattleWin {
				statsItems = append(statsItems, components.MakeLabel("Поражение не наказывается.", t.TextSecondary))
			}
			if resolvedRecord != nil {
				statsItems = append(statsItems,
					components.MakeLabel(fmt.Sprintf("Точность: %.1f%%  |  Криты: %d  |  Раундов: %d", resolvedRecord.Accuracy, state.TotalCrits, state.Round), t.TextSecondary),
					components.MakeLabel(fmt.Sprintf("Урон нанесён: %d  |  Урон получен: %d", state.DamageDealt, state.DamageTaken), t.TextSecondary),
				)
			}

			statsBox := container.NewVBox(statsItems...)
			resultCard := components.MakeCard(container.NewVBox(resultContent, widget.NewSeparator(), statsBox))
			centerRef.Add(container.NewCenter(resultCard))

			nextLabel := "Закрыть"
			nextIcon := theme.CancelIcon()
			if state.Result == models.BattleWin {
				nextLabel = "Дальше"
				nextIcon = theme.NavigateNextIcon()
			}
			closeBtn := widget.NewButtonWithIcon(nextLabel, nextIcon, func() {
				battleWindow.Close()
				a.refreshCharacterPanel()
				a.refreshStatsPanel()
			})
			closeBtn.Importance = widget.HighImportance
			bottomRef.Add(container.NewHBox(layout.NewSpacer(), closeBtn, layout.NewSpacer()))

			topRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()
			return
		}

		// --- Active round ---
		choices := make([]int, 0, state.CellsToShow)
		selected := make(map[int]struct{}, state.CellsToShow)
		shown := make(map[int]struct{}, len(state.ShownCells))
		for _, idx := range state.ShownCells {
			shown[idx] = struct{}{}
		}
		acceptingInput := false
		roundSubmitted := false

		// Show round damage info if not first round
		if state.Round > 1 && state.LastRoundDamage > 0 {
			showDamageOverlay(topRef, state.LastRoundDamage, state.LastRoundCrit)
		}

		primaryStatus = components.MakeLabel("Запомни подсвеченные клетки", components.T().Gold)
		primaryStatus.TextSize = components.TextBodyLG
		primaryStatus.TextStyle = fyne.TextStyle{Bold: true}
		secondaryStatus = components.MakeLabel(
			fmt.Sprintf("Сетка %dx%d • Выбрано 0/%d", state.GridSize, state.GridSize, state.CellsToShow),
			components.T().TextSecondary,
		)
		secondaryStatus.TextSize = components.TextNumberSM

		cellCount := state.GridSize * state.GridSize
		cells = make([]*battleCell, cellCount)
		var gridCells []fyne.CanvasObject
		cellSize := cellSizeForGrid(battleWindow, state.GridSize)
		for i := 0; i < cellCount; i++ {
			cell := newBattleCell(cellSize)
			cell.Disable()
			cell.SetState(battleCellStateIdle)
			cells[i] = cell
			gridCells = append(gridCells, cell)
		}

		gridContainer := container.NewGridWithColumns(state.GridSize, gridCells...)
		fieldCard := components.MakeCard(container.NewPadded(gridContainer))

		logWidget := buildRoundLog()
		gridCol := container.NewVBox(
			container.NewCenter(primaryStatus),
			container.NewCenter(fieldCard),
			container.NewCenter(secondaryStatus),
			logWidget,
		)
		centerContent := container.NewBorder(nil, nil, nil, nil, gridCol)
		centerRef.Add(centerContent)

		updateSelectionStatus := func() {
			secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Выбрано %d/%d", state.GridSize, state.GridSize, len(choices), state.CellsToShow)
			secondaryStatus.Refresh()
		}

		refreshAllCells := func() {
			for _, cell := range cells {
				cell.Refresh()
			}
		}

		showRoundResult := func() {
			for i := 0; i < cellCount; i++ {
				_, wasSelected := selected[i]
				_, wasShown := shown[i]
				switch {
				case wasSelected && wasShown:
					cells[i].state = battleCellStateResultCorrect
				case wasSelected && !wasShown:
					cells[i].state = battleCellStateResultWrong
				case wasShown:
					cells[i].state = battleCellStateShown
				default:
					cells[i].state = battleCellStateIdle
				}
				cells[i].disabled = true
			}
			refreshAllCells()
		}

		submitRound := func() {
			if !acceptingInput || roundSubmitted {
				return
			}
			acceptingInput = false
			roundSubmitted = true
			confirmBtn.Disable()
			resetBtn.Disable()

			showRoundResult()
			accuracy := memory.ComputeAccuracy(state.ShownCells, choices)
			primaryStatus.Text = fmt.Sprintf("Точность %.0f%% • расчёт урона...", accuracy*100)
			primaryStatus.Color = components.T().Gold
			primaryStatus.Refresh()

			roundChoices := append([]int(nil), choices...)
			go func() {
				time.Sleep(450 * time.Millisecond)
				err := a.engine.ProcessRound(state, roundChoices)
				runOnMain(func() {
					if err != nil {
						dialog.ShowError(err, battleWindow)
						return
					}
					rebuildScreen()
				})
			}()
		}

		confirmBtn = widget.NewButtonWithIcon("Готово", theme.ConfirmIcon(), submitRound)
		confirmBtn.Importance = widget.HighImportance
		confirmBtn.Disable()

		resetBtn = widget.NewButtonWithIcon("Сбросить выбор", theme.ViewRefreshIcon(), func() {
			if !acceptingInput || roundSubmitted {
				return
			}
			choices = choices[:0]
			selected = make(map[int]struct{}, state.CellsToShow)
			for _, c := range cells {
				c.state = battleCellStateIdle
			}
			refreshAllCells()
			updateSelectionStatus()
			confirmBtn.Disable()
		})
		resetBtn.Importance = widget.MediumImportance
		resetBtn.Disable()

		surrenderBtn = widget.NewButtonWithIcon("Сдаться", theme.CancelIcon(), func() {
			state.BattleOver = true
			state.Result = models.BattleLose
			state.PlayerHP = 0
			rebuildScreen()
		})
		surrenderBtn.Importance = widget.DangerImportance

		bottomRef.Add(container.NewHBox(confirmBtn, resetBtn, layout.NewSpacer(), surrenderBtn))

		topRef.Refresh()
		centerRef.Refresh()
		bottomRef.Refresh()

		showTimeMs, _ := a.engine.GetShowTimeMs(state.ShowTimeMs)
		if showTimeMs <= 0 {
			showTimeMs = 1000
		}
		go func() {
			runOnMain(func() {
				for _, idx := range state.ShownCells {
					if idx >= 0 && idx < cellCount {
						cells[idx].state = battleCellStateShown
					}
				}
				refreshAllCells()
			})

			time.Sleep(time.Duration(showTimeMs) * time.Millisecond)

			runOnMain(func() {
				primaryStatus.Text = fmt.Sprintf("Выбери клетки: %d", state.CellsToShow)
				primaryStatus.Color = components.T().Text
				primaryStatus.Refresh()

				for i := 0; i < cellCount; i++ {
					idx := i
					cells[idx].state = battleCellStateIdle
					cells[idx].disabled = false
					cells[idx].SetOnTapped(func() {
						if !acceptingInput || roundSubmitted {
							return
						}
						if _, exists := selected[idx]; exists {
							return
						}

						selected[idx] = struct{}{}
						choices = append(choices, idx)
						cells[idx].SetState(battleCellStateSelected)
						updateSelectionStatus()

						if len(choices) >= state.CellsToShow {
							submitRound()
							return
						}
						confirmBtn.Enable()
					})
				}
				refreshAllCells()

				acceptingInput = true
				resetBtn.Enable()
				updateSelectionStatus()
			})
		}()
	}

	root := container.NewBorder(topRef, bottomRef, nil, nil, container.NewVScroll(centerRef))
	battleWindow.SetContent(container.NewPadded(root))
	battleWindow.Show()
	rebuildScreen()
}

func (a *App) showBossScreen() {
	state := a.currentBoss

	battleWindow := a.app.NewWindow(fmt.Sprintf("Босс: %s", state.Enemy.Name))
	battleWindow.Resize(fyne.NewSize(820, 780))
	battleWindow.CenterOnScreen()

	topRef := container.NewVBox()
	centerRef := container.NewVBox()
	bottomRef := container.NewVBox()

	var cells []*battleCell
	var primaryStatus *canvas.Text
	var secondaryStatus *canvas.Text
	var confirmBtn *widget.Button
	var resetBtn *widget.Button
	var surrenderBtn *widget.Button
	var resolved bool
	var resolvedRecord *models.BattleRecord
	var resolvedErr error
	var roundLogBox *fyne.Container

	runOnMain := func(fn func()) {
		if d, ok := a.app.Driver().(interface{ RunOnMain(func()) }); ok {
			d.RunOnMain(fn)
			return
		}
		fn()
	}

	// --- VS panel builder ---
	buildBossVSPanel := func() fyne.CanvasObject {
		t := components.T()
		// Player side
		playerIcon := canvas.NewText("⚔️", t.Accent)
		playerIcon.TextSize = 40
		playerIcon.Alignment = fyne.TextAlignCenter
		playerIconBg := canvas.NewRectangle(t.BGPanel)
		playerIconBg.CornerRadius = components.RadiusLG
		playerIconBg.SetMinSize(fyne.NewSize(80, 80))
		playerIconBg.StrokeWidth = components.BorderThin
		playerIconBg.StrokeColor = t.Border
		playerPortrait := container.NewStack(playerIconBg, container.NewCenter(playerIcon))

		playerName := components.MakeTitle("Охотник", t.Text, 14)
		playerHP := makeBattleMiniHPBar(state.PlayerHP, state.PlayerMaxHP, t.Success)
		playerHPLabel := components.MakeLabel(fmt.Sprintf("HP: %d / %d", state.PlayerHP, state.PlayerMaxHP), t.Success)
		playerHPLabel.TextSize = components.TextBodySM
		playerSide := container.NewVBox(
			container.NewCenter(playerPortrait),
			container.NewCenter(playerName),
			playerHP,
			container.NewCenter(playerHPLabel),
		)

		// VS text
		vsText := canvas.NewText("VS", t.Danger)
		vsText.TextSize = 28
		vsText.TextStyle = fyne.TextStyle{Bold: true}
		vsText.Alignment = fyne.TextAlignCenter
		phaseText := components.MakeLabel(phaseDisplay(state.Phase), t.Gold)
		phaseText.TextSize = components.TextHeadingSM
		phaseText.Alignment = fyne.TextAlignCenter
		roundText := components.MakeLabel(fmt.Sprintf("Раунд %d", state.Round), t.TextSecondary)
		roundText.TextSize = components.TextBodySM
		roundText.Alignment = fyne.TextAlignCenter
		vsSide := container.NewVBox(layout.NewSpacer(), container.NewCenter(vsText), container.NewCenter(phaseText), container.NewCenter(roundText), layout.NewSpacer())

		// Enemy (boss) side
		bossIcon := canvas.NewText("👑", t.Gold)
		bossIcon.TextSize = 40
		bossIcon.Alignment = fyne.TextAlignCenter
		bossIconBg := canvas.NewRectangle(withAlpha(t.DangerDim, 40))
		bossIconBg.CornerRadius = components.RadiusLG
		bossIconBg.SetMinSize(fyne.NewSize(80, 80))
		bossIconBg.StrokeWidth = components.BorderThick
		bossIconBg.StrokeColor = withAlpha(t.Danger, 200)
		bossPortrait := container.NewStack(bossIconBg, container.NewCenter(bossIcon))

		bossName := components.MakeTitle(state.Enemy.Name, t.Danger, 14)
		bossName.TextStyle = fyne.TextStyle{Bold: true}
		rankBadge := components.MakeRankBadge(state.Enemy.Rank)
		bossLabel := components.MakeLabel("BOSS", t.Danger)
		bossLabel.TextSize = 10
		bossLabel.TextStyle = fyne.TextStyle{Bold: true}

		enemyHP := makeBattleMiniHPBar(state.EnemyHP, state.EnemyMaxHP, t.Danger)
		enemyHPLabel := components.MakeLabel(fmt.Sprintf("HP: %d / %d", state.EnemyHP, state.EnemyMaxHP), t.Danger)
		enemyHPLabel.TextSize = components.TextBodySM

		enemySide := container.NewVBox(
			container.NewCenter(bossPortrait),
			container.NewHBox(layout.NewSpacer(), bossName, rankBadge, bossLabel, layout.NewSpacer()),
			enemyHP,
			container.NewCenter(enemyHPLabel),
		)

		bg := canvas.NewRectangle(t.BGPanel)
		bg.CornerRadius = components.RadiusLG
		bg.StrokeWidth = components.BorderMedium
		bg.StrokeColor = withAlpha(t.Danger, 180)

		vsGrid := container.NewGridWithColumns(3, playerSide, vsSide, enemySide)
		return container.NewStack(bg, container.NewPadded(vsGrid))
	}

	// --- Round log builder ---
	buildBossRoundLog := func() fyne.CanvasObject {
		t := components.T()
		roundLogBox = container.NewVBox()
		if len(state.RoundLog) > 0 {
			for _, line := range state.RoundLog {
				var clr color.Color = t.TextSecondary
				if strings.Contains(line, "Крит") {
					clr = t.Gold
				} else if strings.Contains(line, "Враг атакует") {
					clr = t.Danger
				}
				l := components.MakeLabel(line, clr)
				l.TextSize = components.TextBodySM
				roundLogBox.Add(l)
			}
		}
		return roundLogBox
	}

	var rebuildScreen func()
	rebuildScreen = func() {
		topRef.Objects = nil
		centerRef.Objects = nil
		bottomRef.Objects = nil

		// VS panel always on top
		topRef.Add(buildBossVSPanel())

		if state.Phase == boss.PhaseWin || state.Phase == boss.PhaseLose {
			t := components.T()
			var resultText string
			var resultEmoji string
			var resultColor color.Color
			if state.Phase == boss.PhaseWin {
				resultText = "ПОБЕДА НАД БОССОМ!"
				resultEmoji = "🏆"
				resultColor = t.Gold
			} else {
				resultText = "ПОРАЖЕНИЕ"
				resultEmoji = "💀"
				resultColor = t.Danger
			}

			if !resolved {
				if state.Phase == boss.PhaseWin {
					resolvedRecord, resolvedErr = a.engine.FinishBoss(state)
				} else {
					resolvedRecord, resolvedErr = a.engine.FailBoss(state)
				}
				resolved = true
			}

			bigEmoji := canvas.NewText(resultEmoji, resultColor)
			bigEmoji.TextSize = 56
			bigEmoji.Alignment = fyne.TextAlignCenter
			bigTitle := components.MakeTitle(resultText, resultColor, 28)
			bigTitle.Alignment = fyne.TextAlignCenter
			subtitle := components.MakeLabel("Бои не дают EXP. Это испытание силы.", t.TextSecondary)
			subtitle.Alignment = fyne.TextAlignCenter

			resultContent := container.NewVBox(
				container.NewCenter(bigEmoji),
				container.NewCenter(bigTitle),
				container.NewCenter(subtitle),
			)

			statsItems := []fyne.CanvasObject{}
			if resolvedErr != nil {
				statsItems = append(statsItems, components.MakeLabel("Ошибка: "+resolvedErr.Error(), t.Danger))
			} else if state.Phase != boss.PhaseWin {
				statsItems = append(statsItems, components.MakeLabel("Поражение не наказывается.", t.TextSecondary))
			}
			if resolvedRecord != nil {
				statsItems = append(statsItems,
					components.MakeLabel(fmt.Sprintf("Точность: %.1f%%  |  Криты: %d  |  Раундов: %d", resolvedRecord.Accuracy, state.TotalCrits, state.Round), t.TextSecondary),
					components.MakeLabel(fmt.Sprintf("Урон нанесён: %d  |  Урон получен: %d", state.DamageDealt, state.DamageTaken), t.TextSecondary),
				)
			}

			statsBox := container.NewVBox(statsItems...)
			resultCard := components.MakeCard(container.NewVBox(resultContent, widget.NewSeparator(), statsBox))
			centerRef.Add(container.NewCenter(resultCard))

			nextLabel := "Закрыть"
			nextIcon := theme.CancelIcon()
			if state.Phase == boss.PhaseWin {
				nextLabel = "Дальше"
				nextIcon = theme.NavigateNextIcon()
			}
			closeBtn := widget.NewButtonWithIcon(nextLabel, nextIcon, func() {
				battleWindow.Close()
				a.refreshCharacterPanel()
				a.refreshStatsPanel()
			})
			closeBtn.Importance = widget.HighImportance
			bottomRef.Add(container.NewHBox(layout.NewSpacer(), closeBtn, layout.NewSpacer()))

			topRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()
			return
		}

		if state.Phase == boss.PhaseMemory {
			choices := make([]int, 0, state.Memory.CellsToShow)
			selected := make(map[int]struct{}, state.Memory.CellsToShow)
			shown := make(map[int]struct{}, len(state.Memory.ShownCells))
			for _, idx := range state.Memory.ShownCells {
				shown[idx] = struct{}{}
			}
			acceptingInput := false
			roundSubmitted := false

			primaryStatus = components.MakeLabel("Запомни подсвеченные клетки", components.T().Gold)
			primaryStatus.TextSize = components.TextBodyLG
			primaryStatus.TextStyle = fyne.TextStyle{Bold: true}
			secondaryStatus = components.MakeLabel(
				fmt.Sprintf("Сетка %dx%d • Выбрано 0/%d", state.Memory.GridSize, state.Memory.GridSize, state.Memory.CellsToShow),
				components.T().TextSecondary,
			)
			secondaryStatus.TextSize = components.TextNumberSM

			cellCount := state.Memory.GridSize * state.Memory.GridSize
			cells = make([]*battleCell, cellCount)
			var gridCells []fyne.CanvasObject
			cellSize := cellSizeForGrid(battleWindow, state.Memory.GridSize)
			for i := 0; i < cellCount; i++ {
				cell := newBattleCell(cellSize)
				cell.Disable()
				cell.SetState(battleCellStateIdle)
				cells[i] = cell
				gridCells = append(gridCells, cell)
			}

			gridContainer := container.NewGridWithColumns(state.Memory.GridSize, gridCells...)
			fieldCard := components.MakeCard(container.NewPadded(gridContainer))
			logWidget := buildBossRoundLog()

			gridCol := container.NewVBox(
				container.NewCenter(primaryStatus),
				container.NewCenter(fieldCard),
				container.NewCenter(secondaryStatus),
				logWidget,
			)
			centerRef.Add(gridCol)

			updateSelectionStatus := func() {
				secondaryStatus.Text = fmt.Sprintf("Сетка %dx%d • Выбрано %d/%d", state.Memory.GridSize, state.Memory.GridSize, len(choices), state.Memory.CellsToShow)
				secondaryStatus.Refresh()
			}

			refreshAllCells := func() {
				for _, cell := range cells {
					cell.Refresh()
				}
			}

			showRoundResult := func() {
				for i := 0; i < cellCount; i++ {
					_, wasSelected := selected[i]
					_, wasShown := shown[i]
					switch {
					case wasSelected && wasShown:
						cells[i].state = battleCellStateResultCorrect
					case wasSelected && !wasShown:
						cells[i].state = battleCellStateResultWrong
					case wasShown:
						cells[i].state = battleCellStateShown
					default:
						cells[i].state = battleCellStateIdle
					}
					cells[i].disabled = true
				}
				refreshAllCells()
			}

			submitRound := func() {
				if !acceptingInput || roundSubmitted {
					return
				}
				acceptingInput = false
				roundSubmitted = true
				confirmBtn.Disable()
				resetBtn.Disable()
				showRoundResult()

				accuracy := memory.ComputeAccuracy(state.Memory.ShownCells, choices)
				primaryStatus.Text = fmt.Sprintf("Точность %.0f%% • расчёт урона...", accuracy*100)
				primaryStatus.Color = components.T().Gold
				primaryStatus.Refresh()

				roundChoices := append([]int(nil), choices...)
				go func() {
					time.Sleep(450 * time.Millisecond)
					err := a.engine.ProcessBossMemory(state, roundChoices)
					runOnMain(func() {
						if err != nil {
							dialog.ShowError(err, battleWindow)
							return
						}
						rebuildScreen()
					})
				}()
			}

			confirmBtn = widget.NewButtonWithIcon("Готово", theme.ConfirmIcon(), submitRound)
			confirmBtn.Importance = widget.HighImportance
			confirmBtn.Disable()

			resetBtn = widget.NewButtonWithIcon("Сбросить выбор", theme.ViewRefreshIcon(), func() {
				if !acceptingInput || roundSubmitted {
					return
				}
				choices = choices[:0]
				selected = make(map[int]struct{}, state.Memory.CellsToShow)
				for _, c := range cells {
					c.state = battleCellStateIdle
				}
				refreshAllCells()
				updateSelectionStatus()
				confirmBtn.Disable()
			})
			resetBtn.Importance = widget.MediumImportance
			resetBtn.Disable()

			surrenderBtn = widget.NewButtonWithIcon("Сдаться", theme.CancelIcon(), func() {
				state.Phase = boss.PhaseLose
				state.PlayerHP = 0
				rebuildScreen()
			})
			surrenderBtn.Importance = widget.DangerImportance

			bottomRef.Add(container.NewHBox(confirmBtn, resetBtn, layout.NewSpacer(), surrenderBtn))

			topRef.Refresh()
			centerRef.Refresh()
			bottomRef.Refresh()

			showTimeMs, _ := a.engine.GetShowTimeMs(state.Memory.ShowTimeMs)
			if showTimeMs <= 0 {
				showTimeMs = 1000
			}
			go func() {
				runOnMain(func() {
					for _, idx := range state.Memory.ShownCells {
						if idx >= 0 && idx < cellCount {
							cells[idx].state = battleCellStateShown
						}
					}
					refreshAllCells()
				})

				time.Sleep(time.Duration(showTimeMs) * time.Millisecond)

				runOnMain(func() {
					primaryStatus.Text = fmt.Sprintf("Выбери клетки: %d", state.Memory.CellsToShow)
					primaryStatus.Color = components.T().Text
					primaryStatus.Refresh()

					for i := 0; i < cellCount; i++ {
						idx := i
						cells[idx].state = battleCellStateIdle
						cells[idx].disabled = false
						cells[idx].SetOnTapped(func() {
							if !acceptingInput || roundSubmitted {
								return
							}
							if _, exists := selected[idx]; exists {
								return
							}

							selected[idx] = struct{}{}
							choices = append(choices, idx)
							cells[idx].SetState(battleCellStateSelected)
							updateSelectionStatus()

							if len(choices) >= state.Memory.CellsToShow {
								submitRound()
								return
							}
							confirmBtn.Enable()
						})
					}
					refreshAllCells()

					acceptingInput = true
					resetBtn.Enable()
					updateSelectionStatus()
				})
			}()
			return
		}
	}

	root := container.NewBorder(topRef, bottomRef, nil, nil, container.NewVScroll(centerRef))
	battleWindow.SetContent(container.NewPadded(root))
	battleWindow.Show()
	rebuildScreen()
}

func phaseDisplay(p boss.Phase) string {
	switch p {
	case boss.PhaseMemory:
		return "Visual Memory"
	case boss.PhaseWin:
		return "Победа"
	case boss.PhaseLose:
		return "Поражение"
	default:
		return string(p)
	}
}

func (a *App) buildBattleHistoryCard(b models.BattleRecord) *fyne.Container {
	t := components.T()
	var resultText string
	var resultColor color.Color
	if b.Result == models.BattleWin {
		resultText = "Победа"
		resultColor = t.Success
	} else {
		resultText = "Поражение"
		resultColor = t.Danger
	}

	nameText := components.MakeTitle(b.EnemyName, t.Text, 14)
	result := components.MakeLabel(resultText, resultColor)
	result.TextStyle = fyne.TextStyle{Bold: true}
	dateText := components.MakeLabel(b.FoughtAt.Format("02.01.2006 15:04"), t.TextSecondary)

	statsText := components.MakeLabel(
		fmt.Sprintf("Урон: %d | Точность: %.0f%% | Криты: %d", b.DamageDealt, b.Accuracy, b.CriticalHits),
		t.TextSecondary,
	)

	topRow := container.NewHBox(nameText, result, layout.NewSpacer(), dateText)
	content := container.NewVBox(topRow, statsText)
	return components.MakeCard(content)
}

func cellSizeForGrid(win fyne.Window, grid int) float32 {
	if grid <= 0 {
		return 44
	}

	// Fallback size while canvas is not ready.
	fallback := float32(54)
	switch grid {
	case 3:
		fallback = 96
	case 4:
		fallback = 78
	case 5:
		fallback = 64
	case 6:
		fallback = 54
	case 7:
		fallback = 46
	case 8:
		fallback = 40
	default:
		if grid > 8 {
			fallback = 36
		} else {
			fallback = 60
		}
	}

	if win == nil {
		return fallback
	}

	canvasSize := win.Canvas().Size()
	if canvasSize.Width <= 0 || canvasSize.Height <= 0 {
		return fallback
	}

	// Keep the board responsive to the window size.
	availableWidth := canvasSize.Width * 0.62
	availableHeight := canvasSize.Height * 0.52
	sideByWidth := availableWidth / float32(grid)
	sideByHeight := availableHeight / float32(grid)

	side := sideByWidth
	if sideByHeight < side {
		side = sideByHeight
	}

	if side < 28 {
		side = 28
	}
	if side > 100 {
		side = 100
	}
	return side
}

func buildBattleHPRow(name string, current, max int, fillColor color.Color) fyne.CanvasObject {
	t := components.T()
	label := components.MakeLabel(fmt.Sprintf("%s %d/%d", name, current, max), t.Text)
	label.TextSize = components.TextHeadingSM
	bar := makeBattleMiniHPBar(current, max, fillColor)
	return container.NewVBox(label, bar)
}

func buildBattleAttemptsBox(attempts int) fyne.CanvasObject {
	t := components.T()
	label := components.MakeLabel(fmt.Sprintf("Попытки %d/%d", attempts, models.MaxAttempts), t.Text)
	label.TextSize = components.TextHeadingSM
	bar := makeBattleMiniHPBar(attempts, models.MaxAttempts, t.Accent)
	return container.NewVBox(label, bar)
}

func makeBattleMiniHPBar(current, max int, fillColor color.Color) *fyne.Container {
	t := components.T()
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

	bg := canvas.NewRectangle(t.BGPanel)
	bg.CornerRadius = components.RadiusMD
	bg.SetMinSize(fyne.NewSize(180, 12))
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.BorderActive

	fill := canvas.NewRectangle(fillColor)
	fill.CornerRadius = components.RadiusMD

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
	battleCellStateShown
	battleCellStateSelected
	battleCellStateResultCorrect
	battleCellStateResultWrong
)

type battleCell struct {
	widget.BaseWidget

	minSize  fyne.Size
	state    battleCellState
	disabled bool
	onTapped func()
}

func newBattleCell(side float32) *battleCell {
	if side < 28 {
		side = 28
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
	t := components.T()

	type palette struct {
		fill   color.NRGBA
		border color.NRGBA
		glow   color.NRGBA
	}

	idle := palette{
		fill:   withAlpha(t.BGCard, 220),
		border: withAlpha(t.BorderActive, 220),
		glow:   color.NRGBA{A: 0},
	}
	shown := palette{
		fill:   withAlpha(t.Accent, 60),
		border: t.Accent,
		glow:   t.AccentGlow,
	}
	selected := palette{
		fill:   withAlpha(t.Accent, 40),
		border: withAlpha(t.Accent, 140),
		glow:   color.NRGBA{A: 0},
	}
	resultCorrect := palette{
		fill:   withAlpha(t.Success, 80),
		border: t.Success,
		glow:   withAlpha(t.Success, 160),
	}
	resultWrong := palette{
		fill:   withAlpha(t.Danger, 80),
		border: t.Danger,
		glow:   withAlpha(t.Danger, 160),
	}

	current := idle
	switch state {
	case battleCellStateShown:
		current = shown
	case battleCellStateSelected:
		current = selected
	case battleCellStateResultCorrect:
		current = resultCorrect
	case battleCellStateResultWrong:
		current = resultWrong
	}

	if disabled {
		current.fill.A = dimAlpha(current.fill.A, 45)
		current.border.A = dimAlpha(current.border.A, 65)
		current.glow.A = dimAlpha(current.glow.A, 85)
	}

	return current.fill, current.border, current.glow
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: a}
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
