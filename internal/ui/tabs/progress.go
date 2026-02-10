package tabs

import (
	"fmt"
	"math"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildProgress(ctx *Context) fyne.CanvasObject {
	ctx.StatsPanel = container.NewVBox()
	RefreshProgress(ctx)
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(components.MakeSectionHeader("Статистика Охотника"), ctx.StatsPanel),
	))
}

func RefreshProgress(ctx *Context) {
	if ctx.StatsPanel == nil {
		return
	}
	ctx.StatsPanel.Objects = nil

	stats, err := ctx.Engine.GetStatistics()
	if err != nil {
		ctx.StatsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		ctx.StatsPanel.Refresh()
		return
	}

	overallCard := buildOverallStatsCard(stats)
	ctx.StatsPanel.Add(overallCard)

	rankCard := buildRankStatsCard(stats)
	ctx.StatsPanel.Add(rankCard)

	if ctx.Features.Combat && !ctx.Features.MinimalMode {
		battleStats, err := ctx.Engine.GetBattleStats()
		if err == nil && battleStats.TotalBattles > 0 {
			bCard := buildBattleStatsCard(battleStats)
			ctx.StatsPanel.Add(bCard)
		}
	}

	chartCard := buildActivityChart(ctx)
	ctx.StatsPanel.Add(chartCard)

	ctx.StatsPanel.Refresh()
}

func buildOverallStatsCard(stats *models.Statistics) *fyne.Container {
	header := components.MakeTitle("Общая статистика", components.ColorAccentBright, 16)

	totalCompleted := components.MakeLabel(fmt.Sprintf("Выполнено заданий: %d", stats.TotalQuestsCompleted), components.ColorText)
	totalFailed := components.MakeLabel(fmt.Sprintf("Провалено заданий: %d", stats.TotalQuestsFailed), components.ColorText)
	successRate := components.MakeLabel(fmt.Sprintf("Процент успеха: %.1f%%", stats.SuccessRate), components.ColorGreen)
	totalEXP := components.MakeLabel(fmt.Sprintf("Всего EXP получено: %d", stats.TotalEXPEarned), components.ColorGold)
	streak := components.MakeLabel(fmt.Sprintf("Текущий Streak: %d дней подряд", stats.CurrentStreak), components.ColorAccentBright)
	bestStat := components.MakeLabel(
		fmt.Sprintf("Лучший стат: %s %s (Ур. %d)", stats.BestStat.Icon(), stats.BestStat.DisplayName(), stats.BestStatLevel),
		components.ColorText,
	)

	content := container.NewVBox(header, widget.NewSeparator(), totalCompleted, totalFailed, successRate, totalEXP, streak, bestStat)
	return components.MakeCard(content)
}

func buildRankStatsCard(stats *models.Statistics) *fyne.Container {
	header := components.MakeTitle("Задания по рангам", components.ColorAccentBright, 16)

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())

	for _, rank := range models.AllRanks {
		count := stats.QuestsByRank[rank]
		clr := components.ParseHexColor(rank.Color())
		label := components.MakeLabel(fmt.Sprintf("Ранг %s: %d заданий", string(rank), count), clr)
		rows = append(rows, label)
	}

	content := container.NewVBox(rows...)
	return components.MakeCard(content)
}

func buildBattleStatsCard(stats *models.BattleStatistics) *fyne.Container {
	header := components.MakeTitle("Боевая статистика", components.ColorAccentBright, 16)

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())
	rows = append(rows, components.MakeLabel(fmt.Sprintf("Всего боёв: %d", stats.TotalBattles), components.ColorText))
	rows = append(rows, components.MakeLabel(fmt.Sprintf("Побед: %d / Поражений: %d", stats.Wins, stats.Losses), components.ColorText))
	rows = append(rows, components.MakeLabel(fmt.Sprintf("Winrate: %.1f%%", stats.WinRate), components.ColorGreen))
	rows = append(rows, components.MakeLabel(fmt.Sprintf("Общий урон: %d", stats.TotalDamage), components.ColorRed))
	rows = append(rows, components.MakeLabel(fmt.Sprintf("Критов: %d / Уклонений: %d", stats.TotalCrits, stats.TotalDodges), components.ColorAccentBright))

	if len(stats.EnemiesDefeated) > 0 {
		rows = append(rows, widget.NewSeparator())
		rows = append(rows, components.MakeTitle("Побеждённые враги:", components.ColorText, 14))
		var names []string
		for name := range stats.EnemiesDefeated {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			count := stats.EnemiesDefeated[name]
			rows = append(rows, components.MakeLabel(fmt.Sprintf("  %s: %d раз", name, count), components.ColorTextDim))
		}
	}

	content := container.NewVBox(rows...)
	return components.MakeCard(content)
}

func buildActivityChart(ctx *Context) *fyne.Container {
	header := components.MakeTitle("Активность 30 дней", components.ColorAccentBright, 16)

	activities, err := ctx.Engine.DB.GetDailyActivityLast30(ctx.Engine.Character.ID)
	if err != nil {
		return components.MakeCard(components.MakeLabel("Ошибка загрузки активности", components.ColorRed))
	}

	activityMap := make(map[string]models.DailyActivity)
	for _, a := range activities {
		activityMap[a.Date] = a
	}

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())

	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		displayDate := date.Format("02.01")

		act, ok := activityMap[dateStr]
		if !ok {
			rows = append(rows, components.MakeLabel(fmt.Sprintf("  %s: нет данных", displayDate), components.ColorTextDim))
			continue
		}

		var barColor = components.ColorTextDim
		if act.QuestsComplete > 0 {
			intensity := act.QuestsComplete
			if intensity >= 5 {
				barColor = components.ColorGreen
			} else if intensity >= 3 {
				barColor = components.ColorAccentBright
			} else {
				barColor = components.ColorText
			}
		}

		bar := canvas.NewRectangle(barColor)
		barHeight := float32(6)
		barWidth := float32(math.Min(12+float64(act.QuestsComplete*12), 200))
		bar.SetMinSize(fyne.NewSize(barWidth, barHeight))

		row := container.NewHBox(
			components.MakeLabel(displayDate, components.ColorTextDim),
			bar,
			components.MakeLabel(
				fmt.Sprintf("  %d заданий, +%d EXP", act.QuestsComplete, act.EXPEarned),
				components.ColorTextDim,
			),
		)
		rows = append(rows, row)
	}

	content := container.NewVBox(rows...)
	return components.MakeCard(content)
}
