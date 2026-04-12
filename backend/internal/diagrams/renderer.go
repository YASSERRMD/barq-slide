package diagrams

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
)

const (
	// defaultMermaidCLI is the expected name of the mermaid-cli binary.
	defaultMermaidCLI = "mmdc"
	// renderTimeoutSec is the subprocess timeout in seconds.
	renderTimeoutSec = 30
)

// Renderer converts Mermaid DSL to SVG bytes using the mermaid-cli subprocess.
type Renderer struct {
	// MMDCPath is the absolute path to the mmdc binary.
	// Defaults to "mmdc" (resolved via PATH).
	MMDCPath string
}

// NewRenderer creates a Renderer using the mmdc binary on PATH.
func NewRenderer() *Renderer {
	return &Renderer{MMDCPath: defaultMermaidCLI}
}

// Render converts Mermaid DSL to SVG bytes, applying the design token CSS theme.
func (r *Renderer) Render(ctx context.Context, dsl string, tokens *barqv1.DesignTokens) ([]byte, error) {
	// Write DSL to a temp file.
	tmpDir, err := os.MkdirTemp("", "barq-diagram-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "diagram.mmd")
	outputFile := filepath.Join(tmpDir, "diagram.svg")
	configFile := filepath.Join(tmpDir, "config.json")
	cssFile := filepath.Join(tmpDir, "theme.css")

	if err := os.WriteFile(inputFile, []byte(dsl), 0600); err != nil {
		return nil, fmt.Errorf("writing diagram file: %w", err)
	}

	// Generate theme CSS from design tokens.
	themeCSSBytes, err := generateThemeCSS(tokens)
	if err != nil {
		return nil, fmt.Errorf("generating theme CSS: %w", err)
	}
	if err := os.WriteFile(cssFile, themeCSSBytes, 0600); err != nil {
		return nil, fmt.Errorf("writing CSS file: %w", err)
	}

	// Generate mermaid config JSON.
	cfg := buildMermaidConfig(tokens)
	if err := os.WriteFile(configFile, []byte(cfg), 0600); err != nil {
		return nil, fmt.Errorf("writing config file: %w", err)
	}

	// Execute mmdc.
	cmd := exec.CommandContext(ctx, r.MMDCPath,
		"--input", inputFile,
		"--output", outputFile,
		"--configFile", configFile,
		"--cssFile", cssFile,
		"--backgroundColor", tokens.GetBackgroundHex(),
		"--width", "1200",
		"--height", "675",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mmdc failed: %w\nstderr: %s", err, stderr.String())
	}

	svgBytes, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("reading rendered SVG: %w", err)
	}

	return svgBytes, nil
}

// IsMermaidCLIAvailable checks whether the mmdc binary is on PATH.
func IsMermaidCLIAvailable() bool {
	path, err := exec.LookPath(defaultMermaidCLI)
	return err == nil && path != ""
}

// ── CSS theme generation ───────────────────────────────────────────────────────

var themeCSSTemplate = template.Must(template.New("mermaid-theme").Parse(`
/* barq-slides Mermaid theme — generated from DesignTokens */
:root {
  --primary: {{ .Primary }};
  --secondary: {{ .Secondary }};
  --background: {{ .Background }};
  --surface: {{ .Surface }};
  --text: {{ .OnBackground }};
  --muted: {{ .Muted }};
  --font-family: {{ .FontFamily }};
}

.mermaid {
  font-family: var(--font-family), sans-serif;
}

/* Node boxes */
.node rect, .node circle, .node ellipse, .node polygon {
  fill: {{ .Surface }} !important;
  stroke: {{ .Primary }} !important;
  stroke-width: 2px !important;
}

/* Node labels */
.node .label {
  color: {{ .OnBackground }} !important;
  font-family: var(--font-family), sans-serif !important;
}

/* Edges / arrows */
.edgePath .path {
  stroke: {{ .Primary }} !important;
  stroke-width: 2px !important;
}

.arrowheadPath {
  fill: {{ .Primary }} !important;
}

/* Edge labels */
.edgeLabel {
  background-color: {{ .Background }} !important;
  color: {{ .Muted }} !important;
  font-size: 12px !important;
}

/* Flowchart subgraphs */
.cluster rect {
  fill: {{ .Surface }} !important;
  stroke: {{ .Secondary }} !important;
}

/* Sequence diagram */
.actor {
  fill: {{ .Primary }} !important;
  stroke: {{ .Primary }} !important;
}
.actor-line {
  stroke: {{ .Muted }} !important;
}
.messageLine0, .messageLine1 {
  stroke: {{ .Secondary }} !important;
}
.messageText {
  fill: {{ .OnBackground }} !important;
}

/* Gantt chart */
.task {
  fill: {{ .Secondary }} !important;
  stroke: {{ .Primary }} !important;
}
.taskText {
  fill: {{ .OnBackground }} !important;
}
`))

type themeCSSVars struct {
	Primary      string
	Secondary    string
	Background   string
	Surface      string
	OnBackground string
	Muted        string
	FontFamily   string
}

func generateThemeCSS(tokens *barqv1.DesignTokens) ([]byte, error) {
	fontFamily := "Inter"
	if tokens.GetBody() != nil && tokens.GetBody().GetFontFamily() != "" {
		fontFamily = tokens.GetBody().GetFontFamily()
	}

	vars := themeCSSVars{
		Primary:      tokens.GetPrimaryHex(),
		Secondary:    tokens.GetSecondaryHex(),
		Background:   tokens.GetBackgroundHex(),
		Surface:      tokens.GetSurfaceHex(),
		OnBackground: tokens.GetOnBackgroundHex(),
		Muted:        tokens.GetMutedHex(),
		FontFamily:   fontFamily,
	}

	var buf bytes.Buffer
	if err := themeCSSTemplate.Execute(&buf, vars); err != nil {
		return nil, fmt.Errorf("rendering theme CSS: %w", err)
	}
	return buf.Bytes(), nil
}

// buildMermaidConfig produces the mermaid-cli config.json content.
func buildMermaidConfig(tokens *barqv1.DesignTokens) string {
	bg := tokens.GetBackgroundHex()
	primary := tokens.GetPrimaryHex()

	// Escape colours for JSON.
	bg = strings.ReplaceAll(bg, `"`, `\"`)
	primary = strings.ReplaceAll(primary, `"`, `\"`)

	return fmt.Sprintf(`{
  "theme": "base",
  "themeVariables": {
    "primaryColor": "%s",
    "primaryTextColor": "%s",
    "primaryBorderColor": "%s",
    "lineColor": "%s",
    "background": "%s",
    "mainBkg": "%s",
    "nodeBorder": "%s",
    "clusterBkg": "%s",
    "titleColor": "%s",
    "edgeLabelBackground": "%s",
    "fontFamily": "%s"
  }
}`,
		primary,
		tokens.GetOnBackgroundHex(),
		primary,
		primary,
		bg,
		tokens.GetSurfaceHex(),
		primary,
		tokens.GetSurfaceHex(),
		tokens.GetOnBackgroundHex(),
		bg,
		func() string {
			if tokens.GetBody() != nil {
				return tokens.GetBody().GetFontFamily()
			}
			return "Inter"
		}(),
	)
}
