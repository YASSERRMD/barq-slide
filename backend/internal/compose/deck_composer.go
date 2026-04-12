package compose

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

// DeckComposer renders a full deck of SlideNodes to a multi-slide HTML document.
type DeckComposer struct {
	slideComposer *HTMLComposer
}

// NewDeckComposer creates a DeckComposer.
func NewDeckComposer() *DeckComposer {
	return &DeckComposer{slideComposer: NewHTMLComposer()}
}

// ComposeSlide renders a single slide to HTML.
func (d *DeckComposer) ComposeSlide(node *barqv1.SlideNode) (string, error) {
	return d.slideComposer.Compose(node)
}

// ComposeDeck renders all slides into a scrollable multi-slide HTML document.
// Used for full-deck preview in the browser.
func (d *DeckComposer) ComposeDeck(slides []*barqv1.SlideNode, tokens *barqv1.DesignTokens) (string, error) {
	var slideHTMLs []string

	for _, slide := range slides {
		// Each slide uses the deck-level tokens if the slide doesn't have its own.
		if slide.GetTokens() == nil {
			slide.Tokens = tokens
		}

		html, err := d.slideComposer.Compose(slide)
		if err != nil {
			return "", fmt.Errorf("composing slide %d: %w", slide.GetIndex(), err)
		}

		// Extract just the body content from the self-contained slide HTML.
		slideHTML := extractBodyContent(html)
		slideHTMLs = append(slideHTMLs, slideHTML)
	}

	data := deckTemplateData{
		CSSVars:    buildCSSVars(tokens),
		FontCSS:    buildFontCSS(tokens),
		Background: tokens.GetBackgroundHex(),
		SlideHTMLs: slideHTMLs,
		TotalSlides: len(slides),
	}

	var buf bytes.Buffer
	if err := deckTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("composing deck: %w", err)
	}

	return buf.String(), nil
}

type deckTemplateData struct {
	CSSVars     string
	FontCSS     string
	Background  string
	SlideHTMLs  []string
	TotalSlides int
}

var deckTemplate = template.Must(template.New("deck").Funcs(template.FuncMap{
	"raw": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<style>
{{ .FontCSS }}
:root { {{ .CSSVars }} }
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #1a1a2e; font-family: sans-serif; }

.deck-viewer {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  gap: 24px;
}

.slide-wrapper {
  width: {{ $.Width }}px;
  height: {{ $.Height }}px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}

.slide-number {
  color: #888;
  font-size: 12px;
  text-align: center;
  margin-top: 8px;
}
</style>
</head>
<body>
<div class="deck-viewer">
{{ range $i, $html := .SlideHTMLs }}
  <div>
    <div class="slide-wrapper" style="width:1200px;height:675px;">
      {{ raw $html }}
    </div>
    <div class="slide-number">Slide {{ add $i 1 }} of {{ $.TotalSlides }}</div>
  </div>
{{ end }}
</div>
</body>
</html>
`))

// extractBodyContent extracts the content inside the <body> tag from a full HTML document.
func extractBodyContent(html string) string {
	start := strings.Index(html, "<body>")
	end := strings.LastIndex(html, "</body>")
	if start >= 0 && end > start {
		return html[start+6 : end]
	}
	return html
}
