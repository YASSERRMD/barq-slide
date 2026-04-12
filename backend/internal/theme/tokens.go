package theme

import (
	"fmt"
	"image/color"
	"math"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// EMU conversion constants.
const (
	emuPerInch      = 914400
	emuPerPt        = 12700
	slideWidthEMU   = 9144000  // 10 inches
	slideHeightEMU  = 5143500  // 5.625 inches (16:9)
)

// moodPalette defines the deterministic colour system for each mood.
type moodPalette struct {
	primary    string
	secondary  string
	background string
	surface    string
}

// palettes maps each mood to its colour values (all hex strings with #).
var palettes = map[Mood]moodPalette{
	MoodCorporate: {
		primary:    "#1A3C5E",
		secondary:  "#2E86AB",
		background: "#FFFFFF",
		surface:    "#F5F7FA",
	},
	MoodCreative: {
		primary:    "#6B2D8B",
		secondary:  "#FF6B6B",
		background: "#FAFAFA",
		surface:    "#F3E8FF",
	},
	MoodAcademic: {
		primary:    "#1B3A4B",
		secondary:  "#4A7FA5",
		background: "#FFFFFF",
		surface:    "#EEF2F7",
	},
	MoodTechnology: {
		primary:    "#0D1F2D",
		secondary:  "#00B4D8",
		background: "#0D1117",
		surface:    "#161B22",
	},
	MoodHealthcare: {
		primary:    "#005F73",
		secondary:  "#0A9396",
		background: "#FFFFFF",
		surface:    "#EEF9FA",
	},
	MoodFinance: {
		primary:    "#1C2B2D",
		secondary:  "#2ECC71",
		background: "#FFFFFF",
		surface:    "#F0FAF4",
	},
	MoodMarketing: {
		primary:    "#E63946",
		secondary:  "#FF9F1C",
		background: "#FFFFFF",
		surface:    "#FFF8F0",
	},
	MoodSustainable: {
		primary:    "#2D6A4F",
		secondary:  "#52B788",
		background: "#FFFFFF",
		surface:    "#F0FDF4",
	},
	MoodStartup: {
		primary:    "#7209B7",
		secondary:  "#F72585",
		background: "#FFFFFF",
		surface:    "#FDF4FF",
	},
	MoodGovernment: {
		primary:    "#1D3557",
		secondary:  "#457B9D",
		background: "#FFFFFF",
		surface:    "#F1F5F9",
	},
}

// fontPairings maps each mood to a heading/body font pair (OSS fonts).
type fontPairing struct {
	heading string
	body    string
}

var fontPairings = map[Mood]fontPairing{
	MoodCorporate:   {heading: "Inter", body: "Inter"},
	MoodCreative:    {heading: "Plus Jakarta Sans", body: "DM Sans"},
	MoodAcademic:    {heading: "Libre Baskerville", body: "Source Sans 3"},
	MoodTechnology:  {heading: "Space Grotesk", body: "JetBrains Mono"},
	MoodHealthcare:  {heading: "Nunito Sans", body: "Open Sans"},
	MoodFinance:     {heading: "IBM Plex Sans", body: "IBM Plex Sans"},
	MoodMarketing:   {heading: "Montserrat", body: "Lato"},
	MoodSustainable: {heading: "DM Serif Display", body: "DM Sans"},
	MoodStartup:     {heading: "Plus Jakarta Sans", body: "Inter"},
	MoodGovernment:  {heading: "Source Serif 4", body: "Source Sans 3"},
}

// GenerateTokens produces a full DesignTokens from a mood.
func GenerateTokens(mood Mood) *barqv1.DesignTokens {
	palette := palettes[mood]
	if palette.primary == "" {
		palette = palettes[MoodCorporate]
	}

	fonts := fontPairings[mood]
	if fonts.heading == "" {
		fonts = fontPairings[MoodCorporate]
	}

	isDark := mood == MoodTechnology

	onPrimary := "#FFFFFF"
	onBackground := "#111111"
	if isDark {
		onBackground = "#E6EDF3"
	}

	mutedHex := "#6B7280"
	if isDark {
		mutedHex = "#8B949E"
	}

	return &barqv1.DesignTokens{
		PrimaryHex:      palette.primary,
		SecondaryHex:    palette.secondary,
		BackgroundHex:   palette.background,
		SurfaceHex:      palette.surface,
		OnPrimaryHex:    onPrimary,
		OnBackgroundHex: onBackground,
		MutedHex:        mutedHex,
		Mood:            string(mood),

		Heading: &barqv1.Typography{
			FontFamily:  fonts.heading,
			SizePt:      36,
			Weight:      700,
			LineHeight:  1.2,
			IsUppercase: false,
		},
		Subheading: &barqv1.Typography{
			FontFamily:  fonts.heading,
			SizePt:      24,
			Weight:      600,
			LineHeight:  1.3,
			IsUppercase: false,
		},
		Body: &barqv1.Typography{
			FontFamily:  fonts.body,
			SizePt:      16,
			Weight:      400,
			LineHeight:  1.6,
			IsUppercase: false,
		},
		Caption: &barqv1.Typography{
			FontFamily:  fonts.body,
			SizePt:      11,
			Weight:      400,
			LineHeight:  1.4,
			IsUppercase: false,
		},

		BorderRadiusEmu: int32(6 * emuPerPt),
		GutterEmu:       int32(emuPerInch / 4),
	}
}

// hexToRGB parses a hex colour string (#RRGGBB) into RGB components.
func hexToRGB(hex string) (r, g, b uint8, err error) {
	hex = trimHash(hex)
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex colour: #%s", hex)
	}
	var ri, gi, bi uint64
	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parsing hex colour: %w", err)
	}
	return uint8(ri), uint8(gi), uint8(bi), nil
}

