// Package compose generates HTML/CSS representations of SlideNodes
// using Go's html/template engine and CSS Grid for layout.
package compose

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/layout"
)

// slideCanvasWidth and slideCanvasHeight define the HTML preview canvas size in px.
const (
	slideCanvasWidth  = 1200
	slideCanvasHeight = 675
	emuToPixelFactor  = float64(slideCanvasWidth) / float64(layout.SlideWidthEMU)
)

// HTMLComposer converts a SlideNode into a self-contained HTML document.
type HTMLComposer struct{}

// NewHTMLComposer creates an HTMLComposer.
func NewHTMLComposer() *HTMLComposer { return &HTMLComposer{} }

// Compose renders a SlideNode to a self-contained HTML string
// suitable for browser rendering or chromedp screenshot.
func (c *HTMLComposer) Compose(node *barqv1.SlideNode) (string, error) {
	geo := layout.Find(node.GetLayout().GetLayoutId())
	tokens := node.GetTokens()

	vars := buildCSSVars(tokens)
	regions := buildRegions(geo, tokens)
	contentBlocks := buildContentBlocks(node, geo)

	data := slideTemplateData{
		Width:         slideCanvasWidth,
		Height:        slideCanvasHeight,
		CSSVars:       vars,
		FontCSS:       buildFontCSS(tokens),
		BackgroundHex: tokens.GetBackgroundHex(),
		Regions:       regions,
		ContentBlocks: contentBlocks,
		SlideID:       node.GetId(),
		SlideRole:     node.GetRole().String(),
	}

	var buf bytes.Buffer
	if err := slideTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("compose: rendering slide template: %w", err)
	}

	return buf.String(), nil
}

// ── Template data types ────────────────────────────────────────────────────────

type slideTemplateData struct {
	Width         int
	Height        int
	CSSVars       string
	FontCSS       string
	BackgroundHex string
	Regions       []regionData
	ContentBlocks []contentBlockData
	SlideID       string
	SlideRole     string
}

type regionData struct {
	ID     string
	X      int
	Y      int
	W      int
	H      int
	Role   string // "title", "subtitle", "body-0", "body-1", "media"
}

type contentBlockData struct {
	RegionID string
	HTML     template.HTML
	Role     string
}

// ── Template ───────────────────────────────────────────────────────────────────

var slideTemplate = template.Must(template.New("slide").Funcs(template.FuncMap{
	"raw": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec // controlled content
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<style>
{{ .FontCSS }}
:root { {{ .CSSVars }} }

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  width: {{ .Width }}px;
  height: {{ .Height }}px;
  overflow: hidden;
  font-family: var(--font-body), sans-serif;
  background-color: var(--color-background);
  color: var(--color-on-background);
  position: relative;
}

.slide-canvas {
  width: {{ .Width }}px;
  height: {{ .Height }}px;
  position: relative;
  overflow: hidden;
  background-color: var(--color-background);
}

.slide-region {
  position: absolute;
  display: flex;
  flex-direction: column;
  justify-content: center;
  overflow: hidden;
}

.region-title h1, .region-title h2 {
  font-family: var(--font-heading), sans-serif;
  font-size: var(--size-heading);
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-on-background);
}

.region-subtitle p, .region-subtitle span {
  font-family: var(--font-heading), sans-serif;
  font-size: var(--size-subheading);
  font-weight: 400;
  line-height: 1.4;
  color: var(--color-muted);
}

.region-body ul {
  list-style: none;
  padding: 0;
}

.region-body ul li {
  font-family: var(--font-body), sans-serif;
  font-size: var(--size-body);
  line-height: var(--line-height-body);
  color: var(--color-on-background);
  padding: 6px 0 6px 24px;
  position: relative;
}

.region-body ul li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--color-primary);
}

.region-media {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: var(--border-radius);
}

.region-media svg {
  width: 100%;
  height: 100%;
}

.kpi-card {
  background: var(--color-surface);
  border-radius: var(--border-radius);
  padding: 20px;
  text-align: center;
  border-left: 4px solid var(--color-primary);
}

.kpi-value {
  font-family: var(--font-heading), sans-serif;
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--color-primary);
}

