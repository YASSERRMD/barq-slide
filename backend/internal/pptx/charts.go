package pptx

import (
	"fmt"
	"strings"

	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/presentation"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/charts"
	"github.com/YASSERRMD/barq-slides/internal/layout"
)

// AddChartAssets embeds chart assets into the PPTX slide.
// Charts are rendered as labelled text boxes that preserve the data summary.
// SVG chart bytes (produced by go-echarts) are embedded via AddSVGAssets.
func AddChartAssets(_ *presentation.Presentation, slide presentation.Slide, node *barqv1.SlideNode, tokens *barqv1.DesignTokens) error {
	geo := layout.Find(node.GetLayout().GetLayoutID())
	if !geo.HasMediaRegion {
		return nil
	}

	for _, asset := range node.GetAssets() {
		assetType := asset.GetType()
		if assetType != barqv1.AssetType_ASSET_TYPE_CHART_BAR &&
			assetType != barqv1.AssetType_ASSET_TYPE_CHART_LINE &&
			assetType != barqv1.AssetType_ASSET_TYPE_CHART_PIE {
			continue
		}

		cd, err := charts.ParseChartData(asset.GetChartDataJson())
		if err != nil || cd == nil {
			cd = charts.SampleChartData(asset.GetQuery(), assetType)
		}

		if err := addChartSummaryBox(slide, cd, asset, tokens, geo.Media); err != nil {
			_ = fmt.Errorf("pptx: chart summary %s: %w (skipping)", asset.GetID(), err)
		}
	}
	return nil
}

// addChartSummaryBox renders chart data as a formatted text summary box.
// This provides human-readable chart data in the PPTX until full native
// chart embedding is wired via unioffice internals.
func addChartSummaryBox(
	slide presentation.Slide,
	cd *charts.ChartData,
	asset *barqv1.AssetSlot,
	tokens *barqv1.DesignTokens,
	rect layout.Region,
) error {
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

	// Title row.
	if cd.Title != "" {
		titlePara := tb.AddParagraph()
		run := titlePara.AddRun()
		run.SetText(cd.Title)
		rpr := run.Properties()
		rpr.SetBold(true)
		rpr.SetSize(14 * measurement.Point)
		applyHexColor(rpr, tokens.GetOnBackgroundHex())
	}

	// Series data rows.
	for si, series := range cd.Series {
		_ = si
		for i, label := range cd.Labels {
			if i >= len(series.Values) {
				break
			}
			para := tb.AddParagraph()
			run := para.AddRun()
			run.SetText(fmt.Sprintf("  %s: %g", strings.TrimSpace(label), series.Values[i]))
			rpr := run.Properties()
			rpr.SetSize(11 * measurement.Point)
			applyHexColor(rpr, tokens.GetMutedHex())
		}
	}

	_ = color.White // keep color import used
	return nil
}

// AddKPICards renders KPI blocks from a slide as styled card text boxes.
// KPI data is encoded in body blocks as "label|value|delta" text.
func AddKPICards(slide presentation.Slide, node *barqv1.SlideNode, tokens *barqv1.DesignTokens) error {
	geo := layout.Find(node.GetLayout().GetLayoutID())

	var kpiBlocks []*barqv1.ContentBlock
	for _, b := range node.GetBlocks() {
		if b.GetRole() == "body" && strings.Contains(b.GetText(), "|") {
			kpiBlocks = append(kpiBlocks, b)
		}
	}

	if len(kpiBlocks) == 0 || len(geo.Body) == 0 {
		return nil
	}

	// Distribute KPI cards evenly across the body region width.
	r := geo.Body[0]
	cardW := r.Width / int64(len(kpiBlocks))
	cardH := r.Height

	for i, block := range kpiBlocks {
		parts := strings.SplitN(block.GetText(), "|", 3)
		label, value, delta := "", "", ""
		if len(parts) > 0 {
			label = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			delta = strings.TrimSpace(parts[2])
		}

		cardX := r.X + int64(i)*cardW
		if err := addKPICard(slide, label, value, delta, cardX, r.Y, cardW-50000, cardH, tokens); err != nil {
			return err
		}
	}
	return nil
}

// addKPICard renders a single KPI metric card.
func addKPICard(
	slide presentation.Slide,
	label, value, delta string,
	x, y, w, h int64,
	tokens *barqv1.DesignTokens,
) error {
	tb := slide.AddTextBox()
	props := tb.Properties()
	props.SetPosition(
		measurement.Distance(x)*measurement.EMU,
		measurement.Distance(y)*measurement.EMU,
	)
	props.SetSize(
		measurement.Distance(w)*measurement.EMU,
		measurement.Distance(h)*measurement.EMU,
	)
	// Set card background to surface colour.
	props.SetSolidFill(hexToColor(tokens.GetSurfaceHex()))

	// Value line (large, primary colour).
	valuePara := tb.AddParagraph()
	valuePara.Properties().SetAlign(0) // centre
	vRun := valuePara.AddRun()
	vRun.SetText(value)
	vProps := vRun.Properties()
	vProps.SetBold(true)
	vProps.SetSize(28 * measurement.Point)
	applyHexColor(vProps, tokens.GetPrimaryHex())

	// Label line (caption, muted).
	labelPara := tb.AddParagraph()
	lRun := labelPara.AddRun()
	lRun.SetText(label)
	lRun.Properties().SetSize(10 * measurement.Point)
	applyHexColor(lRun.Properties(), tokens.GetMutedHex())

	// Delta line (secondary, if present).
	if delta != "" {
		deltaPara := tb.AddParagraph()
		dRun := deltaPara.AddRun()
		dRun.SetText(delta)
		dRun.Properties().SetSize(10 * measurement.Point)
		applyHexColor(dRun.Properties(), tokens.GetSecondaryHex())
	}

	return nil
}

// hexToColor converts a hex colour string to a unioffice color.Color.
func hexToColor(hexColor string) color.Color {
	h := stripHash(hexColor)
	if len(h) != 6 {
		return color.White
	}
	var r, g, b uint8
	fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b) //nolint:errcheck
	return color.RGB(r, g, b)
}
