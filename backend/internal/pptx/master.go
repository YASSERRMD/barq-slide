// Package pptx assembles native OpenXML (.pptx) presentations from SlideNodes
// using the unioffice library.
package pptx

import (
	"fmt"

	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/schema/soo/dml"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// Assembler builds a unioffice Presentation from SlideNodes and DesignTokens.
type Assembler struct {
	tokens *barqv1.DesignTokens
	prs    *presentation.Presentation
}

// NewAssembler creates an Assembler, sets slide dimensions, and applies
// DesignTokens to the theme colour scheme.
func NewAssembler(tokens *barqv1.DesignTokens) (*Assembler, error) {
	prs := presentation.New()

	// Widescreen 16:9 at native PowerPoint EMU dimensions.
	ss := prs.SlideSize()
	ss.SetSize(presentation.SlideScreenSize16x9)

	a := &Assembler{tokens: tokens, prs: prs}
	if err := a.applyThemeColors(); err != nil {
		return nil, fmt.Errorf("pptx: apply theme colors: %w", err)
	}
	return a, nil
}

// applyThemeColors patches the first theme's colour scheme from DesignTokens so
// that native PowerPoint shapes inherit the brand colours automatically.
func (a *Assembler) applyThemeColors() error {
	themes := a.prs.Themes()
	if len(themes) == 0 {
		return nil // no theme to patch — not fatal
	}

	theme := themes[0]
	cs := theme.ThemeElements.ClrScheme // *dml.CT_ColorScheme

	primaryHex := stripHash(a.tokens.GetPrimaryHex())
	secondaryHex := stripHash(a.tokens.GetSecondaryHex())
	bgHex := stripHash(a.tokens.GetBackgroundHex())
	onBgHex := stripHash(a.tokens.GetOnBackgroundHex())

	if primaryHex != "" {
		cs.Accent1 = dml.NewCT_Color()
		cs.Accent1.SrgbClr = dml.NewCT_SRgbColor()
		cs.Accent1.SrgbClr.ValAttr = primaryHex
	}
	if secondaryHex != "" {
		cs.Accent2 = dml.NewCT_Color()
		cs.Accent2.SrgbClr = dml.NewCT_SRgbColor()
		cs.Accent2.SrgbClr.ValAttr = secondaryHex
	}
	// Lt1 maps to the "background" colour in PowerPoint's colour picker.
	if bgHex != "" {
		cs.Lt1 = dml.NewCT_Color()
		cs.Lt1.SrgbClr = dml.NewCT_SRgbColor()
		cs.Lt1.SrgbClr.ValAttr = bgHex
	}
	// Dk1 maps to the primary text colour.
	if onBgHex != "" {
		cs.Dk1 = dml.NewCT_Color()
		cs.Dk1.SrgbClr = dml.NewCT_SRgbColor()
		cs.Dk1.SrgbClr.ValAttr = onBgHex
	}

	return nil
}

// Presentation returns the underlying unioffice Presentation.
func (a *Assembler) Presentation() *presentation.Presentation { return a.prs }

// stripHash removes a leading '#' from a hex colour string.
func stripHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}
