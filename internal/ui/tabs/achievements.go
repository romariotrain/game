package tabs

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"solo-leveling/internal/models"
	"solo-leveling/internal/ui/components"
)

func BuildAchievements(ctx *Context) fyne.CanvasObject {
	ctx.AchievementsPanel = container.NewMax()
	RefreshAchievements(ctx)
	return ctx.AchievementsPanel
}

func RefreshAchievements(ctx *Context) {
	if ctx.AchievementsPanel == nil {
		return
	}

	list, err := ctx.Engine.GetAchievements()
	if err != nil {
		t := components.T()
		ctx.AchievementsPanel.Objects = []fyne.CanvasObject{
			container.NewPadded(
				components.MakeLabel("Ошибка: "+err.Error(), t.Danger),
			),
		}
		ctx.AchievementsPanel.Refresh()
		return
	}

	achievementsTab := container.NewTabItem(
		"Достижения",
		container.NewVScroll(buildAchievementsTabContent(list)),
	)
	enemyGalleryTab := container.NewTabItem(
		"Галерея врагов",
		container.NewVScroll(buildEnemyGalleryTabContent(ctx)),
	)

	tabs := container.NewAppTabs(achievementsTab, enemyGalleryTab)
	tabs.SetTabLocation(container.TabLocationTop)

	header := container.NewPadded(components.MakeSectionHeader("Путь охотника"))
	content := container.NewBorder(header, nil, nil, nil, container.NewPadded(tabs))

	ctx.AchievementsPanel.Objects = []fyne.CanvasObject{content}
	ctx.AchievementsPanel.Refresh()
}

func buildAchievementsTabContent(list []models.Achievement) fyne.CanvasObject {
	if len(list) == 0 {
		return components.MakeEmptyState("Достижения пока не инициализированы.")
	}

	var cards []fyne.CanvasObject
	for _, a := range list {
		cards = append(cards, buildAchievementTile(a))
	}

	grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 170)), cards...)
	return container.NewVBox(grid)
}

func buildEnemyGalleryTabContent(ctx *Context) fyne.CanvasObject {
	t := components.T()
	defeated, err := ctx.Engine.DB.GetDefeatedEnemies(ctx.Engine.Character.ID)
	if err != nil {
		return components.MakeHUDPanel(container.NewVBox(
			components.MakeSystemHeaderCompact("Галерея убитых врагов"),
			components.MakeLabel("Ошибка: "+err.Error(), t.Danger),
		))
	}

	title := components.MakeSystemHeaderCompact("Галерея убитых врагов")
	description := components.MakeLabel("Нажмите на зону, затем на изображение врага, чтобы открыть его историю.", t.TextSecondary)
	description.TextSize = components.TextBodySM

	if len(defeated) == 0 {
		empty := components.MakeLabel("Побеждённых врагов пока нет.", t.TextSecondary)
		empty.TextSize = components.TextBodySM
		content := container.NewVBox(
			makeAchievementsGap(components.SpaceLG),
			title,
			description,
			empty,
		)
		return centerWithMaxWidth(content, 960)
	}

	zones := map[int][]models.DefeatedEnemy{}
	for _, enemy := range defeated {
		zones[enemy.Zone] = append(zones[enemy.Zone], enemy)
	}

	accordion := widget.NewAccordion()
	first := true
	for zone := 1; zone <= 5; zone++ {
		enemies := zones[zone]
		detail := buildZoneEnemyList(ctx, enemies)
		if len(enemies) == 0 {
			empty := components.MakeLabel("В этой зоне пока нет побеждённых врагов.", t.TextSecondary)
			empty.TextSize = components.TextBodySM
			detail = container.NewVBox(empty)
		}
		item := widget.NewAccordionItem(
			fmt.Sprintf("Zone %d · %s (%d)", zone, zoneLabel(zone), len(enemies)),
			detail,
		)
		if first {
			item.Open = true
			first = false
		}
		accordion.Append(item)
	}

	content := container.NewVBox(
		makeAchievementsGap(components.SpaceLG),
		title,
		description,
		makeAchievementsGap(components.SpaceSM),
		accordion,
	)
	return centerWithMaxWidth(content, 1200)
}

