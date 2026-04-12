package pptx

import (
	"fmt"
	"html"

	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/drawing"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/schema/soo/dml"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/layout"
)

// AddSlide appends a unioffice slide built from a SlideNode to the presentation.
// EMU coordinates from the layout Geometry are mapped directly to p:sp shapes.
func AddSlide(prs *presentation.Presentation, node *barqv1.SlideNode, tokens *barqv1.DesignTokens) error {
	masters := prs.SlideMasters()
	if len(masters) == 0 {
		return fmt.Errorf("pptx: no slide masters")
	}

	layouts := masters[0].SlideLayouts()
	if len(layouts) == 0 {
		return fmt.Errorf("pptx: no slide layouts")
	}

	slide, err := prs.AddDefaultSlideWithLayout(layouts[0])
	if err != nil {
		return fmt.Errorf("pptx: adding slide %d: %w", node.GetIndex(), err)
	}

	// Clear all placeholder shapes — we draw everything from scratch.
	for _, ph := range slide.PlaceHolders() {
		_ = ph.Remove()
	}

	geo := layout.Find(node.GetLayout().GetLayoutId())

	// ── Heading text shape ──────────────────────────────────────────────────
	for _, block := range node.GetBlocks() {
		if block.GetRole() == "heading" && geo.Title.Width > 0 {
			addTextBox(slide, block.GetText(), geo.Title,
				tokens.GetHeading(), tokens.GetOnBackgroundHex(), true)
		}
		if block.GetRole() == "subtitle" && geo.Subtitle.Width > 0 {
			addTextBox(slide, block.GetText(), geo.Subtitle,
				tokens.GetSubheading(), tokens.GetMutedHex(), false)
		}
	}

	// ── Bullet body shape ───────────────────────────────────────────────────
	var bullets []string
	for _, block := range node.GetBlocks() {
		if block.GetRole() == "body" {
			bullets = append(bullets, html.UnescapeString(block.GetText()))
		}
	}
	if len(bullets) > 0 && len(geo.Body) > 0 {
		addBulletBox(slide, bullets, geo.Body[0],
			tokens.GetBody(), tokens.GetOnBackgroundHex())
	}

	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func addTextBox(
	slide presentation.Slide,
	text string,
	rect layout.Region,
	typo *barqv1.Typography,
	hexColor string,
	bold bool,
) {
	tb := slide.AddTextBox()
	props := tb.Properties()
	props.SetPosition(
		measurement.Distance(rect.X)*measurement.EMU,
		measurement.Distance(rect.Y)*measurement.EMU,
	)
	props.SetSize(
		measurement.Distance(rect.Width)*measurement.EMU,
		measurement.Distance(rect.Height)*measurement.EMU,
	)

	para := tb.AddParagraph()
	run := para.AddRun()
	run.SetText(text)

	rpr := run.Properties()
	if bold {
		rpr.SetBold(true)
	}
	if typo != nil && typo.GetSizePt() > 0 {
		rpr.SetSize(measurement.Distance(typo.GetSizePt()) * measurement.Point)
	}
	if typo != nil && typo.GetFontFamily() != "" {
		rpr.SetFont(typo.GetFontFamily())
	}
	applyHexColor(rpr, hexColor)
}

func addBulletBox(
	slide presentation.Slide,
	bullets []string,
	rect layout.Region,
	typo *barqv1.Typography,
	hexColor string,
) {
	tb := slide.AddTextBox()
	props := tb.Properties()
	props.SetPosition(
		measurement.Distance(rect.X)*measurement.EMU,
		measurement.Distance(rect.Y)*measurement.EMU,
	)
	props.SetSize(
		measurement.Distance(rect.Width)*measurement.EMU,
		measurement.Distance(rect.Height)*measurement.EMU,
	)

	for _, bullet := range bullets {
		para := tb.AddParagraph()
		// Use a filled circle (•) as the bullet character.
		para.Properties().SetBulletChar("•")

		run := para.AddRun()
		run.SetText(bullet)

		rpr := run.Properties()
		if typo != nil && typo.GetSizePt() > 0 {
			rpr.SetSize(measurement.Distance(typo.GetSizePt()) * measurement.Point)
		}
		if typo != nil && typo.GetFontFamily() != "" {
			rpr.SetFont(typo.GetFontFamily())
		}
		applyHexColor(rpr, hexColor)
	}
}

// applyHexColor sets a solid RGB fill on a RunProperties via the unioffice
// color.Color helper, which requires the 6-digit hex without '#'.
func applyHexColor(rpr drawing.RunProperties, hexColor string) {
	hexColor = stripHash(hexColor)
	if len(hexColor) != 6 {
		return
	}
	var r, g, b uint8
	fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b) //nolint:errcheck
	rpr.SetSolidFill(color.RGB(r, g, b))
}

// Ensure dml is used (CT_TextBodyProperties is referenced in p:sp).
var _ = dml.NewCT_TextBodyProperties
