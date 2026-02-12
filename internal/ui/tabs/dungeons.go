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

	t := components.T()

	ctx.Engine.RefreshDungeonStatuses()

	dungeons, err := ctx.Engine.DB.GetAllDungeons()
	if err != nil {
		ctx.DungeonsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), t.Danger))
		ctx.DungeonsPanel.Refresh()
		return
	}

	completedDungeons, _ := ctx.Engine.DB.GetCompletedDungeons(ctx.Engine.Character.ID)
	if len(completedDungeons) > 0 {
		ctx.DungeonsPanel.Add(components.MakeTitle("Пройденные данжи", t.Gold, components.TextHeadingMD))
		for _, cd := range completedDungeons {
			ctx.DungeonsPanel.Add(components.MakeLabel("  "+cd.EarnedTitle, t.Purple))
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
	t := components.T()
	statusIcon := ""
	var statusColor = t.TextSecondary
	statusText := ""

	switch d.Status {
	case models.DungeonLocked:
		statusIcon = "🔒"
		statusColor = t.TextSecondary
		statusText = "Закрыт"
	case models.DungeonAvailable:
		statusIcon = "⚔️"
		statusColor = t.Success
		statusText = "Доступен"
	case models.DungeonInProgress:
		statusIcon = "⏳"
		statusColor = t.Blue
		statusText = "В процессе"
	case models.DungeonCompleted:
		statusIcon = "✅"
		statusColor = t.Gold
		statusText = "Пройден"
	}

	nameText := components.MakeTitle(d.Name, t.Text, components.TextHeadingMD)
	statusBadge := components.MakeLabel(statusIcon+" "+statusText, statusColor)
	statusBadge.TextStyle = fyne.TextStyle{Bold: true}

	descText := components.MakeLabel(d.Description, t.TextSecondary)

	var reqParts []string
	for _, req := range d.Requirements {
		reqParts = append(reqParts, fmt.Sprintf("%s %d", req.StatType.DisplayName(), req.MinLevel))
	}
	reqText := components.MakeLabel("Требования: "+strings.Join(reqParts, ", "), t.TextSecondary)

	rewardText := components.MakeLabel(
		fmt.Sprintf("Награда: Титул '%s' + %d EXP", d.RewardTitle, d.RewardEXP),
		t.Gold,
	)

	contentItems := []fyne.CanvasObject{nameText, statusBadge, descText, reqText, rewardText}

	if d.Status == models.DungeonInProgress {
		completed, total, err := ctx.Engine.GetDungeonProgress(d.ID)
		if err == nil {
			progressText := components.MakeLabel(fmt.Sprintf("Прогресс: %d / %d заданий", completed, total), t.Accent)
			progressBar := components.MakeEXPBar(completed, total, t.Accent)
			contentItems = append(contentItems, progressText, progressBar)
		}

		allQuests, err := ctx.Engine.DB.GetDungeonAllQuests(ctx.Engine.Character.ID, d.ID)
		if err == nil && len(allQuests) > 0 {
			contentItems = append(contentItems, widget.NewSeparator())
			for _, q := range allQuests {
				qStatus := ""
				var qColor = t.Text
				switch q.Status {
				case models.QuestCompleted:
					qStatus = "[✓]"
					qColor = t.Success
				case models.QuestActive:
					qStatus = "[ ]"
					qColor = t.Text
				default:
					qStatus = "[X]"
					qColor = t.Danger
				}
				ql := components.MakeLabel(fmt.Sprintf("  %s %s (%s)", qStatus, q.Title, string(q.Rank)), qColor)
				contentItems = append(contentItems, ql)
			}
		}
	}

	if d.Status == models.DungeonLocked || d.Status == models.DungeonAvailable {
		contentItems = append(contentItems, components.MakeLabel(fmt.Sprintf("Заданий в данже: %d", len(d.QuestDefinitions)), t.TextSecondary))
		for _, qd := range d.QuestDefinitions {
			ql := components.MakeLabel(
				fmt.Sprintf("  - %s (Ранг %s, %s)", qd.Title, string(qd.Rank), qd.TargetStat.DisplayName()),
				t.TextSecondary,
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
		contentItems = append(contentItems, components.MakeLabel("Данж пройден!", t.Gold))
	}

	content := container.NewVBox(contentItems...)
	return components.MakeCard(content)
}