func buildZoneEnemyList(ctx *Context, enemies []models.DefeatedEnemy) fyne.CanvasObject {
	list := container.NewVBox()
	for i, enemy := range enemies {
		if i > 0 {
			list.Add(makeAchievementsGap(components.SpaceLG))
		}
		list.Add(buildDefeatedEnemyGalleryCard(ctx, enemy))
	}
	return list
}

// achievementIcon maps achievement key -> emoji icon.
func achievementIcon(key string) string {
	switch key {
	case "first_task":
		return "⚔️"
	case "first_battle":
		return "🗡️"
	case "streak_7":
		return "🔥"
	case "first_expedition":
		return "🧭"
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
	t := components.T()

	bg := canvas.NewRectangle(t.BGCard)
	bg.CornerRadius = components.RadiusLG
	bg.SetMinSize(fyne.NewSize(w, h))
	bg.StrokeWidth = components.BorderMedium
	bg.StrokeColor = t.AccentDim

	icon := canvas.NewText(achievementIcon(a.Key), t.Accent)
	icon.TextSize = 44
	icon.Alignment = fyne.TextAlignCenter
	iconRow := container.NewHBox(layout.NewSpacer(), icon, layout.NewSpacer())

	title := canvas.NewText(a.Title, t.Text)
	title.TextSize = components.TextHeadingSM
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	titleRow := container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer())

	desc := canvas.NewText(a.Description, t.TextSecondary)
	desc.TextSize = components.TextBodySM - 1
	desc.Alignment = fyne.TextAlignCenter
	descRow := container.NewHBox(layout.NewSpacer(), desc, layout.NewSpacer())

	var dateRow fyne.CanvasObject
	if a.ObtainedAt != nil {
		dateText := canvas.NewText(a.ObtainedAt.Local().Format("02.01.2006"), t.Success)
		dateText.TextSize = components.TextBodySM - 1
		dateText.Alignment = fyne.TextAlignCenter
		dateRow = container.NewHBox(layout.NewSpacer(), dateText, layout.NewSpacer())
	} else {
		dateText := canvas.NewText("Получено", t.Success)
		dateText.TextSize = components.TextBodySM - 1
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

	inset := container.New(layout.NewCustomPaddedLayout(
		components.SpaceSM, components.SpaceSM, components.SpaceSM, components.SpaceSM,
	), content)

	if t.CornerBrackets {
		brackets := components.MakeCornerBracketsFor(t.AccentDim)
		return container.NewStack(bg, brackets, inset)
	}
	return container.NewStack(bg, inset)
}

func buildLockedTile(a models.Achievement, w, h float32) fyne.CanvasObject {
	t := components.T()

	bg := canvas.NewRectangle(t.BGPanel)
	bg.CornerRadius = components.RadiusLG
	bg.SetMinSize(fyne.NewSize(w, h))
	bg.StrokeWidth = components.BorderThin
	bg.StrokeColor = t.Border

	icon := canvas.NewText(achievementIcon(a.Key), t.TextMuted)
	icon.TextSize = 44
	icon.Alignment = fyne.TextAlignCenter
	iconRow := container.NewHBox(layout.NewSpacer(), icon, layout.NewSpacer())

	title := canvas.NewText(a.Title, t.TextMuted)
	title.TextSize = components.TextHeadingSM
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	titleRow := container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer())

	desc := canvas.NewText(a.Description, t.TextMuted)
	desc.TextSize = components.TextBodySM - 1
	desc.Alignment = fyne.TextAlignCenter
	descRow := container.NewHBox(layout.NewSpacer(), desc, layout.NewSpacer())

	lock := canvas.NewText("🔒", t.TextMuted)
	lock.TextSize = 14
	lockCorner := container.NewHBox(layout.NewSpacer(), lock)

	status := canvas.NewText("Не получено", t.TextMuted)
	status.TextSize = components.TextBodySM - 1
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

	inset := container.New(layout.NewCustomPaddedLayout(
		components.SpaceXS, components.SpaceSM, components.SpaceSM, components.SpaceSM,
	), content)
	return container.NewStack(bg, inset)
}

