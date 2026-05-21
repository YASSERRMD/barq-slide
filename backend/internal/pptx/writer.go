// Package pptx generates OpenXML .pptx files using only the Go standard
// library (archive/zip + strings/fmt). No commercial license required.
package pptx

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// Assembler builds a .pptx from SlideNodes and DesignTokens.
type Assembler struct {
	tokens *barqv1.DesignTokens
	slides []*barqv1.SlideNode
}

// NewAssembler creates an Assembler with the given design tokens.
func NewAssembler(tokens *barqv1.DesignTokens) (*Assembler, error) {
	if tokens == nil {
		tokens = &barqv1.DesignTokens{}
	}
	return &Assembler{tokens: tokens}, nil
}

// AddSlide appends a slide to the presentation.
func (a *Assembler) AddSlide(node *barqv1.SlideNode, _ *barqv1.DesignTokens) error {
	a.slides = append(a.slides, node)
	return nil
}

// Save writes the complete .pptx to w.
func (a *Assembler) Save(w io.Writer) error {
	zw := zip.NewWriter(w)

	type entry struct{ path, body string }
	n := len(a.slides)

	entries := []entry{
		{"[Content_Types].xml", contentTypesXML(n)},
		{"_rels/.rels", rootRelsXML()},
		{"ppt/presentation.xml", presentationXML(n)},
		{"ppt/_rels/presentation.xml.rels", presentationRelsXML(n)},
		{"ppt/theme/theme1.xml", themeXML(a.tokens)},
		{"ppt/slideMasters/slideMaster1.xml", slideMasterXML()},
		{"ppt/slideMasters/_rels/slideMaster1.xml.rels", slideMasterRelsXML()},
		{"ppt/slideLayouts/slideLayout1.xml", slideLayoutXML()},
		{"ppt/slideLayouts/_rels/slideLayout1.xml.rels", slideLayoutRelsXML()},
	}

	for i, node := range a.slides {
		sn := i + 1
		entries = append(entries,
			entry{fmt.Sprintf("ppt/slides/slide%d.xml", sn), generateSlideXML(node, a.tokens)},
			entry{fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", sn), slideRelsXML()},
		)
	}

	for _, e := range entries {
		f, err := zw.Create(e.path)
		if err != nil {
			return fmt.Errorf("pptx create %s: %w", e.path, err)
		}
		if _, err := io.WriteString(f, e.body); err != nil {
			return fmt.Errorf("pptx write %s: %w", e.path, err)
		}
	}
	return zw.Close()
}

// ── shape ID generator ────────────────────────────────────────────────────────

type idGen struct{ n int }

func (g *idGen) next() int { g.n++; return g.n }

// ── XML helpers ───────────────────────────────────────────────────────────────

// xmlEsc escapes the five XML special characters.
func xmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// hex6 strips a leading '#' and ensures 6 hex digits.
func hex6(h string) string {
	h = strings.TrimPrefix(h, "#")
	if len(h) == 6 {
		return strings.ToUpper(h)
	}
	return "111827"
}

// pptxFont extracts the primary font name from a CSS font-family string and
// returns a safe single typeface name for use in OOXML latin typeface attributes.
// E.g. "Inter, sans-serif" → "Inter", "" → "Calibri".
func pptxFont(family string) string {
	if idx := strings.IndexByte(family, ','); idx >= 0 {
		family = strings.TrimSpace(family[:idx])
	}
	family = strings.Trim(family, `"'`)
	if family == "" {
		return "Calibri"
	}
	return family
}

// ── common shape XML builders ─────────────────────────────────────────────────

// rectShape renders a solid-filled rectangle with no border.
func rectShape(id int, x, y, w, h int64, colorHex string) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="r%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
<a:solidFill><a:srgbClr val="%s"/></a:solidFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(colorHex))
}

// rectShapeAlpha renders a semi-transparent solid rectangle (alpha 0-100000).
func rectShapeAlpha(id int, x, y, w, h int64, colorHex string, alpha int) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="ra%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
<a:solidFill><a:srgbClr val="%s"><a:alpha val="%d"/></a:srgbClr></a:solidFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(colorHex), alpha)
}

// roundRectShape renders a solid-filled rounded rectangle (no border).
func roundRectShape(id int, x, y, w, h int64, colorHex string) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="rr%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="roundRect"><a:avLst><a:gd name="adj" fmla="val 20000"/></a:avLst></a:prstGeom>
<a:solidFill><a:srgbClr val="%s"/></a:solidFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(colorHex))
}

// gradRect renders a horizontal gradient rectangle (left→right).
func gradRect(id int, x, y, w, h int64, c1, c2 string) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="g%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
<a:gradFill><a:gsLst>
<a:gs pos="0"><a:srgbClr val="%s"/></a:gs>
<a:gs pos="100000"><a:srgbClr val="%s"/></a:gs>
</a:gsLst><a:lin ang="0" scaled="0"/></a:gradFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(c1), hex6(c2))
}

// gradRectDiag renders a diagonal gradient (top-left→bottom-right).
func gradRectDiag(id int, x, y, w, h int64, c1, c2 string) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="gd%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
<a:gradFill><a:gsLst>
<a:gs pos="0"><a:srgbClr val="%s"/></a:gs>
<a:gs pos="100000"><a:srgbClr val="%s"/></a:gs>
</a:gsLst><a:lin ang="2700000" scaled="0"/></a:gradFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(c1), hex6(c2))
}

