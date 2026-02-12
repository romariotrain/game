package components

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// ----------------------------------------------------------------
// HUD Panel — main themed container with optional corner brackets
// ----------------------------------------------------------------

// MakeHUDPanel wraps content in a themed card with border and optional corner brackets.
func MakeHUDPanel(content fyne.CanvasObject) *fyne.Container {
	t := T()
	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = RadiusLG
	bg.StrokeWidth = BorderThin
	bg.StrokeColor = t.Border

	inset := container.New(layout.NewCustomPaddedLayout(SpaceLG, SpaceLG, SpaceLG, SpaceLG), content)

	if t.CornerBrackets {
		brackets := makeCornerBrackets(t.AccentDim)
		return container.NewStack(bg, brackets, inset)
	}
	return container.NewStack(bg, inset)
}

// MakeHUDPanelSized wraps content in a themed card with minSize.
func MakeHUDPanelSized(content fyne.CanvasObject, minSize fyne.Size) *fyne.Container {
	t := T()
	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = RadiusLG
	bg.StrokeWidth = BorderThin
	bg.StrokeColor = t.Border
	bg.SetMinSize(minSize)

	inset := container.New(layout.NewCustomPaddedLayout(SpaceLG, SpaceLG, SpaceLG, SpaceLG), content)

	if t.CornerBrackets {
		brackets := makeCornerBrackets(t.AccentDim)
		return container.NewStack(bg, brackets, inset)
	}
	return container.NewStack(bg, inset)
}

// MakeHUDPanelAccent wraps content with an accent-colored border (for enemy/boss).
func MakeHUDPanelAccent(content fyne.CanvasObject, borderColor color.NRGBA) *fyne.Container {
	t := T()
	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = RadiusLG
	bg.StrokeWidth = BorderThick
	bg.StrokeColor = borderColor

	inset := container.New(layout.NewCustomPaddedLayout(SpaceLG, SpaceLG, SpaceLG, SpaceLG), content)

	if t.CornerBrackets {
		brackets := makeCornerBrackets(borderColor)
		return container.NewStack(bg, brackets, inset)
	}
	return container.NewStack(bg, inset)
}

// ----------------------------------------------------------------
// Corner Brackets — decorative L-shaped corners
// ----------------------------------------------------------------

// makeCornerBrackets creates 4 L-shaped corners using canvas lines.
// Each bracket is two lines of length bracketLen from each corner.
func makeCornerBrackets(clr color.NRGBA) fyne.CanvasObject {
	return container.New(&cornerBracketLayout{
		color:      clr,
		bracketLen: 14,
		thickness:  BorderMedium,
	})
}

type cornerBracketLayout struct {
	color      color.NRGBA
	bracketLen float32
	thickness  float32
}

func (l *cornerBracketLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

func (l *cornerBracketLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	// Layout is called but lines are in Objects() — we use a different approach.
	// The actual lines are created in a wrapper that draws them.
}

// Instead of a custom layout for lines, we use a Stack with positioned rectangles.
// This is simpler and more reliable in Fyne.

// MakeCornerBracketsFor creates corner bracket decorations at given size.
func MakeCornerBracketsFor(clr color.NRGBA) fyne.CanvasObject {
	const bLen float32 = 14
	const bThick float32 = 1.5

	// We create thin rectangles for each bracket arm
	// Top-left: horizontal + vertical
	tlH := canvas.NewRectangle(clr)
	tlH.SetMinSize(fyne.NewSize(bLen, bThick))
	tlV := canvas.NewRectangle(clr)
	tlV.SetMinSize(fyne.NewSize(bThick, bLen))

	// Top-right
	trH := canvas.NewRectangle(clr)
	trH.SetMinSize(fyne.NewSize(bLen, bThick))
	trV := canvas.NewRectangle(clr)
	trV.SetMinSize(fyne.NewSize(bThick, bLen))

	// Bottom-left
	blH := canvas.NewRectangle(clr)
	blH.SetMinSize(fyne.NewSize(bLen, bThick))
	blV := canvas.NewRectangle(clr)
	blV.SetMinSize(fyne.NewSize(bThick, bLen))

	// Bottom-right
	brH := canvas.NewRectangle(clr)
	brH.SetMinSize(fyne.NewSize(bLen, bThick))
	brV := canvas.NewRectangle(clr)
	brV.SetMinSize(fyne.NewSize(bThick, bLen))

	objects := []fyne.CanvasObject{tlH, tlV, trH, trV, blH, blV, brH, brV}
	return container.New(&bracketPositioner{bracketLen: bLen, thickness: bThick}, objects...)
}

type bracketPositioner struct {
	bracketLen float32
	thickness  float32
}

func (bp *bracketPositioner) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

func (bp *bracketPositioner) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 8 {
		return
	}
	bL := bp.bracketLen
	bT := bp.thickness
	w := size.Width
	h := size.Height

	const m float32 = 4 // margin from edge

	// Top-left horizontal
	objects[0].Move(fyne.NewPos(m, m))
	objects[0].Resize(fyne.NewSize(bL, bT))
	// Top-left vertical
	objects[1].Move(fyne.NewPos(m, m))
	objects[1].Resize(fyne.NewSize(bT, bL))

	// Top-right horizontal
	objects[2].Move(fyne.NewPos(w-m-bL, m))
	objects[2].Resize(fyne.NewSize(bL, bT))
	// Top-right vertical
	objects[3].Move(fyne.NewPos(w-m-bT, m))
	objects[3].Resize(fyne.NewSize(bT, bL))

	// Bottom-left horizontal
	objects[4].Move(fyne.NewPos(m, h-m-bT))
	objects[4].Resize(fyne.NewSize(bL, bT))
	// Bottom-left vertical
	objects[5].Move(fyne.NewPos(m, h-m-bL))
	objects[5].Resize(fyne.NewSize(bT, bL))

	// Bottom-right horizontal
	objects[6].Move(fyne.NewPos(w-m-bL, h-m-bT))
	objects[6].Resize(fyne.NewSize(bL, bT))
	// Bottom-right vertical
	objects[7].Move(fyne.NewPos(w-m-bT, h-m-bL))
	objects[7].Resize(fyne.NewSize(bT, bL))
}