func buildDefeatedEnemyGalleryCard(ctx *Context, enemy models.DefeatedEnemy) fyne.CanvasObject {
	t := components.T()
	panelBg := canvas.NewRectangle(t.BGCard)
	panelBg.CornerRadius = components.RadiusXL
	panelBg.StrokeWidth = components.BorderThin
	panelBg.StrokeColor = t.Border

	name := components.MakeTitle(enemy.Name, t.Text, components.TextBodyLG)
	rankBadge := components.MakeRankBadge(enemy.Rank)
	nameRow := container.NewHBox(rankBadge, name)

	dateText := "дата неизвестна"
	if enemy.DefeatedAt != nil {
		dateText = enemy.DefeatedAt.Local().Format("02.01.2006 15:04")
	}
	metaText := fmt.Sprintf("Побеждён: %s", dateText)
	if enemy.IsBoss {
		metaText += " · BOSS"
	}
	metaLabel := components.MakeLabel(metaText, t.TextSecondary)
	metaLabel.TextSize = components.TextBodySM

	hint := components.MakeLabel("Нажмите на портрет, чтобы открыть историю", t.AccentDim)
	hint.TextSize = components.TextBodySM

	content := container.NewVBox(
		nameRow,
		metaLabel,
		hint,
		makeAchievementsGap(components.SpaceSM),
		buildEnemyPortrait(ctx, enemy),
	)
	inset := container.New(layout.NewCustomPaddedLayout(
		components.SpaceMD, components.SpaceMD, components.SpaceMD, components.SpaceMD,
	), content)
	return container.NewStack(panelBg, inset)
}

func buildEnemyPortrait(ctx *Context, enemy models.DefeatedEnemy) fyne.CanvasObject {
	t := components.T()
	const maxWidth float32 = 320
	portraitSize := fyne.NewSize(220, 160)
	frame := canvas.NewRectangle(color.Transparent)
	frame.CornerRadius = components.RadiusMD
	frame.StrokeWidth = components.BorderThin
	frame.StrokeColor = t.Border
	frame.SetMinSize(portraitSize)

	var portrait fyne.CanvasObject
	path := enemyPortraitPath(enemy)
	if path != "" {
		portraitSize = enemyPortraitSize(path, maxWidth)
		frame.SetMinSize(portraitSize)
		img := canvas.NewImageFromFile(path)
		// Keep original proportions: no forced 3:4 stretch.
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(portraitSize)
		portrait = img
	} else {
		bg := canvas.NewRectangle(t.BGPanel)
		bg.CornerRadius = components.RadiusMD
		bg.StrokeWidth = components.BorderThin
		bg.StrokeColor = t.Border
		bg.SetMinSize(portraitSize)
		img := canvas.NewImageFromResource(enemyPortraitResource(enemy))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(52, 52))
		portrait = container.NewStack(bg, container.NewCenter(img))
	}

	showStory := func() {
		if ctx == nil || ctx.Window == nil {
			return
		}
		dialog.ShowInformation(enemy.Name, enemyLore(enemy), ctx.Window)
	}

	stack := []fyne.CanvasObject{portrait, frame, newTapOverlay(showStory)}
	if enemy.IsBoss {
		bossTag := components.MakeLabel("BOSS", t.Danger)
		bossTag.TextSize = components.TextBodySM
		bossTag.TextStyle = fyne.TextStyle{Bold: true}
		tagLayer := container.NewBorder(container.NewHBox(layout.NewSpacer(), bossTag), nil, nil, nil, nil)
		stack = append(stack, tagLayer)
	}
	return container.NewStack(stack...)
}