// ovalShape renders a semi-transparent circle.
func ovalShape(id int, x, y, w, h int64, colorHex string, alpha int) string {
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="o%d"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="ellipse"><a:avLst/></a:prstGeom>
<a:solidFill><a:srgbClr val="%s"><a:alpha val="%d"/></a:srgbClr></a:solidFill>
<a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p/></p:txBody></p:sp>`,
		id, id, x, y, w, h, hex6(colorHex), alpha)
}

// textRun builds a single <a:r> run.
func textRun(text, colorHex string, sizePt100 int, bold, italic bool, font string) string {
	bVal := "0"
	if bold {
		bVal = "1"
	}
	iVal := "0"
	if italic {
		iVal = "1"
	}
	if font == "" {
		font = "Calibri"
	}
	return fmt.Sprintf(
		`<a:r><a:rPr lang="en-US" b="%s" i="%s" sz="%d" dirty="0"><a:solidFill><a:srgbClr val="%s"/></a:solidFill><a:latin typeface="%s"/></a:rPr><a:t>%s</a:t></a:r>`,
		bVal, iVal, sizePt100, hex6(colorHex), font, xmlEsc(text))
}

// textBox renders a text box with a single paragraph.
// font is the OOXML typeface name (e.g. "Inter"); pass "" to use "Calibri".
func textBox(id int, x, y, w, h int64, text, colorHex string, sizePt100 int, bold, italic bool, align, font string) string {
	if align == "" {
		align = "l"
	}
	run := textRun(text, colorHex, sizePt100, bold, italic, pptxFont(font))
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="t%d"/><p:cNvSpPr txBox="1"><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr wrap="square" lIns="0" rIns="0" tIns="0" bIns="0"><a:normAutofit/></a:bodyPr>
<a:lstStyle/><a:p><a:pPr algn="%s"/>%s</a:p></p:txBody></p:sp>`,
		id, id, x, y, w, h, align, run)
}

// bulletBox renders a text box with multiple bullet paragraphs.
// Each element of lines is one bullet. If a line contains \n, it is split further.
// font is the OOXML typeface name for bullet text; pass "" to use "Calibri".
func bulletBox(id int, x, y, w, h int64, lines []string, colorHex, dotColor string, sizePt100 int, font string) string {
	f := pptxFont(font)
	var paras strings.Builder
	for _, raw := range lines {
		sublines := strings.Split(raw, "\n")
		for _, line := range sublines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			run := textRun(line, colorHex, sizePt100, false, false, f)
			// bullet character paragraph
			paras.WriteString(fmt.Sprintf(
				`<a:p><a:pPr indent="-180000" marL="180000"><a:buFont typeface="Arial"/><a:buChar char="•"/></a:pPr>`+
					`<a:r><a:rPr lang="en-US" sz="1000" dirty="0"><a:solidFill><a:srgbClr val="%s"/></a:solidFill></a:rPr><a:t>  </a:t></a:r>`+
					`%s</a:p>`,
				hex6(dotColor), run))
		}
	}
	return fmt.Sprintf(`<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="b%d"/><p:cNvSpPr txBox="1"><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr wrap="square" lIns="0" rIns="0" tIns="91440" bIns="91440"><a:normAutofit/></a:bodyPr>
<a:lstStyle/>%s</p:txBody></p:sp>`,
		id, id, x, y, w, h, paras.String())
}

// kpiCard renders a rounded card with label/value/delta text inside.
// headingFont is used for the large value; bodyFont for the label and delta.
func kpiCard(id int, x, y, w, h int64, label, value, delta, cardBg, accent, textColor, mutedColor, headingFont, bodyFont string) string {
	// For positive delta, use green; negative use red; else muted
	deltaColor := mutedColor
	if strings.HasPrefix(delta, "+") || strings.HasPrefix(delta, "↑") {
		deltaColor = "10B981"
	} else if strings.HasPrefix(delta, "-") || strings.HasPrefix(delta, "↓") {
		deltaColor = "EF4444"
	}

	hFont := pptxFont(headingFont)
	bFont := pptxFont(bodyFont)

	var paras strings.Builder
	// value paragraph (large)
	paras.WriteString(fmt.Sprintf(`<a:p><a:pPr algn="l"/>%s</a:p>`,
		textRun(value, accent, 4200, true, false, hFont)))
	// label paragraph
	paras.WriteString(fmt.Sprintf(`<a:p><a:pPr algn="l"/>%s</a:p>`,
		textRun(strings.ToUpper(label), mutedColor, 1100, false, false, bFont)))
	// delta paragraph
	if delta != "" {
		paras.WriteString(fmt.Sprintf(`<a:p><a:pPr algn="l"/>%s</a:p>`,
			textRun(delta, deltaColor, 1400, true, false, bFont)))
	}

	return fmt.Sprintf(`%s<p:sp>
<p:nvSpPr><p:cNvPr id="%d" name="kpi%d"/><p:cNvSpPr txBox="1"><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr wrap="square" lIns="180000" rIns="180000" tIns="180000" bIns="180000" anchor="ctr"><a:normAutofit/></a:bodyPr>
<a:lstStyle/>%s</p:txBody></p:sp>`,
		// top accent bar + card background: uses ids id, id+1; text sp uses id+2
		rectShape(id, x, y, w, 60000, accent)+
			roundRectShape(id+1, x, y+60000, w, h-60000, cardBg),
		id+2, id+2, x, y+60000, w, h-60000, paras.String())
}

// ── block helpers ─────────────────────────────────────────────────────────────

func blockText(node *barqv1.SlideNode, roles ...string) string {
	for _, b := range node.GetBlocks() {
		for _, r := range roles {
			if b.GetRole() == r {
				return b.GetText()
			}
		}
	}
	return ""
}

func blockLines(node *barqv1.SlideNode, role string) []string {
	var out []string
	for _, b := range node.GetBlocks() {
		if b.GetRole() == role {
			for _, line := range strings.Split(b.GetText(), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					out = append(out, line)
				}
			}
		}
	}
	return out
}