.kpi-label {
  font-size: var(--size-caption);
  color: var(--color-muted);
  margin-top: 4px;
}

.kpi-delta {
  font-size: var(--size-caption);
  color: var(--color-secondary);
  font-weight: 600;
}
</style>
</head>
<body>
<div class="slide-canvas" id="slide-{{ .SlideID }}" data-role="{{ .SlideRole }}">
{{ range .Regions }}
  <div class="slide-region region-{{ .Role }}"
       style="left:{{ .X }}px;top:{{ .Y }}px;width:{{ .W }}px;height:{{ .H }}px;"
       id="region-{{ .ID }}">
    {{ range $.ContentBlocks }}{{ if eq .RegionID $.SlideID }}_{{ end }}{{ end }}
  </div>
{{ end }}
{{ range .ContentBlocks }}
  <div class="slide-region region-{{ .Role }}"
       style="position:absolute;" id="cb-{{ .RegionID }}">
    {{ raw .HTML }}
  </div>
{{ end }}
</div>
</body>
</html>
`))

// ── Helpers ────────────────────────────────────────────────────────────────────

func buildCSSVars(tokens *barqv1.DesignTokens) string {
	headingSize := "2.4rem"
	subheadingSize := "1.6rem"
	bodySize := "1rem"
	captionSize := "0.75rem"
	lineHeight := "1.6"
	headingFont := "Inter"
	bodyFont := "Inter"
	borderRadius := "6px"

	if h := tokens.GetHeading(); h != nil {
		headingSize = fmt.Sprintf("%.2frem", h.GetSizePt()/16.0)
		headingFont = h.GetFontFamily()
	}
	if s := tokens.GetSubheading(); s != nil {
		subheadingSize = fmt.Sprintf("%.2frem", s.GetSizePt()/16.0)
	}
	if b := tokens.GetBody(); b != nil {
		bodySize = fmt.Sprintf("%.2frem", b.GetSizePt()/16.0)
		lineHeight = fmt.Sprintf("%.2f", b.GetLineHeight())
		bodyFont = b.GetFontFamily()
	}
	if c := tokens.GetCaption(); c != nil {
		captionSize = fmt.Sprintf("%.2frem", c.GetSizePt()/16.0)
	}
	if tokens.GetBorderRadiusEmu() > 0 {
		borderRadius = fmt.Sprintf("%.0fpx", float64(tokens.GetBorderRadiusEmu())/12700.0)
	}

	return fmt.Sprintf(`
  --color-primary: %s;
  --color-secondary: %s;
  --color-background: %s;
  --color-surface: %s;
  --color-on-primary: %s;
  --color-on-background: %s;
  --color-muted: %s;
  --font-heading: '%s';
  --font-body: '%s';
  --size-heading: %s;
  --size-subheading: %s;
  --size-body: %s;
  --size-caption: %s;
  --line-height-body: %s;
  --border-radius: %s;`,
		tokens.GetPrimaryHex(), tokens.GetSecondaryHex(),
		tokens.GetBackgroundHex(), tokens.GetSurfaceHex(),
		tokens.GetOnPrimaryHex(), tokens.GetOnBackgroundHex(),
		tokens.GetMutedHex(),
		headingFont, bodyFont,
		headingSize, subheadingSize, bodySize, captionSize, lineHeight,
		borderRadius,
	)
}

func buildFontCSS(tokens *barqv1.DesignTokens) string {
	fonts := map[string]bool{}
	if h := tokens.GetHeading(); h != nil {
		fonts[h.GetFontFamily()] = true
	}
	if b := tokens.GetBody(); b != nil {
		fonts[b.GetFontFamily()] = true
	}

	var sb strings.Builder
	for family := range fonts {
		slug := strings.ReplaceAll(family, " ", "+")
		fmt.Fprintf(&sb, `@import url('https://fonts.googleapis.com/css2?family=%s:wght@400;600;700&display=swap');`+"\n", slug)
	}
	return sb.String()
}

func emuToPx(emu int64) int {
	return int(float64(emu) * emuToPixelFactor)
}

func buildRegions(geo layout.Geometry, _ *barqv1.DesignTokens) []regionData {
	var regions []regionData

	if geo.Title.Width > 0 {
		regions = append(regions, regionData{
			ID: "title", Role: "title",
			X: emuToPx(geo.Title.X), Y: emuToPx(geo.Title.Y),
			W: emuToPx(geo.Title.Width), H: emuToPx(geo.Title.Height),
		})
	}
	if geo.Subtitle.Width > 0 {
		regions = append(regions, regionData{
			ID: "subtitle", Role: "subtitle",
			X: emuToPx(geo.Subtitle.X), Y: emuToPx(geo.Subtitle.Y),
			W: emuToPx(geo.Subtitle.Width), H: emuToPx(geo.Subtitle.Height),
		})
	}
	for i, b := range geo.Body {
		regions = append(regions, regionData{
			ID: fmt.Sprintf("body-%d", i), Role: fmt.Sprintf("body-%d", i),
			X: emuToPx(b.X), Y: emuToPx(b.Y),
			W: emuToPx(b.Width), H: emuToPx(b.Height),
		})
	}
	if geo.HasMediaRegion && geo.Media.Width > 0 {
		regions = append(regions, regionData{
			ID: "media", Role: "media",
			X: emuToPx(geo.Media.X), Y: emuToPx(geo.Media.Y),
			W: emuToPx(geo.Media.Width), H: emuToPx(geo.Media.Height),
		})
	}

	return regions
}

func buildContentBlocks(node *barqv1.SlideNode, geo layout.Geometry) []contentBlockData {
	var blocks []contentBlockData

	// Heading → title region.
	for _, b := range node.GetBlocks() {
		if b.GetRole() == "heading" {
			blocks = append(blocks, contentBlockData{
				RegionID: fmt.Sprintf("title_%s", node.GetId()),
				Role:     "title",
				HTML: template.HTML(fmt.Sprintf(
					`<div class="region-title" style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;display:flex;align-items:center;"><h1>%s</h1></div>`,
					emuToPx(geo.Title.X), emuToPx(geo.Title.Y),
					emuToPx(geo.Title.Width), emuToPx(geo.Title.Height),
					template.HTMLEscapeString(b.GetText()),
				)),
			})
		}
	}

	// Body bullets.
	var bullets []string
	for _, b := range node.GetBlocks() {
		if b.GetRole() == "body" {
			bullets = append(bullets, b.GetText())
		}
	}
	if len(bullets) > 0 && len(geo.Body) > 0 {
		r := geo.Body[0]
		var items strings.Builder
		for _, bullet := range bullets {
			fmt.Fprintf(&items, "<li>%s</li>", template.HTMLEscapeString(bullet))
		}
		blocks = append(blocks, contentBlockData{
			RegionID: fmt.Sprintf("body_%s", node.GetId()),
			Role:     "body",
			HTML: template.HTML(fmt.Sprintf(
				`<div class="region-body" style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;"><ul>%s</ul></div>`,
				emuToPx(r.X), emuToPx(r.Y), emuToPx(r.Width), emuToPx(r.Height),
				items.String(),
			)),
		})
	}

	// Asset slots (SVG/image).
	for _, asset := range node.GetAssets() {
		if len(asset.GetRenderedBytes()) > 0 && geo.HasMediaRegion {
			r := geo.Media
			var html string
			if strings.HasPrefix(asset.GetRenderedMime(), "image/svg") {
				html = fmt.Sprintf(
					`<div class="region-media" style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;">%s</div>`,
					emuToPx(r.X), emuToPx(r.Y), emuToPx(r.Width), emuToPx(r.Height),
					string(asset.GetRenderedBytes()),
				)
			} else {
				b64 := base64.StdEncoding.EncodeToString(asset.GetRenderedBytes())
				html = fmt.Sprintf(
					`<img class="region-media" style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;" src="data:%s;base64,%s" alt=""/>`,
					emuToPx(r.X), emuToPx(r.Y), emuToPx(r.Width), emuToPx(r.Height),
					asset.GetRenderedMime(), b64,
				)
			}
			blocks = append(blocks, contentBlockData{
				RegionID: fmt.Sprintf("media_%s_%s", node.GetId(), asset.GetId()),
				Role:     "media",
				HTML:     template.HTML(html), //nolint:gosec // controlled content
			})
		}
	}

	return blocks
}
