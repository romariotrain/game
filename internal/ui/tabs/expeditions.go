package tabs

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildExpeditions(ctx *Context) fyne.CanvasObject {
	ctx.ExpeditionsPanel = container.NewVBox()
	RefreshExpeditions(ctx)
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(components.MakeSectionHeader("Экспедиции"), ctx.ExpeditionsPanel),
	))
}

func RefreshExpeditions(ctx *Context) {
	if ctx.ExpeditionsPanel == nil {
		return
	}
	ctx.ExpeditionsPanel.Objects = nil

	t := components.T()

	if err := ctx.Engine.RefreshExpeditionStatuses(ctx.Features.FailExpiredExpeditions); err != nil {
		ctx.ExpeditionsPanel.Add(components.MakeLabel("Ошибка обновления: "+err.Error(), t.Danger))
	}

	expeditions, err := ctx.Engine.DB.GetAllExpeditions()
	if err != nil {
		ctx.ExpeditionsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), t.Danger))
		ctx.ExpeditionsPanel.Refresh()
		return
	}

	completed, _ := ctx.Engine.DB.GetCompletedExpeditions(ctx.Engine.Character.ID)
	if len(completed) > 0 {
		ctx.ExpeditionsPanel.Add(components.MakeTitle(fmt.Sprintf("Завершено экспедиций: %d", len(completed)), t.Gold, components.TextHeadingMD))
		ctx.ExpeditionsPanel.Add(widget.NewSeparator())
	}

	for _, ex := range expeditions {
		ctx.ExpeditionsPanel.Add(buildExpeditionCard(ctx, ex))
	}

	ctx.ExpeditionsPanel.Refresh()
}

func buildExpeditionCard(ctx *Context, ex models.Expedition) *fyne.Container {
	t := components.T()
	statusIcon := ""
	statusText := ""
	statusColor := t.TextSecondary

	switch ex.Status {
	case models.ExpeditionActive:
		statusIcon = "🧭"
		statusText = "Активна"
		statusColor = t.Success
	case models.ExpeditionCompleted:
		statusIcon = "✅"
		statusText = "Завершена"
		statusColor = t.Gold
	case models.ExpeditionFailed:
		statusIcon = "⛔"
		statusText = "Провалена"
		statusColor = t.Danger
	default:
		statusIcon = "•"
		statusText = string(ex.Status)
	}

	nameText := components.MakeTitle(ex.Name, t.Text, components.TextHeadingMD)
	statusBadge := components.MakeLabel(statusIcon+" "+statusText, statusColor)
	statusBadge.TextStyle = fyne.TextStyle{Bold: true}
	descText := components.MakeLabel(ex.Description, t.TextSecondary)

	deadlineText := "Дедлайн: без ограничения"
	if ex.Deadline != nil {
		deadlineText = "Дедлайн: " + ex.Deadline.Local().Format("02.01.2006")
	}
	deadlineLabel := components.MakeLabel(deadlineText, t.TextSecondary)

	rewardText := components.MakeLabel(
		fmt.Sprintf("Награда: +%d EXP всем статам | Бонус: %s", ex.RewardEXP, formatRewardStats(ex.RewardStats)),
		t.Gold,
	)

	completedTasks, totalTasks, percent, err := ctx.Engine.GetExpeditionProgress(ex.ID)
	progressText := components.MakeLabel("Прогресс: 0 / 0 задач (0%)", t.Accent)
	if err == nil {
		progressText.Text = fmt.Sprintf("Прогресс: %d / %d задач (%.0f%%)", completedTasks, totalTasks, percent)
		progressText.Refresh()
	}
	progressBar := components.MakeEXPBar(completedTasks, max(1, totalTasks), t.Accent)

	contentItems := []fyne.CanvasObject{nameText, statusBadge, descText, deadlineLabel, rewardText, progressText, progressBar}

	if len(ex.Tasks) > 0 {
		contentItems = append(contentItems, widget.NewSeparator())
		for _, task := range ex.Tasks {
			icon := "[ ]"
			color := t.Text
			if task.IsCompleted {
				icon = "[✓]"
				color = t.Success
			}
			line := components.MakeLabel(
				fmt.Sprintf("  %s %s (%d/%d)", icon, task.Title, task.ProgressCurrent, max(1, task.ProgressTarget)),
				color,
			)
			contentItems = append(contentItems, line)
		}
	}

	if ex.Status == models.ExpeditionActive {
		startBtn := widget.NewButtonWithIcon("Начать / Продолжить", theme.MediaPlayIcon(), func() {
			spawned, err := ctx.Engine.StartExpedition(ex.ID)
			if err != nil {
				dialog.ShowError(err, ctx.Window)
				return
			}
			if spawned == 0 {
				dialog.ShowInformation("Экспедиция", "Новых задач не создано: либо всё уже в работе, либо экспедиция завершена.", ctx.Window)
			}
			RefreshExpeditions(ctx)
			RefreshQuests(ctx)
		})
		startBtn.Importance = widget.HighImportance
		contentItems = append(contentItems, startBtn)
	}

	if ex.Status == models.ExpeditionCompleted {
		contentItems = append(contentItems, components.MakeLabel("Экспедиция завершена. Награды выданы.", t.Gold))
	}
	if ex.Status == models.ExpeditionFailed {
		contentItems = append(contentItems, components.MakeLabel("Экспедиция провалена по дедлайну. Награды не выдаются.", t.Danger))
	}

	return components.MakeCard(container.NewVBox(contentItems...))
}

func formatRewardStats(stats map[models.StatType]int) string {
	if len(stats) == 0 {
		return "нет"
	}
	parts := make([]string, 0, len(stats))
	keys := make([]string, 0, len(stats))
	for stat := range stats {
		keys = append(keys, string(stat))
	}
	sort.Strings(keys)
	for _, key := range keys {
		stat := models.StatType(key)
		parts = append(parts, fmt.Sprintf("%s %+d", stat.DisplayName(), stats[stat]))
	}
	return strings.Join(parts, ", ")
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
