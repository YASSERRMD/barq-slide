// Package diagrams generates Mermaid diagram DSL from natural language queries
// and renders them to SVG via the mermaid-cli subprocess.
package diagrams

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/llm"
)

// DiagramType maps to Mermaid diagram types.
type DiagramType string

const (
	DiagramFlowchart DiagramType = "flowchart"
	DiagramSequence  DiagramType = "sequenceDiagram"
	DiagramMindmap   DiagramType = "mindmap"
	DiagramGantt     DiagramType = "gantt"
	DiagramPie       DiagramType = "pie"
	DiagramER        DiagramType = "erDiagram"
	DiagramClass     DiagramType = "classDiagram"
)

// assetTypeToDiagramType maps proto AssetType to Mermaid diagram type.
var assetTypeToDiagramType = map[barqv1.AssetType]DiagramType{
	barqv1.AssetType_ASSET_TYPE_DIAGRAM_FLOWCHART: DiagramFlowchart,
	barqv1.AssetType_ASSET_TYPE_DIAGRAM_SEQUENCE:  DiagramSequence,
	barqv1.AssetType_ASSET_TYPE_DIAGRAM_MINDMAP:   DiagramMindmap,
	barqv1.AssetType_ASSET_TYPE_DIAGRAM_GANTT:     DiagramGantt,
}

// mermaidLLMResponse is the JSON the LLM fills for diagram generation.
type mermaidLLMResponse struct {
	DiagramType string `json:"diagram_type"`
	MermaidDSL  string `json:"mermaid_dsl"`
}

var mermaidSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "diagram_type": {
      "type": "string",
      "enum": ["flowchart","sequenceDiagram","mindmap","gantt","pie","erDiagram"]
    },
    "mermaid_dsl": {
      "type": "string",
      "description": "Valid Mermaid diagram DSL. Must start with the diagram type keyword."
    }
  },
  "required": ["diagram_type","mermaid_dsl"]
}`)

// Generator generates Mermaid DSL from natural language queries using an LLM.
type Generator struct {
	provider llm.Provider
}

// NewGenerator creates a Generator backed by the given LLM provider.
func NewGenerator(provider llm.Provider) *Generator {
	return &Generator{provider: provider}
}

// Generate produces Mermaid DSL for the given asset slot query.
func (g *Generator) Generate(ctx context.Context, slot *barqv1.AssetSlot, tokens *barqv1.DesignTokens) (string, error) {
	diagramType := assetTypeToDiagramType[slot.GetType()]
	if diagramType == "" {
		diagramType = DiagramFlowchart
	}

	// If the slot already has mermaid source, use it directly.
	if src := strings.TrimSpace(slot.GetMermaidSource()); src != "" {
		return src, nil
	}

	userMsg := buildMermaidPrompt(slot.GetQuery(), diagramType, tokens)

	resp, err := g.provider.GenerateStructured(ctx, llm.GenerateStructuredRequest{
		System: mermaidSystemPrompt,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: userMsg},
		},
		ToolName:        "output_formatter",
		ToolDescription: "Generate valid Mermaid diagram DSL for the given query.",
		Schema:          mermaidSchema,
		MaxTokens:       2048,
	})
	if err != nil {
		return "", fmt.Errorf("diagram generator: LLM call failed: %w", err)
	}

	var llmResp mermaidLLMResponse
	if err := resp.UnmarshalInto(&llmResp); err != nil {
		return "", fmt.Errorf("diagram generator: decoding response: %w", err)
	}

	dsl := strings.TrimSpace(llmResp.MermaidDSL)
	if dsl == "" {
		return "", fmt.Errorf("diagram generator: empty DSL returned")
	}

	// Validate that the DSL starts with a recognised Mermaid keyword.
	if !isValidMermaidStart(dsl) {
		return "", fmt.Errorf("diagram generator: invalid Mermaid DSL (must start with diagram type): %q", dsl[:min(50, len(dsl))])
	}

	return dsl, nil
}

const mermaidSystemPrompt = `You are a Mermaid diagram expert.

Generate valid, clean Mermaid diagram DSL that renders correctly.

Rules:
- Start the DSL with the diagram type keyword (e.g., "flowchart LR", "sequenceDiagram", "gantt")
- Use clear, concise node labels (max 40 chars each)
- Limit to 8 nodes/items for readability on a slide
- For flowcharts, use LR (left-to-right) direction
- Do NOT include markdown fences (no triple backticks)
- Do NOT include emoji
- Use only ASCII characters in node labels

Respond using the output_formatter tool.`

func buildMermaidPrompt(query string, dtype DiagramType, tokens *barqv1.DesignTokens) string {
	return fmt.Sprintf(`Generate a Mermaid %s diagram for this concept:

"%s"

The diagram should be clear, professional, and suitable for a %s-mood presentation.
Keep it concise — max 8 nodes.`,
		dtype,
		query,
		tokens.GetMood(),
	)
}

// isValidMermaidStart checks that the DSL begins with a known Mermaid diagram type.
func isValidMermaidStart(dsl string) bool {
	keywords := []string{
		"flowchart", "graph", "sequenceDiagram", "classDiagram",
		"stateDiagram", "erDiagram", "gantt", "pie", "mindmap",
		"timeline", "gitGraph", "quadrantChart", "requirementDiagram",
	}
	lower := strings.ToLower(dsl)
	for _, kw := range keywords {
		if strings.HasPrefix(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
