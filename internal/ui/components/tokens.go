package components

import "image/color"

// ThemeTokens defines all design tokens for a visual theme.
type ThemeTokens struct {
	// Backgrounds
	BG          color.NRGBA
	BGPanel     color.NRGBA
	BGCard      color.NRGBA
	BGCardHover color.NRGBA

	// Borders
	Border       color.NRGBA
	BorderActive color.NRGBA

	// Accent
	Accent     color.NRGBA
	AccentDim  color.NRGBA
	AccentGlow color.NRGBA // used for corner brackets and subtle glow

	// Text
	Text          color.NRGBA
	TextSecondary color.NRGBA
	TextMuted     color.NRGBA

	// Semantic
	Gold       color.NRGBA
	GoldDim    color.NRGBA
	Danger     color.NRGBA
	DangerDim  color.NRGBA
	Success    color.NRGBA
	SuccessDim color.NRGBA
	Warning    color.NRGBA
	Blue       color.NRGBA
	Purple     color.NRGBA
	Orange     color.NRGBA

	// Stat colors
	StatSTR color.NRGBA
	StatAGI color.NRGBA
	StatINT color.NRGBA
	StatSTA color.NRGBA

	// Decorative
	CornerBrackets  bool
	HeaderUppercase bool
	HeaderLetterGap string // e.g. "  " between chars for logo
}

// ----------------------------------------------------------------
// Theme presets
// ----------------------------------------------------------------

// SystemTheme — cold cyan HUD, futuristic scanner feel.
var SystemTheme = ThemeTokens{
	BG:          color.NRGBA{R: 10, G: 14, B: 23, A: 255},
	BGPanel:     color.NRGBA{R: 15, G: 21, B: 32, A: 255},
	BGCard:      color.NRGBA{R: 20, G: 28, B: 43, A: 255},
	BGCardHover: color.NRGBA{R: 26, G: 36, B: 54, A: 255},

	Border:       color.NRGBA{R: 30, G: 42, B: 58, A: 255},
	BorderActive: color.NRGBA{R: 42, G: 58, B: 80, A: 255},

	Accent:     color.NRGBA{R: 0, G: 212, B: 255, A: 255},
	AccentDim:  color.NRGBA{R: 0, G: 136, B: 170, A: 255},
	AccentGlow: color.NRGBA{R: 0, G: 212, B: 255, A: 64},

	Text:          color.NRGBA{R: 224, G: 230, B: 240, A: 255},
	TextSecondary: color.NRGBA{R: 122, G: 139, B: 160, A: 255},
	TextMuted:     color.NRGBA{R: 74, G: 85, B: 104, A: 255},

	Gold:       color.NRGBA{R: 255, G: 184, B: 0, A: 255},
	GoldDim:    color.NRGBA{R: 170, G: 122, B: 0, A: 255},
	Danger:     color.NRGBA{R: 255, G: 59, B: 92, A: 255},
	DangerDim:  color.NRGBA{R: 170, G: 32, B: 64, A: 255},
	Success:    color.NRGBA{R: 0, G: 230, B: 118, A: 255},
	SuccessDim: color.NRGBA{R: 0, G: 153, B: 80, A: 255},
	Warning:    color.NRGBA{R: 255, G: 179, B: 0, A: 255},
	Blue:       color.NRGBA{R: 64, G: 196, B: 255, A: 255},
	Purple:     color.NRGBA{R: 179, G: 136, B: 255, A: 255},
	Orange:     color.NRGBA{R: 255, G: 152, B: 0, A: 255},

	StatSTR: color.NRGBA{R: 255, G: 82, B: 82, A: 255},
	StatAGI: color.NRGBA{R: 64, G: 196, B: 255, A: 255},
	StatINT: color.NRGBA{R: 179, G: 136, B: 255, A: 255},
	StatSTA: color.NRGBA{R: 105, G: 240, B: 174, A: 255},

	CornerBrackets:  true,
	HeaderUppercase: true,
	HeaderLetterGap: "  ",
}

