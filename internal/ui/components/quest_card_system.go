package components

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
)

type QuestCardSystemData struct {
	Rank        models.QuestRank
	Title       string
	MetaStat    string
	EXP         int
	Description string
	Tag         string
	Priority    bool
}

type QuestCardSystemActions struct {
	OnComplete func()
	OnFail     func()
	OnDelete   func()
}

// MakeQuestCardSystem renders a compact HUD-style quest card.
func MakeQuestCardSystem(data QuestCardSystemData, actions QuestCardSystemActions) *fyne.Container {
	t := QuestTokens()
	rankColor := QuestRankColor(data.Rank)

	borderColor := t.PanelBorderSoft
	if data.Priority {
		borderColor = colorWithAlpha(rankColor, 190)
	}

	bg := canvas.NewRectangle(t.PanelBg)
	bg.CornerRadius = t.CornerRadiusPanel
	bg.StrokeWidth = BorderThin
	bg.StrokeColor = borderColor

	shadow := canvas.NewRectangle(colorWithAlpha(t.PanelBorder, 56))
	shadow.CornerRadius = t.CornerRadiusPanel

	leftAccent := canvas.NewRectangle(rankColor)
	leftAccent.SetMinSize(fyne.NewSize(4, 0))

	rankBadge := makeQuestRankBadge(data.Rank, rankColor)
	title := canvas.NewText(strings.ToUpper(strings.TrimSpace(data.Title)), t.TextPrimary)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}

	headerLeft := container.NewHBox(rankBadge, title)
	headerRow := headerLeft
	if data.Tag != "" {
		tagLabel := canvas.NewText(strings.ToUpper(data.Tag), t.TextMuted)
		tagLabel.TextSize = TextBodySM
		headerRow = container.NewBorder(nil, nil, nil, tagLabel, headerLeft)
	}

	metaStat := canvas.NewText(strings.ToUpper(strings.TrimSpace(data.MetaStat)), t.TextMuted)
	metaStat.TextSize = TextBodySM
	metaExp := canvas.NewText(fmt.Sprintf("+%d EXP", data.EXP), t.Accent)
	metaExp.TextSize = TextBodySM
	sep1 := canvas.NewText(" • ", t.TextMuted)
	sep1.TextSize = TextBodySM
	metaRow := container.NewHBox(metaStat, sep1, metaExp)

	descText := strings.TrimSpace(data.Description)
	bodyItems := []fyne.CanvasObject{headerRow, metaRow}
	if descText != "" {
		descLabel := canvas.NewText(descText, t.TextMuted)
		descLabel.TextSize = TextBodySM
		bodyItems = append(bodyItems, descLabel)
	}

	completeBtn := widget.NewButtonWithIcon("Выполнить", theme.ConfirmIcon(), actions.OnComplete)
	completeBtn.Importance = widget.MediumImportance
	completeWrap := container.NewGridWrap(fyne.NewSize(116, 30), completeBtn)

	failBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), actions.OnFail)
	failBtn.Importance = widget.LowImportance
	failWrap := container.NewGridWrap(fyne.NewSize(30, 30), failBtn)

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), actions.OnDelete)
	deleteBtn.Importance = widget.LowImportance
	deleteWrap := container.NewGridWrap(fyne.NewSize(30, 30), deleteBtn)

	actionRow := container.NewHBox(completeWrap, failWrap, deleteWrap)
	actionsCol := container.NewVBox(layout.NewSpacer(), actionRow)

	body := container.NewVBox(bodyItems...)
	content := container.NewBorder(
		nil, nil,
		leftAccent,
		container.New(layout.NewCustomPaddedLayout(0, 0, SpaceMD, SpaceSM), actionsCol),
		container.New(layout.NewCustomPaddedLayout(SpaceMD, SpaceMD, SpaceXL, SpaceMD), body),
	)

	panel := container.NewStack(bg, content)
	return container.New(&questShadowLayout{offsetY: 2}, shadow, panel)
}

func makeQuestRankBadge(rank models.QuestRank, rankColor color.NRGBA) fyne.CanvasObject {
	bg := canvas.NewRectangle(rankColor)
	bg.CornerRadius = 8
	bg.StrokeWidth = BorderThin
	bg.StrokeColor = colorWithAlpha(rankColor, 190)
	bg.SetMinSize(fyne.NewSize(28, 28))

	letter := canvas.NewText(strings.ToUpper(string(rank)), color.NRGBA{R: 245, G: 248, B: 252, A: 255})
	letter.TextSize = TextBodySM
	letter.TextStyle = fyne.TextStyle{Bold: true}
	letter.Alignment = fyne.TextAlignCenter

	return container.NewStack(bg, container.NewCenter(letter))
}

func colorWithAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

type questShadowLayout struct {
	offsetX float32
	offsetY float32
}

func (l *questShadowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	base := objects[0].MinSize()
	for i := 1; i < len(objects); i++ {
		s := objects[i].MinSize()
		if s.Width > base.Width {
			base.Width = s.Width
		}
		if s.Height > base.Height {
			base.Height = s.Height
		}
	}

	if l.offsetX > 0 {
		base.Width += l.offsetX
	}
	if l.offsetY > 0 {
		base.Height += l.offsetY
	}
	return base
}

func (l *questShadowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}

	width := size.Width
	height := size.Height
	if l.offsetX > 0 {
		width -= l.offsetX
	}
	if l.offsetY > 0 {
		height -= l.offsetY
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	panelSize := fyne.NewSize(width, height)
	objects[0].Move(fyne.NewPos(l.offsetX, l.offsetY))
	objects[0].Resize(panelSize)
	objects[1].Move(fyne.NewPos(0, 0))
	objects[1].Resize(panelSize)
}
