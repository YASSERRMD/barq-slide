package theme

import (
	"fmt"
	"math"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

const (
	// wcagAANormal is the minimum contrast ratio for normal text (WCAG AA).
	wcagAANormal = 4.5
	// wcagAALarge is the minimum contrast ratio for large text (18pt+) (WCAG AA).
	wcagAALarge = 3.0
	// maxAdjustmentSteps limits the darkening/lightening iterations.
	maxAdjustmentSteps = 20
	// adjustmentStep is the HSV value delta per iteration.
	adjustmentStep = 0.05
)

// ValidateContrast checks and adjusts DesignTokens to meet WCAG AA requirements.
// It modifies the tokens in-place, darkening or lightening text colours as needed.
func ValidateContrast(tokens *barqv1.DesignTokens) error {
	// 1. On-background text vs background.
	adjusted, err := ensureContrast(
		tokens.GetOnBackgroundHex(),
		tokens.GetBackgroundHex(),
		wcagAANormal,
	)
	if err != nil {
		return fmt.Errorf("wcag: on-background vs background: %w", err)
	}
	tokens.OnBackgroundHex = adjusted

	// 2. On-primary text vs primary.
	adjusted, err = ensureContrast(
		tokens.GetOnPrimaryHex(),
		tokens.GetPrimaryHex(),
		wcagAANormal,
	)
	if err != nil {
		return fmt.Errorf("wcag: on-primary vs primary: %w", err)
	}
	tokens.OnPrimaryHex = adjusted

	// 3. Muted text vs background (large text threshold is acceptable).
	adjusted, err = ensureContrast(
		tokens.GetMutedHex(),
		tokens.GetBackgroundHex(),
		wcagAALarge,
	)
	if err != nil {
		return fmt.Errorf("wcag: muted vs background: %w", err)
	}
	tokens.MutedHex = adjusted

	return nil
}

// ensureContrast adjusts foreground until it meets the minimum contrast ratio
// against the background. Returns the adjusted hex string.
func ensureContrast(fgHex, bgHex string, minRatio float64) (string, error) {
	ratio, err := ContrastRatio(fgHex, bgHex)
	if err != nil {
		return fgHex, fmt.Errorf("computing contrast ratio: %w", err)
	}

	if ratio >= minRatio {
		return fgHex, nil // Already compliant.
	}

	// Determine whether to darken or lighten the foreground.
	fgR, fgG, fgB, err := hexToRGB(fgHex)
	if err != nil {
		return fgHex, err
	}
	bgR, bgG, bgB, err := hexToRGB(bgHex)
	if err != nil {
		return fgHex, err
	}

	fgLum := relativeLuminance(fgR, fgG, fgB)
	bgLum := relativeLuminance(bgR, bgG, bgB)

	// If background is lighter than foreground, we darken foreground; otherwise lighten.
	shouldDarken := bgLum > fgLum

	h, s, v := rgbToHSV(fgR, fgG, fgB)

	for i := 0; i < maxAdjustmentSteps; i++ {
		if shouldDarken {
			v = math.Max(0, v-adjustmentStep)
		} else {
			v = math.Min(1, v+adjustmentStep)
		}

		nr, ng, nb := hsvToRGB(h, s, v)
		candidate := fmt.Sprintf("#%02X%02X%02X", nr, ng, nb)

		r, err := ContrastRatio(candidate, bgHex)
		if err != nil {
			return fgHex, err
		}

		if r >= minRatio {
			return candidate, nil
		}
	}

	// Fallback to pure black or white.
	if shouldDarken {
		return "#000000", nil
	}
	return "#FFFFFF", nil
}

// ContrastReport summarises the contrast check for all token pairs.
type ContrastReport struct {
	Pairs []ContrastPair
	Pass  bool
}

// ContrastPair holds a single foreground/background contrast measurement.
type ContrastPair struct {
	Name      string
	Foreground string
	Background string
	Ratio     float64
	Threshold float64
	Pass      bool
}

// AuditContrast produces a human-readable contrast report for the tokens.
func AuditContrast(tokens *barqv1.DesignTokens) (*ContrastReport, error) {
	type check struct {
		name  string
		fg    string
		bg    string
		thresh float64
	}

	checks := []check{
		{"body text / background", tokens.GetOnBackgroundHex(), tokens.GetBackgroundHex(), wcagAANormal},
		{"on-primary / primary", tokens.GetOnPrimaryHex(), tokens.GetPrimaryHex(), wcagAANormal},
		{"muted / background", tokens.GetMutedHex(), tokens.GetBackgroundHex(), wcagAALarge},
		{"secondary / background", tokens.GetSecondaryHex(), tokens.GetBackgroundHex(), wcagAALarge},
	}

	report := &ContrastReport{Pass: true}

	for _, c := range checks {
		ratio, err := ContrastRatio(c.fg, c.bg)
		if err != nil {
			return nil, fmt.Errorf("audit %s: %w", c.name, err)
		}
		pass := ratio >= c.thresh
		if !pass {
			report.Pass = false
		}
		report.Pairs = append(report.Pairs, ContrastPair{
			Name:      c.name,
			Foreground: c.fg,
			Background: c.bg,
			Ratio:     ratio,
			Threshold: c.thresh,
			Pass:      pass,
		})
	}

	return report, nil
}
