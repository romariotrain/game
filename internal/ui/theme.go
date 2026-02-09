package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"solo-leveling/internal/ui/components"
)

// SoloLevelingTheme - dark theme inspired by Solo Leveling
type SoloLevelingTheme struct{}

var _ fyne.Theme = (*SoloLevelingTheme)(nil)

func (t *SoloLevelingTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return components.ColorBG
	case theme.ColorNameButton:
		return components.ColorAccent
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 50, G: 50, B: 70, A: 255}
	case theme.ColorNameForeground:
		return components.ColorText
	case theme.ColorNamePlaceHolder:
		return components.ColorTextDim
	case theme.ColorNameHover:
		return color.NRGBA{R: 40, G: 40, B: 65, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 20, G: 20, B: 35, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 60, G: 60, B: 90, A: 255}
	case theme.ColorNamePrimary:
		return components.ColorAccent
	case theme.ColorNameFocus:
		return components.ColorAccentBright
	case theme.ColorNameSelection:
		return color.NRGBA{R: 60, G: 50, B: 120, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 45, G: 45, B: 70, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 10, G: 10, B: 20, A: 230}
	case theme.ColorNameMenuBackground:
		return components.ColorBGSecondary
	case theme.ColorNameHeaderBackground:
		return components.ColorBGCard
	case theme.ColorNameDisabled:
		return components.ColorTextDim
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 60, B: 90, A: 180}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *SoloLevelingTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *SoloLevelingTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *SoloLevelingTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameInputBorder:
		return 1
	default:
		return theme.DefaultTheme().Size(name)
	}
}
