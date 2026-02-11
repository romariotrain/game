package tabs

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildAchievements(ctx *Context) fyne.CanvasObject {
	ctx.AchievementsPanel = container.NewVBox()
	RefreshAchievements(ctx)
	return container.NewVScroll(container.NewPadded(
		container.NewVBox(
			components.MakeSectionHeader("Путь охотника"),
			ctx.AchievementsPanel,
		),
	))
}

func RefreshAchievements(ctx *Context) {
	if ctx.AchievementsPanel == nil {
		return
	}
	ctx.AchievementsPanel.Objects = nil

	list, err := ctx.Engine.GetAchievements()
	if err != nil {
		ctx.AchievementsPanel.Add(components.MakeLabel("Ошибка: "+err.Error(), components.ColorRed))
		ctx.AchievementsPanel.Refresh()
		return
	}

	if len(list) == 0 {
		ctx.AchievementsPanel.Add(components.MakeEmptyState("Достижения пока не инициализированы."))
		ctx.AchievementsPanel.Refresh()
		return
	}

	for _, a := range list {
		ctx.AchievementsPanel.Add(buildAchievementCard(a))
	}
	ctx.AchievementsPanel.Refresh()
}

func buildAchievementCard(a models.Achievement) fyne.CanvasObject {
	titleColor := components.ColorTextDim
	descColor := components.ColorTextDim
	statusColor := components.ColorTextDim
	statusText := "Не получено"
	if a.IsUnlocked {
		titleColor = components.ColorGold
		descColor = components.ColorText
		statusColor = components.ColorGreen
		if a.ObtainedAt != nil {
			statusText = "Получено: " + a.ObtainedAt.Local().Format("02.01.2006 15:04")
		} else {
			statusText = "Получено"
		}
	}

	title := components.MakeTitle(a.Title, titleColor, 15)
	desc := components.MakeLabel(a.Description, descColor)
	status := components.MakeLabel(statusText, statusColor)
	category := components.MakeLabel(fmt.Sprintf("Категория: %s", a.Category), components.ColorTextDim)

	content := container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		category,
		status,
	)

	if a.IsUnlocked {
		return components.MakeCard(content)
	}

	bg := canvas.NewRectangle(color.NRGBA{R: 24, G: 24, B: 32, A: 255})
	bg.CornerRadius = 8
	return container.NewStack(bg, container.NewPadded(content))
}
