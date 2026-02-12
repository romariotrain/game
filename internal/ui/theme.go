package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"solo-leveling/internal/ui/components"
)

// SoloLevelingTheme - dark theme driven by components.T() tokens.
type SoloLevelingTheme struct{}

var _ fyne.Theme = (*SoloLevelingTheme)(nil)

func (th *SoloLevelingTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	t := components.T()
	switch name {
	case theme.ColorNameBackground:
		return t.BG
	case theme.ColorNameButton:
		return t.AccentDim
	case theme.ColorNameDisabledButton:
		return t.BGCardHover
	case theme.ColorNameForeground:
		return t.Text
	case theme.ColorNamePlaceHolder:
		return t.TextSecondary
	case theme.ColorNameHover:
		return t.BGCardHover
	case theme.ColorNameInputBackground:
		return t.BGPanel
	case theme.ColorNameInputBorder:
		return t.Border
	case theme.ColorNamePrimary:
		return t.Accent
	case theme.ColorNameFocus:
		return t.Accent
	case theme.ColorNameSelection:
		return t.BorderActive
	case theme.ColorNameSeparator:
		return t.Border
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: t.BG.R, G: t.BG.G, B: t.BG.B, A: 230}
	case theme.ColorNameMenuBackground:
		return t.BGPanel
	case theme.ColorNameHeaderBackground:
		return t.BGCard
	case theme.ColorNameDisabled:
		return t.TextMuted
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 180}
	case theme.ColorNameError:
		return t.Danger
	case theme.ColorNameSuccess:
		return t.Success
	case theme.ColorNameWarning:
		return t.Warning
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (th *SoloLevelingTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (th *SoloLevelingTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (th *SoloLevelingTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return components.SpaceMD / 2 // 6px
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameText:
		return components.TextBodyMD
	case theme.SizeNameHeadingText:
		return components.TextHeadingXL
	case theme.SizeNameSubHeadingText:
		return components.TextHeadingLG
	case theme.SizeNameInputBorder:
		return components.BorderThin
	default:
		return theme.DefaultTheme().Size(name)
	}
}