func enemyPortraitSize(path string, maxWidth float32) fyne.Size {
	file, err := os.Open(path)
	if err != nil {
		return fyne.NewSize(maxWidth, maxWidth*0.6)
	}
	defer file.Close()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return fyne.NewSize(maxWidth, maxWidth*0.6)
	}

	width := float32(cfg.Width)
	height := float32(cfg.Height)
	if width > maxWidth {
		scale := maxWidth / width
		width = maxWidth
		height *= scale
	}
	return fyne.NewSize(width, height)
}

func enemyPortraitPath(enemy models.DefeatedEnemy) string {
	candidates := []string{}
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		candidates = append(candidates,
			fmt.Sprintf("assets/enemies/enemy_%d%s", enemy.EnemyID, ext),
			fmt.Sprintf("assets/enemies/%s%s", enemyFileSlug(enemy.Name), ext),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func enemyPortraitResource(enemy models.DefeatedEnemy) fyne.Resource {
	var icons = []fyne.Resource{
		theme.VisibilityIcon(),
		theme.WarningIcon(),
		theme.InfoIcon(),
		theme.NavigateNextIcon(),
		theme.MediaPlayIcon(),
		theme.ConfirmIcon(),
		theme.CancelIcon(),
		theme.AccountIcon(),
	}
	if len(icons) == 0 {
		return theme.AccountIcon()
	}
	idx := int(enemy.EnemyID % int64(len(icons)))
	if idx < 0 {
		idx = -idx
	}
	if enemy.IsBoss {
		idx = (idx + 3) % len(icons)
	}
	return icons[idx]
}

func enemyFileSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "_")
	slug = strings.ReplaceAll(slug, "—", "_")
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = strings.ReplaceAll(slug, "__", "_")
	return slug
}

type tapOverlay struct {
	widget.BaseWidget
	onTap func()
}

func newTapOverlay(onTap func()) *tapOverlay {
	t := &tapOverlay{onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapOverlay) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *tapOverlay) TappedSecondary(_ *fyne.PointEvent) {}

func (t *tapOverlay) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	return &tapOverlayRenderer{background: bg}
}

type tapOverlayRenderer struct {
	background *canvas.Rectangle
}

func (r *tapOverlayRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
}

func (r *tapOverlayRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *tapOverlayRenderer) Refresh() {
	r.background.Refresh()
}

func (r *tapOverlayRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background}
}

func (r *tapOverlayRenderer) Destroy() {}

func centerWithMaxWidth(content fyne.CanvasObject, maxWidth float32) fyne.CanvasObject {
	return container.New(&centeredMaxWidthLayout{maxWidth: maxWidth}, content)
}

func makeAchievementsGap(height float32) fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(0, height))
	return gap
}

type centeredMaxWidthLayout struct {
	maxWidth float32
}

func (l *centeredMaxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	if min.Width > l.maxWidth {
		min.Width = l.maxWidth
	}
	return min
}

func (l *centeredMaxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	width := size.Width
	if width > l.maxWidth {
		width = l.maxWidth
	}
	x := (size.Width - width) / 2
	objects[0].Move(fyne.NewPos(x, 0))
	objects[0].Resize(fyne.NewSize(width, size.Height))
}

func enemyLore(enemy models.DefeatedEnemy) string {
	if strings.TrimSpace(enemy.Description) != "" {
		return enemy.Description
	}
	if enemy.IsBoss {
		return fmt.Sprintf("%s удерживал контроль над зоной %d, пока вы не сломили его оборону.", enemy.Name, enemy.Zone)
	}
	return fmt.Sprintf("%s был устранён при зачистке зоны %d.", enemy.Name, enemy.Zone)
}

func zoneLabel(zone int) string {
	switch zone {
	case 1:
		return "Туманные Болота"
	case 2:
		return "Забытые Руины"
	case 3:
		return "Ледяные Пики"
	case 4:
		return "Пепельные Разломы"
	case 5:
		return "Цитадель Бездны"
	default:
		return "Неизвестный сектор"
	}
}
