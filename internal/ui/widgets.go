package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
)

func MakeTitle(text string, clr color.Color, size float32) *canvas.Text {
	t := canvas.NewText(text, clr)
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

func MakeLabel(text string, clr color.Color) *canvas.Text {
	t := canvas.NewText(text, clr)
	t.TextSize = 14
	return t
}

func MakeSeparator() *canvas.Rectangle {
	sep := canvas.NewRectangle(color.NRGBA{R: 60, G: 50, B: 120, A: 100})
	sep.SetMinSize(fyne.NewSize(0, 2))
	return sep
}

func MakeCard(content fyne.CanvasObject) *fyne.Container {
	bg := canvas.NewRectangle(ColorBGCard)
	bg.CornerRadius = 8
	return container.NewStack(bg, container.NewPadded(content))
}

// EXP Progress Bar - custom styled
func MakeEXPBar(current, max int, barColor color.Color) *fyne.Container {
	ratio := float64(current) / float64(max)
	if ratio > 1 {
		ratio = 1
	}

	bgBar := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 50, A: 255})
	bgBar.SetMinSize(fyne.NewSize(200, 16))
	bgBar.CornerRadius = 4

	fillBar := canvas.NewRectangle(barColor)
	fillBar.CornerRadius = 4

	expText := canvas.NewText(fmt.Sprintf("%d / %d", current, max), ColorText)
	expText.TextSize = 11
	expText.Alignment = fyne.TextAlignCenter

	// Use a stack with the background and a container for the fill
	barContainer := container.NewStack(
		bgBar,
		container.New(&progressLayout{ratio: ratio}, fillBar),
		container.NewCenter(expText),
	)

	return barContainer
}

type progressLayout struct {
	ratio float64
}

func (p *progressLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(200, 16)
}

func (p *progressLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, obj := range objects {
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(fyne.NewSize(containerSize.Width*float32(p.ratio), containerSize.Height))
	}
}

// Stat row widget
func MakeStatRow(stat models.StatLevel) *fyne.Container {
	icon := MakeLabel(stat.StatType.Icon(), ColorText)
	name := MakeTitle(stat.StatType.DisplayName(), ColorText, 15)
	levelText := MakeTitle(fmt.Sprintf("Ур. %d", stat.Level), ColorAccentBright, 15)

	required := models.ExpForLevel(stat.Level)
	expBar := MakeEXPBar(stat.CurrentEXP, required, ColorAccent)

	left := container.NewHBox(icon, name, layout.NewSpacer(), levelText)
	return container.NewVBox(left, expBar)
}

// Rank badge
func MakeRankBadge(rank models.QuestRank) *fyne.Container {
	clr := parseHexColor(rank.Color())
	bg := canvas.NewRectangle(clr)
	bg.CornerRadius = 4
	bg.SetMinSize(fyne.NewSize(32, 24))

	text := canvas.NewText(string(rank), color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	text.TextSize = 13
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.Alignment = fyne.TextAlignCenter

	return container.NewStack(bg, container.NewCenter(text))
}

func parseHexColor(hex string) color.NRGBA {
	var r, g, b uint8
	if len(hex) == 7 {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

// Section header
func MakeSectionHeader(title string) *fyne.Container {
	t := MakeTitle(title, ColorAccentBright, 18)
	sep := MakeSeparator()
	return container.NewVBox(t, sep)
}

// Empty state placeholder
func MakeEmptyState(text string) *fyne.Container {
	label := MakeLabel(text, ColorTextDim)
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(label)
}

// Styled button helper
func MakeStyledButton(label string, icon fyne.Resource, fn func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, fn)
	btn.Importance = widget.HighImportance
	return btn
}
