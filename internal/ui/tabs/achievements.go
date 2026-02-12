package tabs

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

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

	// Build grid of cards, 3 per row
	var cards []fyne.CanvasObject
	for _, a := range list {
		cards = append(cards, buildAchievementTile(a))
	}

	grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 170)), cards...)
	ctx.AchievementsPanel.Add(grid)
	ctx.AchievementsPanel.Refresh()
}

// achievementIcon maps achievement key → emoji icon.
func achievementIcon(key string) string {
	switch key {
	case "first_task":
		return "⚔️"
	case "first_battle":
		return "🗡️"
	case "streak_7":
		return "🔥"
	case "first_dungeon":
		return "🏰"
	default:
		return "🏆"
	}
}

func buildAchievementTile(a models.Achievement) fyne.CanvasObject {
	const cardW float32 = 180
	const cardH float32 = 170

	if a.IsUnlocked {
		return buildUnlockedTile(a, cardW, cardH)
	}
	return buildLockedTile(a, cardW, cardH)
}

func buildUnlockedTile(a models.Achievement, w, h float32) fyne.CanvasObject {
	// Background with accent border
	bg := canvas.NewRectangle(components.ColorBGCard)
	bg.CornerRadius = 10
	bg.SetMinSize(fyne.NewSize(w, h))

	border := canvas.NewRectangle(color.NRGBA{R: 80, G: 60, B: 180, A: 255})
	border.CornerRadius = 10
	border.SetMinSize(fyne.NewSize(w, h))
	border.StrokeWidth = 2
	border.StrokeColor = color.NRGBA{R: 100, G: 80, B: 220, A: 255}

	// Icon — large centered emoji
	icon := canvas.NewText(achievementIcon(a.Key), components.ColorAccentBright)
	icon.TextSize = 44
	icon.Alignment = fyne.TextAlignCenter
	iconRow := container.NewHBox(layout.NewSpacer(), icon, layout.NewSpacer())

	// Title
	title := canvas.NewText(a.Title, components.ColorText)
	title.TextSize = 13
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	titleRow := container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer())

	// Description
	desc := canvas.NewText(a.Description, components.ColorTextDim)
	desc.TextSize = 10
	desc.Alignment = fyne.TextAlignCenter
	descRow := container.NewHBox(layout.NewSpacer(), desc, layout.NewSpacer())

	// Date
	var dateRow fyne.CanvasObject
	if a.ObtainedAt != nil {
		dateText := canvas.NewText(
			a.ObtainedAt.Local().Format("02.01.2006"),
			components.ColorGreen,
		)
		dateText.TextSize = 10
		dateText.Alignment = fyne.TextAlignCenter
		dateRow = container.NewHBox(layout.NewSpacer(), dateText, layout.NewSpacer())
	} else {
		dateText := canvas.NewText("Получено", components.ColorGreen)
		dateText.TextSize = 10
		dateText.Alignment = fyne.TextAlignCenter
		dateRow = container.NewHBox(layout.NewSpacer(), dateText, layout.NewSpacer())
	}

	content := container.NewVBox(
		layout.NewSpacer(),
		iconRow,
		titleRow,
		descRow,
		dateRow,
		layout.NewSpacer(),
	)

	inset := container.New(layout.NewCustomPaddedLayout(8, 8, 8, 8), content)
	return container.NewStack(border, bg, inset)
}

func buildLockedTile(a models.Achievement, w, h float32) fyne.CanvasObject {
	// Dark muted background
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 30, A: 255})
	bg.CornerRadius = 10
	bg.SetMinSize(fyne.NewSize(w, h))

	// Icon — dimmed
	icon := canvas.NewText(achievementIcon(a.Key), color.NRGBA{R: 80, G: 80, B: 100, A: 255})
	icon.TextSize = 44
	icon.Alignment = fyne.TextAlignCenter
	iconRow := container.NewHBox(layout.NewSpacer(), icon, layout.NewSpacer())

	// Title — grey
	title := canvas.NewText(a.Title, color.NRGBA{R: 90, G: 90, B: 110, A: 255})
	title.TextSize = 13
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	titleRow := container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer())

	// Description — very dim
	desc := canvas.NewText(a.Description, color.NRGBA{R: 70, G: 70, B: 85, A: 255})
	desc.TextSize = 10
	desc.Alignment = fyne.TextAlignCenter
	descRow := container.NewHBox(layout.NewSpacer(), desc, layout.NewSpacer())

	// Lock icon in corner
	lock := canvas.NewText("🔒", color.NRGBA{R: 70, G: 70, B: 85, A: 255})
	lock.TextSize = 14
	lockCorner := container.NewHBox(layout.NewSpacer(), lock)

	// Status
	status := canvas.NewText("Не получено", color.NRGBA{R: 70, G: 70, B: 85, A: 255})
	status.TextSize = 10
	status.Alignment = fyne.TextAlignCenter
	statusRow := container.NewHBox(layout.NewSpacer(), status, layout.NewSpacer())

	content := container.NewVBox(
		lockCorner,
		layout.NewSpacer(),
		iconRow,
		titleRow,
		descRow,
		statusRow,
		layout.NewSpacer(),
	)

	inset := container.New(layout.NewCustomPaddedLayout(4, 8, 8, 8), content)
	return container.NewStack(bg, inset)
}

// categoryLabel returns a display name for achievement category.
func categoryLabel(category string) string {
	switch category {
	case "discipline":
		return "Дисциплина"
	case "combat":
		return "Бой"
	case "streak":
		return "Серия"
	case "dungeon":
		return "Данжи"
	default:
		return fmt.Sprintf("%s", category)
	}
}