// ClassicTheme — warm purple, mystical/fantasy feel (original palette preserved).
var ClassicTheme = ThemeTokens{
	BG:          color.NRGBA{R: 15, G: 15, B: 25, A: 255},
	BGPanel:     color.NRGBA{R: 22, G: 22, B: 38, A: 255},
	BGCard:      color.NRGBA{R: 28, G: 28, B: 48, A: 255},
	BGCardHover: color.NRGBA{R: 38, G: 35, B: 58, A: 255},

	Border:       color.NRGBA{R: 45, G: 45, B: 70, A: 255},
	BorderActive: color.NRGBA{R: 60, G: 50, B: 120, A: 255},

	Accent:     color.NRGBA{R: 130, G: 100, B: 255, A: 255},
	AccentDim:  color.NRGBA{R: 100, G: 80, B: 220, A: 255},
	AccentGlow: color.NRGBA{R: 130, G: 100, B: 255, A: 64},

	Text:          color.NRGBA{R: 220, G: 220, B: 240, A: 255},
	TextSecondary: color.NRGBA{R: 140, G: 140, B: 170, A: 255},
	TextMuted:     color.NRGBA{R: 90, G: 90, B: 120, A: 255},

	Gold:       color.NRGBA{R: 255, G: 215, B: 0, A: 255},
	GoldDim:    color.NRGBA{R: 180, G: 150, B: 0, A: 255},
	Danger:     color.NRGBA{R: 220, G: 50, B: 50, A: 255},
	DangerDim:  color.NRGBA{R: 150, G: 30, B: 30, A: 255},
	Success:    color.NRGBA{R: 50, G: 200, B: 80, A: 255},
	SuccessDim: color.NRGBA{R: 30, G: 130, B: 50, A: 255},
	Warning:    color.NRGBA{R: 240, G: 200, B: 40, A: 255},
	Blue:       color.NRGBA{R: 70, G: 130, B: 220, A: 255},
	Purple:     color.NRGBA{R: 155, G: 89, B: 182, A: 255},
	Orange:     color.NRGBA{R: 230, G: 126, B: 34, A: 255},

	StatSTR: color.NRGBA{R: 220, G: 60, B: 60, A: 255},
	StatAGI: color.NRGBA{R: 220, G: 180, B: 40, A: 255},
	StatINT: color.NRGBA{R: 60, G: 120, B: 220, A: 255},
	StatSTA: color.NRGBA{R: 50, G: 180, B: 70, A: 255},

	CornerBrackets:  false,
	HeaderUppercase: false,
	HeaderLetterGap: " ",
}

// ----------------------------------------------------------------
// Active theme + accessor
// ----------------------------------------------------------------

// activeTheme is the currently active theme. Default to System.
var activeTheme = &SystemTheme

// T returns the active theme tokens.
func T() *ThemeTokens {
	return activeTheme
}

// SetTheme switches the active theme.
func SetTheme(t *ThemeTokens) {
	activeTheme = t
}

// ----------------------------------------------------------------
// Spacing, radius and typography tokens
// ----------------------------------------------------------------

const (
	SpaceXS  float32 = 4
	SpaceSM  float32 = 8
	SpaceMD  float32 = 12
	SpaceLG  float32 = 16
	SpaceXL  float32 = 24
	SpaceXXL float32 = 32

	RadiusSM float32 = 4
	RadiusMD float32 = 6
	RadiusLG float32 = 10
	RadiusXL float32 = 14

	BorderThin   float32 = 1
	BorderMedium float32 = 1.5
	BorderThick  float32 = 2

	// Typography sizes
	TextHeadingXL float32 = 24
	TextHeadingLG float32 = 18
	TextHeadingMD float32 = 15
	TextHeadingSM float32 = 13
	TextBodyLG    float32 = 15
	TextBodyMD    float32 = 13
	TextBodySM    float32 = 11
	TextNumberXL  float32 = 28
	TextNumberLG  float32 = 20
	TextNumberMD  float32 = 15
	TextNumberSM  float32 = 12
)