// ----------------------------------------------------------------
// System Section Header — [ TITLE ] with line
// ----------------------------------------------------------------

// MakeSystemHeader creates a section header: "[ TITLE ]" with accent color + separator line.
func MakeSystemHeader(title string) *fyne.Container {
	t := T()

	displayTitle := title
	if t.HeaderUppercase {
		displayTitle = strings.ToUpper(title)
	}

	bracketL := canvas.NewText("[", t.TextMuted)
	bracketL.TextSize = TextHeadingLG
	bracketL.TextStyle = fyne.TextStyle{Bold: true}

	titleText := canvas.NewText(" "+displayTitle+" ", t.Accent)
	titleText.TextSize = TextHeadingLG
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	bracketR := canvas.NewText("]", t.TextMuted)
	bracketR.TextSize = TextHeadingLG
	bracketR.TextStyle = fyne.TextStyle{Bold: true}

	line := canvas.NewRectangle(t.Border)
	line.SetMinSize(fyne.NewSize(0, 1))

	headerRow := container.NewHBox(bracketL, titleText, bracketR)
	return container.NewVBox(headerRow, line)
}

// MakeSystemHeaderCompact — smaller version for inside panels.
func MakeSystemHeaderCompact(title string) *fyne.Container {
	t := T()

	displayTitle := title
	if t.HeaderUppercase {
		displayTitle = strings.ToUpper(title)
	}

	titleText := canvas.NewText(displayTitle, t.Accent)
	titleText.TextSize = TextHeadingSM
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	line := canvas.NewRectangle(t.Border)
	line.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(titleText, line)
}

// ----------------------------------------------------------------
// Stat Chip — compact stat display with left accent bar
// ----------------------------------------------------------------

// MakeStatChip creates a compact stat chip with colored left bar.
func MakeStatChip(icon string, code string, level int, currentEXP, maxEXP int, statColor color.NRGBA) *fyne.Container {
	t := T()

	// Left color bar
	leftBar := canvas.NewRectangle(statColor)
	leftBar.SetMinSize(fyne.NewSize(3, 0))

	// Icon
	iconText := canvas.NewText(icon, statColor)
	iconText.TextSize = 14

	// Code + Level
	codeText := canvas.NewText(code, t.Text)
	codeText.TextSize = TextHeadingSM
	codeText.TextStyle = fyne.TextStyle{Bold: true}

	lvlText := canvas.NewText(fmt.Sprintf("Lv.%d", level), statColor)
	lvlText.TextSize = TextNumberMD
	lvlText.TextStyle = fyne.TextStyle{Bold: true}

	// EXP text
	expText := canvas.NewText(fmt.Sprintf("%d/%d", currentEXP, maxEXP), t.TextSecondary)
	expText.TextSize = TextBodySM

	headerRow := container.NewHBox(iconText, codeText, layout.NewSpacer(), lvlText)

	// Mini progress bar
	bar := MakeProgressBarThin(currentEXP, maxEXP, statColor)

	// EXP label below
	content := container.NewVBox(headerRow, bar, expText)

	return container.NewBorder(nil, nil, leftBar, nil,
		container.New(layout.NewCustomPaddedLayout(SpaceSM, SpaceSM, SpaceSM, SpaceSM), content),
	)
}

// MakeProgressBarThin creates a thin 8px progress bar.
func MakeProgressBarThin(current, max int, fillColor color.NRGBA) fyne.CanvasObject {
	t := T()
	if max <= 0 {
		max = 1
	}
	ratio := float64(current) / float64(max)
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}

	bg := canvas.NewRectangle(t.BGPanel)
	bg.CornerRadius = RadiusSM
	bg.SetMinSize(fyne.NewSize(0, 8))
	bg.StrokeWidth = BorderThin
	bg.StrokeColor = t.Border

	fill := canvas.NewRectangle(fillColor)
	fill.CornerRadius = RadiusSM

	return container.NewStack(
		bg,
		container.New(&progressLayout{ratio: ratio}, fill),
	)
}