func trimHash(s string) string {
	if len(s) > 0 && s[0] == '#' {
		return s[1:]
	}
	return s
}

// relativeLuminance returns the WCAG relative luminance of an sRGB colour.
func relativeLuminance(r, g, b uint8) float64 {
	linearize := func(c uint8) float64 {
		v := float64(c) / 255.0
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b)
}

// ContrastRatio computes the WCAG contrast ratio between two hex colours.
func ContrastRatio(hex1, hex2 string) (float64, error) {
	r1, g1, b1, err := hexToRGB(hex1)
	if err != nil {
		return 0, err
	}
	r2, g2, b2, err := hexToRGB(hex2)
	if err != nil {
		return 0, err
	}

	l1 := relativeLuminance(r1, g1, b1)
	l2 := relativeLuminance(r2, g2, b2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)
	return (lighter + 0.05) / (darker + 0.05), nil
}

// lightenHex lightens a hex colour by the given factor (0.0–1.0).
func lightenHex(hex string, factor float64) (string, error) {
	r, g, b, err := hexToRGB(hex)
	if err != nil {
		return "", err
	}

	h, s, v := rgbToHSV(r, g, b)
	v = math.Min(1.0, v+factor)

	nr, ng, nb := hsvToRGB(h, s, v)
	return fmt.Sprintf("#%02X%02X%02X", nr, ng, nb), nil
}

func rgbToHSV(r, g, b uint8) (h, s, v float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	c := color.RGBA{R: r, G: g, B: b, A: 255}
	_ = c

	maxC := math.Max(rf, math.Max(gf, bf))
	minC := math.Min(rf, math.Min(gf, bf))
	delta := maxC - minC

	v = maxC
	if maxC == 0 {
		s = 0
	} else {
		s = delta / maxC
	}

	if delta == 0 {
		h = 0
	} else if maxC == rf {
		h = math.Mod((gf-bf)/delta, 6)
	} else if maxC == gf {
		h = (bf-rf)/delta + 2
	} else {
		h = (rf-gf)/delta + 4
	}

	h = h * 60
	if h < 0 {
		h += 360
	}
	return h, s, v
}

func hsvToRGB(h, s, v float64) (r, g, b uint8) {
	if s == 0 {
		val := uint8(v * 255)
		return val, val, val
	}

	h = h / 60
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var rf, gf, bf float64
	switch int(i) % 6 {
	case 0:
		rf, gf, bf = v, t, p
	case 1:
		rf, gf, bf = q, v, p
	case 2:
		rf, gf, bf = p, v, t
	case 3:
		rf, gf, bf = p, q, v
	case 4:
		rf, gf, bf = t, p, v
	case 5:
		rf, gf, bf = v, p, q
	}

	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}

// Suppress unused import warning for lightenHex.
var _ = lightenHex
