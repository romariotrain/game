package components

import (
	"fmt"
	"image/color"
	"strings"

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
	t := T()
	sep := canvas.NewRectangle(t.Border)
	sep.SetMinSize(fyne.NewSize(0, 1))
	return sep
}

// MakeCard — legacy card (no border). For new code prefer MakeHUDPanel.
func MakeCard(content fyne.CanvasObject) *fyne.Container {
	t := T()
	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = RadiusMD
	return container.NewStack(bg, container.NewPadded(content))
}

// EXP Progress Bar - custom styled
func MakeEXPBar(current, max int, barColor color.Color) *fyne.Container {
	t := T()
	ratio := float64(current) / float64(max)
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}

	bgBar := canvas.NewRectangle(t.BGPanel)
	bgBar.SetMinSize(fyne.NewSize(200, 16))
	bgBar.CornerRadius = RadiusSM
	bgBar.StrokeWidth = BorderThin
	bgBar.StrokeColor = t.Border

	fillBar := canvas.NewRectangle(barColor)
	fillBar.CornerRadius = RadiusSM

	expText := canvas.NewText(fmt.Sprintf("%d / %d", current, max), t.Text)
	expText.TextSize = TextBodySM
	expText.Alignment = fyne.TextAlignCenter

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
	t := T()
	icon := MakeLabel(stat.StatType.Icon(), t.Text)
	name := MakeTitle(stat.StatType.DisplayName(), t.Text, TextHeadingMD)
	levelText := MakeTitle(fmt.Sprintf("Ур. %d", stat.Level), t.Accent, TextHeadingMD)

	required := models.ExpForLevel(stat.Level)
	expBar := MakeEXPBar(stat.CurrentEXP, required, t.AccentDim)

	left := container.NewHBox(icon, name, layout.NewSpacer(), levelText)
	return container.NewVBox(left, expBar)
}

// Rank badge
func MakeRankBadge(rank models.QuestRank) *fyne.Container {
	clr := ParseHexColor(rank.Color())
	bg := canvas.NewRectangle(clr)
	bg.CornerRadius = RadiusSM
	bg.SetMinSize(fyne.NewSize(32, 24))

	text := canvas.NewText(string(rank), color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	text.TextSize = TextHeadingSM
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.Alignment = fyne.TextAlignCenter

	return container.NewStack(bg, container.NewCenter(text))
}

func ParseHexColor(hex string) color.NRGBA {
	var r, g, b uint8
	if len(hex) == 7 {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

// Section header — uses system style when CornerBrackets enabled.
func MakeSectionHeader(title string) *fyne.Container {
	t := T()
	if t.CornerBrackets || t.HeaderUppercase {
		return MakeSystemHeader(title)
	}
	// Classic style
	titleText := MakeTitle(title, t.Accent, TextHeadingLG)
	sep := MakeSeparator()
	return container.NewVBox(titleText, sep)
}

// MakeSystemLabel creates styled label matching system theme.
// Uppercase if theme demands it.
func MakeSystemLabel(text string, clr color.Color, size float32) *canvas.Text {
	t := T()
	display := text
	if t.HeaderUppercase {
		display = strings.ToUpper(text)
	}
	label := canvas.NewText(display, clr)
	label.TextSize = size
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

// Empty state placeholder
func MakeEmptyState(text string) *fyne.Container {
	t := T()
	label := MakeLabel(text, t.TextSecondary)
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(label)
}

// Styled button helper
func MakeStyledButton(label string, icon fyne.Resource, fn func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, fn)
	btn.Importance = widget.HighImportance
	return btn
}
