package components

import (
	"image/color"

	"solo-leveling/internal/models"
)

// QuestDesignTokens holds local UI tokens for quest/HUD rendering.
// Spacing follows the global 8/12/16/24 scale from tokens.go.
type QuestDesignTokens struct {
	PanelBg           color.NRGBA
	PanelBorder       color.NRGBA
	PanelBorderAccent color.NRGBA
	PanelBorderSoft   color.NRGBA
	TextPrimary       color.NRGBA
	TextMuted         color.NRGBA
	Accent            color.NRGBA
	Accent2           color.NRGBA
	Divider           color.NRGBA
	CornerRadiusPanel float32
	CornerRadiusBtn   float32
}

func QuestTokens() QuestDesignTokens {
	t := T()
	return QuestDesignTokens{
		PanelBg:           t.BGCardHover,
		PanelBorder:       t.Border,
		PanelBorderAccent: t.BorderActive,
		PanelBorderSoft:   color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 120},
		TextPrimary:       t.Text,
		TextMuted:         t.TextSecondary,
		Accent:            t.Accent,
		Accent2:           t.Gold,
		Divider:           color.NRGBA{R: t.Border.R, G: t.Border.G, B: t.Border.B, A: 90},
		CornerRadiusPanel: 12,
		CornerRadiusBtn:   8,
	}
}

// QuestRankColor returns muted->bright rank accents for HUD cards.
func QuestRankColor(rank models.QuestRank) color.NRGBA {
	switch rank {
	case models.RankE:
		return color.NRGBA{R: 86, G: 102, B: 120, A: 255}
	case models.RankD:
		return color.NRGBA{R: 64, G: 130, B: 180, A: 255}
	case models.RankC:
		return color.NRGBA{R: 0, G: 176, B: 212, A: 255}
	case models.RankB:
		return color.NRGBA{R: 0, G: 191, B: 140, A: 255}
	case models.RankA:
		return color.NRGBA{R: 255, G: 152, B: 0, A: 255}
	case models.RankS:
		return color.NRGBA{R: 255, G: 214, B: 0, A: 255}
	default:
		return T().Accent
	}
}
