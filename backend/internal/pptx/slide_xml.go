package pptx

import (
	"fmt"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// Slide dimensions (EMU).
const (
	sW int64 = 9144000 // slide width
	sH int64 = 5143500 // slide height
	mL int64 = 500000  // margin left
	mR int64 = 500000  // margin right
	mT int64 = 450000  // margin top
	mB int64 = 380000  // margin bottom
	cW int64 = sW - mL - mR // 8144000
	gp int64 = 200000  // gutter
	hH int64 = 700000  // header height
	bT int64 = mT + hH + gp // body top
	bH int64 = sH - bT - mB // body height
	c2 int64 = (cW - gp) / 2
	c3 int64 = (cW - 2*gp) / 3
)

// generateSlideXML produces the full slide XML for a given SlideNode.
func generateSlideXML(node *barqv1.SlideNode, tokens *barqv1.DesignTokens) string {
	layoutId := barqv1.LayoutID_LAYOUT_ID_BULLET_LIST
	if node.GetLayout() != nil {
		layoutId = node.GetLayout().GetLayoutId()
	}

	// Color palette
	primary   := hex6(tokens.GetPrimaryHex())
	secondary := hex6(tokens.GetSecondaryHex())
	bg        := hex6(tokens.GetBackgroundHex())
	onBg      := hex6(tokens.GetOnBackgroundHex())
	muted     := hex6(tokens.GetMutedHex())
	surface   := hex6(tokens.GetSurfaceHex())

	if primary   == "111827" { primary   = "4F46E5" }
	if secondary == "111827" { secondary = "10B981" }
	if bg        == "111827" { bg        = "FFFFFF" }
	if onBg      == "111827" { onBg      = "111827" }
	if muted     == "111827" { muted     = "6B7280" }
	if surface   == "111827" { surface   = "F9FAFB" }

	if node.GetBgColorHex() != "" {
		bg = hex6(node.GetBgColorHex())
	}

	id := &idGen{n: 1}

	// Extract theme fonts; fall back to Calibri if tokens are sparse.
	headingFont := tokens.GetHeading().GetFontFamily()
	bodyFont := tokens.GetBody().GetFontFamily()

	var body strings.Builder
	switch layoutId {
	case barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER:
		body.WriteString(buildTitleCenter(id, node, primary, secondary, bg, onBg, muted, headingFont, bodyFont))
	case barqv1.LayoutID_LAYOUT_ID_SECTION_DIVIDER:
		body.WriteString(buildSectionDivider(id, node, primary, secondary, headingFont, bodyFont))
	case barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN:
		body.WriteString(buildTwoColumn(id, node, primary, secondary, bg, onBg, muted, surface, headingFont, bodyFont))
	case barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL:
		body.WriteString(buildQuote(id, node, primary, secondary, bg, onBg, muted, headingFont, bodyFont))
	case barqv1.LayoutID_LAYOUT_ID_KPI_CARDS:
		body.WriteString(buildKPICards(id, node, primary, secondary, bg, onBg, muted, surface, headingFont, bodyFont))
	case barqv1.LayoutID_LAYOUT_ID_CLOSING:
		body.WriteString(buildClosing(id, node, primary, secondary, headingFont, bodyFont))
	default:
		body.WriteString(buildBulletList(id, node, primary, secondary, bg, onBg, muted, headingFont, bodyFont))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<p:cSld><p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>
%s
</p:spTree></p:cSld>
</p:sld>`, body.String())
}

// ── layout builders ───────────────────────────────────────────────────────────

func buildTitleCenter(id *idGen, node *barqv1.SlideNode, primary, secondary, bg, onBg, muted, headingFont, bodyFont string) string {
	var b strings.Builder
	// Background
	b.WriteString(rectShape(id.next(), 0, 0, sW, sH, bg))
	// Top gradient bar
	b.WriteString(gradRect(id.next(), 0, 0, sW, 90000, primary, secondary))
	// Decorative circles bottom-right
	b.WriteString(ovalShape(id.next(), sW-2200000, sH-2200000, 3400000, 3400000, primary, 6000))
	b.WriteString(ovalShape(id.next(), sW-1500000, sH-1500000, 2200000, 2200000, secondary, 8000))
	// Decorative circle top-left
	b.WriteString(ovalShape(id.next(), -600000, -600000, 2500000, 2500000, secondary, 4000))

	titleY := sH/2 - 700000
	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		b.WriteString(textBox(id.next(), mL, titleY, cW, 850000, titleText, onBg, 4400, true, false, "l", headingFont))
	}
	// Accent bar
	b.WriteString(gradRect(id.next(), mL, titleY+850000+80000, 700000, 70000, primary, secondary))

	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, titleY+850000+220000, 5800000, 500000, subtitleText, muted, 2200, false, false, "l", bodyFont))
	}
	return b.String()
}

func buildSectionDivider(id *idGen, node *barqv1.SlideNode, primary, secondary, headingFont, bodyFont string) string {
	var b strings.Builder
	// Full gradient background
	b.WriteString(gradRectDiag(id.next(), 0, 0, sW, sH, primary, secondary))
	// Semi-transparent overlay circles
	b.WriteString(ovalShape(id.next(), -500000, -500000, 2800000, 2800000, "FFFFFF", 5000))
	b.WriteString(ovalShape(id.next(), sW-1500000, sH-1500000, 2800000, 2800000, "FFFFFF", 4000))
	// Large decorative number (white, very transparent)
	slideNum := fmt.Sprintf("%d", node.GetIndex())
	b.WriteString(textBox(id.next(), sW-2500000, sH/2-1200000, 2000000, 2400000, slideNum, "FFFFFF", 22000, true, false, "r", headingFont))

	// "SECTION" label
	b.WriteString(textBox(id.next(), mL, sH/2-800000, cW, 280000, "SECTION", "FFFFFF", 1100, false, false, "l", bodyFont))

	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		b.WriteString(textBox(id.next(), mL, sH/2-500000, cW*4/5, 800000, titleText, "FFFFFF", 4000, true, false, "l", headingFont))
	}
	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, sH/2+350000, cW*3/4, 420000, subtitleText, "FFFFFF", 2000, false, false, "l", bodyFont))
	}
	return b.String()
}

func buildBulletList(id *idGen, node *barqv1.SlideNode, primary, secondary, bg, onBg, muted, headingFont, bodyFont string) string {
	var b strings.Builder
	b.WriteString(rectShape(id.next(), 0, 0, sW, sH, bg))
	// Top gradient bar
	b.WriteString(gradRect(id.next(), 0, 0, sW, 85000, primary, secondary))
	// Decorative circle
	b.WriteString(ovalShape(id.next(), sW-1200000, -600000, 2200000, 2200000, primary, 3500))

	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		b.WriteString(textBox(id.next(), mL, mT, cW, hH, titleText, onBg, 3200, true, false, "l", headingFont))
		// Accent underline
		b.WriteString(rectShape(id.next(), mL, mT+hH+80000, 500000, 60000, primary))
	}
	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, bT-200000, cW, 380000, subtitleText, muted, 1800, false, false, "l", bodyFont))
	}

	bullets := blockLines(node, "bullet")
	if len(bullets) > 0 {
		bodyTopOffset := bT
		if subtitleText != "" {
			bodyTopOffset += 200000
		}
		b.WriteString(bulletBox(id.next(), mL, bodyTopOffset, cW, bH, bullets, onBg, primary, 1900, bodyFont))
	}
	return b.String()
}

func buildTwoColumn(id *idGen, node *barqv1.SlideNode, primary, secondary, bg, onBg, muted, surface, headingFont, bodyFont string) string {
	var b strings.Builder
	b.WriteString(rectShape(id.next(), 0, 0, sW, sH, bg))
	b.WriteString(gradRect(id.next(), 0, 0, sW, 85000, primary, secondary))
	b.WriteString(ovalShape(id.next(), sW-900000, -500000, 1800000, 1800000, secondary, 4000))

	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		b.WriteString(textBox(id.next(), mL, mT, cW, hH, titleText, onBg, 3200, true, false, "l", headingFont))
	}
	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, mT+hH+gp/2, cW, 380000, subtitleText, muted, 1700, false, false, "l", bodyFont))
	}

	// Collect bullet blocks: first half → left, second half → right
	allBullets := blockLines(node, "bullet")
	half := (len(allBullets) + 1) / 2
	leftBullets := allBullets
	rightBullets := []string{}
	if half < len(allBullets) {
		leftBullets = allBullets[:half]
		rightBullets = allBullets[half:]
	}

	colBodyTop := bT
	if subtitleText != "" {
		colBodyTop = bT + 200000
	}
	colH := sH - colBodyTop - mB

	// Left card
	b.WriteString(roundRectShape(id.next(), mL, colBodyTop, c2, colH, surface))
	// Left top accent
	b.WriteString(rectShape(id.next(), mL, colBodyTop, c2, 55000, primary))
	if len(leftBullets) > 0 {
		b.WriteString(bulletBox(id.next(), mL+160000, colBodyTop+120000, c2-320000, colH-160000, leftBullets, onBg, primary, 1800, bodyFont))
	}

	// Right card
	rightX := mL + c2 + gp
	b.WriteString(roundRectShape(id.next(), rightX, colBodyTop, c2, colH, surface))
	// Right top accent
	b.WriteString(rectShape(id.next(), rightX, colBodyTop, c2, 55000, secondary))
	if len(rightBullets) > 0 {
		b.WriteString(bulletBox(id.next(), rightX+160000, colBodyTop+120000, c2-320000, colH-160000, rightBullets, onBg, secondary, 1800, bodyFont))
	}
	return b.String()
}

func buildQuote(id *idGen, node *barqv1.SlideNode, primary, secondary, bg, onBg, muted, headingFont, bodyFont string) string {
	var b strings.Builder
	b.WriteString(rectShape(id.next(), 0, 0, sW, sH, bg))
	// Left accent bar
	b.WriteString(gradRectDiag(id.next(), 0, 0, 90000, sH, primary, secondary))
	// Decorative circle
	b.WriteString(ovalShape(id.next(), sW-1500000, -800000, 2500000, 2500000, primary, 3500))

	// Large quotation mark
	b.WriteString(textBox(id.next(), mL+200000, sH/2-1000000, 1000000, 900000, "\u201C", primary, 22000, true, false, "l", headingFont))

	quoteText := blockText(node, "quote")
	if quoteText == "" {
		// fall back to first bullet
		lines := blockLines(node, "bullet")
		if len(lines) > 0 {
			quoteText = lines[0]
		} else {
			quoteText = blockText(node, "heading", "title")
		}
	}
	if quoteText != "" {
		b.WriteString(textBox(id.next(), mL+200000, sH/2-400000, cW-200000, 900000, quoteText, onBg, 2800, false, true, "l", headingFont))
	}

	attrText := blockText(node, "attribution")
	if attrText == "" {
		attrText = blockText(node, "subheading", "subtitle")
	}
	if attrText != "" {
		// Accent dash + attribution
		b.WriteString(rectShape(id.next(), mL+200000, sH/2+580000, 300000, 50000, primary))
		b.WriteString(textBox(id.next(), mL+600000, sH/2+520000, cW-600000, 350000, strings.ToUpper(attrText), primary, 1400, true, false, "l", bodyFont))
	}
	return b.String()
}

func buildKPICards(id *idGen, node *barqv1.SlideNode, primary, secondary, bg, onBg, muted, surface, headingFont, bodyFont string) string {
	var b strings.Builder
	b.WriteString(rectShape(id.next(), 0, 0, sW, sH, bg))
	b.WriteString(gradRect(id.next(), 0, 0, sW, 85000, primary, secondary))
	b.WriteString(ovalShape(id.next(), sW-800000, sH-800000, 1600000, 1600000, secondary, 5000))

	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		b.WriteString(textBox(id.next(), mL, mT, cW, hH, titleText, onBg, 3000, true, false, "l", headingFont))
	}
	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, mT+hH+gp/2, cW, 360000, subtitleText, muted, 1700, false, false, "l", bodyFont))
	}

	// Collect KPI blocks (kpi_value role has "label|value|delta" text)
	var kpis []string
	for _, blk := range node.GetBlocks() {
		if blk.GetRole() == "kpi_value" {
			kpis = append(kpis, blk.GetText())
		}
	}
	// Fallback: use bullet lines as KPI items
	if len(kpis) == 0 {
		kpis = blockLines(node, "bullet")
	}
	if len(kpis) == 0 {
		return b.String()
	}

	if len(kpis) > 4 {
		kpis = kpis[:4]
	}
	n := int64(len(kpis))
	cardW := (cW - gp*(n-1)) / n
	cardH := bH
	if subtitleText != "" {
		cardH -= 250000
	}

	cardY := bT
	if subtitleText != "" {
		cardY += 200000
	}

	accents := []string{primary, secondary, primary, secondary}
	for i, kpiText := range kpis {
		parts := strings.SplitN(kpiText, "|", 3)
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
		// If no pipe separator, treat whole text as value
		if value == "" {
			value = label
			label = ""
		}

		cardX := mL + int64(i)*(cardW+gp)
		accent := accents[i%len(accents)]
		b.WriteString(kpiCard(id.next(), cardX, cardY, cardW, cardH, label, value, delta, surface, accent, onBg, muted, headingFont, bodyFont))
		id.n += 2 // kpiCard uses ids: N, N+1, N+2 (rect + roundRect + text sp)
	}
	return b.String()
}

func buildClosing(id *idGen, node *barqv1.SlideNode, primary, secondary, headingFont, bodyFont string) string {
	var b strings.Builder
	// Full diagonal gradient
	b.WriteString(gradRectDiag(id.next(), 0, 0, sW, sH, primary, secondary))
	// Decorative circles
	b.WriteString(ovalShape(id.next(), -800000, -800000, 3000000, 3000000, "FFFFFF", 5000))
	b.WriteString(ovalShape(id.next(), sW-1800000, sH-1800000, 3000000, 3000000, "FFFFFF", 4000))

	titleText := blockText(node, "heading", "title")
	if titleText != "" {
		titleY := sH/2 - 650000
		b.WriteString(textBox(id.next(), mL, titleY, cW, 800000, titleText, "FFFFFF", 4200, true, false, "ctr", headingFont))
		// White accent bar centered
		barW := int64(800000)
		b.WriteString(rectShapeAlpha(id.next(), sW/2-barW/2, titleY+830000, barW, 70000, "FFFFFF", 60000))
	}
	subtitleText := blockText(node, "subheading", "subtitle")
	if subtitleText != "" {
		b.WriteString(textBox(id.next(), mL, sH/2+250000, cW, 420000, subtitleText, "FFFFFF", 2000, false, false, "ctr", bodyFont))
	}
	return b.String()
}
