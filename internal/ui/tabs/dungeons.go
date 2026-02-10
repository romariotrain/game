package tabs

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildDungeons(ctx *Context) fyne.CanvasObject {
	ctx.DungeonsPanel = container.NewVBox()
	RefreshDungeons(ctx)
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(components.MakeSectionHeader("Данжи"), ctx.DungeonsPanel),
	))
}

func RefreshDungeons(ctx *Context) {
	if ctx.DungeonsPanel == nil {
		return
	}
	ctx.DungeonsPanel.Objects = nil

	ctx.Engine.RefreshDungeonStatuses()

	dungeons, err := ctx.Engine.DB.GetAllDungeons()
	if err != nil {
		ctx.DungeonsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		ctx.DungeonsPanel.Refresh()
		return
	}

	completedDungeons, _ := ctx.Engine.DB.GetCompletedDungeons(ctx.Engine.Character.ID)
	if len(completedDungeons) > 0 {
		ctx.DungeonsPanel.Add(components.MakeTitle("Пройденные данжи", components.ColorGold, 16))
		for _, cd := range completedDungeons {
			ctx.DungeonsPanel.Add(components.MakeLabel("  "+cd.EarnedTitle, components.ColorPurple))
		}
		ctx.DungeonsPanel.Add(widget.NewSeparator())
	}

	for _, dungeon := range dungeons {
		card := buildDungeonCard(ctx, dungeon)
		ctx.DungeonsPanel.Add(card)
	}

	ctx.DungeonsPanel.Refresh()
}

func buildDungeonCard(ctx *Context, d models.Dungeon) *fyne.Container {
	statusIcon := ""
	statusColor := components.ColorTextDim
	statusText := ""

	switch d.Status {
	case models.DungeonLocked:
		statusIcon = "🔒"
		statusColor = components.ColorTextDim
		statusText = "Закрыт"
	case models.DungeonAvailable:
		statusIcon = "⚔️"
		statusColor = components.ColorGreen
		statusText = "Доступен"
	case models.DungeonInProgress:
		statusIcon = "⏳"
		statusColor = components.ColorBlue
		statusText = "В процессе"
	case models.DungeonCompleted:
		statusIcon = "✅"
		statusColor = components.ColorGold
		statusText = "Пройден"
	}

	nameText := components.MakeTitle(d.Name, components.ColorText, 16)
	statusBadge := components.MakeLabel(statusIcon+" "+statusText, statusColor)
	statusBadge.TextStyle = fyne.TextStyle{Bold: true}

	descText := components.MakeLabel(d.Description, components.ColorTextDim)

	var reqParts []string
	for _, req := range d.Requirements {
		reqParts = append(reqParts, fmt.Sprintf("%s %d", req.StatType.DisplayName(), req.MinLevel))
	}
	reqText := components.MakeLabel("Требования: "+strings.Join(reqParts, ", "), components.ColorTextDim)

	rewardText := components.MakeLabel(
		fmt.Sprintf("Награда: Титул '%s' + %d EXP", d.RewardTitle, d.RewardEXP),
		components.ColorGold,
	)

	contentItems := []fyne.CanvasObject{nameText, statusBadge, descText, reqText, rewardText}

	if d.Status == models.DungeonInProgress {
		completed, total, err := ctx.Engine.GetDungeonProgress(d.ID)
		if err == nil {
			progressText := components.MakeLabel(fmt.Sprintf("Прогресс: %d / %d заданий", completed, total), components.ColorAccentBright)
			progressBar := components.MakeEXPBar(completed, total, components.ColorAccentBright)
			contentItems = append(contentItems, progressText, progressBar)
		}

		allQuests, err := ctx.Engine.DB.GetDungeonAllQuests(ctx.Engine.Character.ID, d.ID)
		if err == nil && len(allQuests) > 0 {
			contentItems = append(contentItems, widget.NewSeparator())
			for _, q := range allQuests {
				qStatus := ""
				qColor := components.ColorText
				switch q.Status {
				case models.QuestCompleted:
					qStatus = "[✓]"
					qColor = components.ColorGreen
				case models.QuestActive:
					qStatus = "[ ]"
					qColor = components.ColorText
				default:
					qStatus = "[X]"
					qColor = components.ColorRed
				}
				ql := components.MakeLabel(fmt.Sprintf("  %s %s (%s)", qStatus, q.Title, string(q.Rank)), qColor)
				contentItems = append(contentItems, ql)
			}
		}
	}

	if d.Status == models.DungeonLocked || d.Status == models.DungeonAvailable {
		contentItems = append(contentItems, components.MakeLabel(fmt.Sprintf("Заданий в данже: %d", len(d.QuestDefinitions)), components.ColorTextDim))
		for _, qd := range d.QuestDefinitions {
			ql := components.MakeLabel(
				fmt.Sprintf("  - %s (Ранг %s, %s)", qd.Title, string(qd.Rank), qd.TargetStat.DisplayName()),
				components.ColorTextDim,
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
						if err := ctx.Engine.EnterDungeon(d.ID); err != nil {
							dialog.ShowError(err, ctx.Window)
							return
						}
						RefreshDungeons(ctx)
						RefreshQuests(ctx)
					}
				}, ctx.Window)
		})
		enterBtn.Importance = widget.HighImportance
		contentItems = append(contentItems, enterBtn)
	}

	if d.Status == models.DungeonCompleted {
		contentItems = append(contentItems, components.MakeLabel("Данж пройден!", components.ColorGold))
	}

	content := container.NewVBox(contentItems...)
	return components.MakeCard(content)
}
