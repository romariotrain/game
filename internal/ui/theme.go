package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// SoloLevelingTheme - dark theme inspired by Solo Leveling
type SoloLevelingTheme struct{}

var _ fyne.Theme = (*SoloLevelingTheme)(nil)

// Color palette
var (
	ColorBG           = color.NRGBA{R: 15, G: 15, B: 25, A: 255}
	ColorBGSecondary  = color.NRGBA{R: 22, G: 22, B: 38, A: 255}
	ColorBGCard       = color.NRGBA{R: 28, G: 28, B: 48, A: 255}
	ColorAccent       = color.NRGBA{R: 100, G: 80, B: 220, A: 255}
	ColorAccentBright = color.NRGBA{R: 130, G: 100, B: 255, A: 255}
	ColorText         = color.NRGBA{R: 220, G: 220, B: 240, A: 255}
	ColorTextDim      = color.NRGBA{R: 140, G: 140, B: 170, A: 255}
	ColorGold         = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	ColorRed          = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
	ColorGreen        = color.NRGBA{R: 50, G: 200, B: 80, A: 255}
	ColorBlue         = color.NRGBA{R: 70, G: 130, B: 220, A: 255}
	ColorPurple       = color.NRGBA{R: 155, G: 89, B: 182, A: 255}
)

func (t *SoloLevelingTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBG
	case theme.ColorNameButton:
		return ColorAccent
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 50, G: 50, B: 70, A: 255}
	case theme.ColorNameForeground:
		return ColorText
	case theme.ColorNamePlaceHolder:
		return ColorTextDim
	case theme.ColorNameHover:
		return color.NRGBA{R: 40, G: 40, B: 65, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 20, G: 20, B: 35, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 60, G: 60, B: 90, A: 255}
	case theme.ColorNamePrimary:
		return ColorAccent
	case theme.ColorNameFocus:
		return ColorAccentBright
	case theme.ColorNameSelection:
		return color.NRGBA{R: 60, G: 50, B: 120, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 45, G: 45, B: 70, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 10, G: 10, B: 20, A: 230}
	case theme.ColorNameMenuBackground:
		return ColorBGSecondary
	case theme.ColorNameHeaderBackground:
		return ColorBGCard
	case theme.ColorNameDisabled:
		return ColorTextDim
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
