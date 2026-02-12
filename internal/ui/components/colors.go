package components

import "image/color"

// Legacy color aliases — delegated to active theme tokens.
// These are kept as var so existing code does not break.
// New code should use T().Accent, T().Danger, etc. directly.
var (
	ColorBG           color.NRGBA
	ColorBGSecondary  color.NRGBA
	ColorBGCard       color.NRGBA
	ColorAccent       color.NRGBA
	ColorAccentBright color.NRGBA
	ColorText         color.NRGBA
	ColorTextDim      color.NRGBA
	ColorGold         color.NRGBA
	ColorRed          color.NRGBA
	ColorGreen        color.NRGBA
	ColorBlue         color.NRGBA
	ColorPurple       color.NRGBA
	ColorYellow       color.NRGBA
	ColorOrange       color.NRGBA
)

func init() {
	SyncLegacyColors()
}

// SyncLegacyColors refreshes the legacy Color* vars from the active theme.
// Call this after SetTheme() so all existing code picks up new colors.
func SyncLegacyColors() {
	t := T()
	ColorBG = t.BG
	ColorBGSecondary = t.BGPanel
	ColorBGCard = t.BGCard
	ColorAccent = t.AccentDim
	ColorAccentBright = t.Accent
	ColorText = t.Text
	ColorTextDim = t.TextSecondary
	ColorGold = t.Gold
	ColorRed = t.Danger
	ColorGreen = t.Success
	ColorBlue = t.Blue
	ColorPurple = t.Purple
	ColorYellow = t.Warning
	ColorOrange = t.Orange
}
