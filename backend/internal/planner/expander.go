package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/intent"
	"github.com/YASSERRMD/barq-slides/internal/llm"
)

// slideExpandLLMResponse is the JSON shape the LLM fills per slide.
type slideExpandLLMResponse struct {
	Slides []slideExpandItem `json:"slides"`
}

type slideExpandItem struct {
	Role         string         `json:"role"`
	Title        string         `json:"title"`
	BodyPoints   []string       `json:"body_points"`
	SpeakerNotes string         `json:"speaker_notes"`
	NeedsChart   bool           `json:"needs_chart"`
	NeedsDiagram bool           `json:"needs_diagram"`
	NeedsPhoto   bool           `json:"needs_photo"`
	NeedsIcons   bool           `json:"needs_icons"`
	ChartQuery   string         `json:"chart_query"`
	DiagramQuery string         `json:"diagram_query"`
	PhotoQuery   string         `json:"photo_query"`
	IconQueries  []string       `json:"icon_queries"`
	KPIValues    []kpiValue     `json:"kpi_values"`
}

type kpiValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
}

var expandSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "slides": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "role":          {"type": "string"},
          "title":         {"type": "string"},
          "body_points":   {"type": "array", "items": {"type": "string"}, "maxItems": 5},
          "speaker_notes": {"type": "string"},
          "needs_chart":   {"type": "boolean"},
          "needs_diagram": {"type": "boolean"},
          "needs_photo":   {"type": "boolean"},
          "needs_icons":   {"type": "boolean"},
          "chart_query":   {"type": "string"},
          "diagram_query": {"type": "string"},
          "photo_query":   {"type": "string"},
          "icon_queries":  {"type": "array", "items": {"type": "string"}, "maxItems": 4},
          "kpi_values":    {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "label": {"type": "string"},
                "value": {"type": "string"},
                "delta": {"type": "string"}
              },
              "required": ["label","value"]
            },
            "maxItems": 4
          }
        },
        "required": ["role","title","body_points","speaker_notes"]
      }
    }
  },
  "required": ["slides"]
}`)

// Expander expands a DeckPlan's sections into fully fleshed SlideNodes.
type Expander struct {
	provider llm.Provider
}

// NewExpander creates a slide Expander.
func NewExpander(provider llm.Provider) *Expander {
	return &Expander{provider: provider}
}

// Expand calls the LLM once per section to fill each slide with content and
// sets asset flags (needs_chart, needs_diagram, etc.) for downstream engines.
func (e *Expander) Expand(
	ctx context.Context,
	plan *barqv1.DeckPlan,
	parsed *intent.ParsedIntent,
) ([]*barqv1.SlideNode, error) {
	var allSlides []*barqv1.SlideNode
	slideIndex := 0

	for _, section := range plan.GetSections() {
		slides, err := e.expandSection(ctx, plan, section, parsed, slideIndex)
		if err != nil {
			return nil, fmt.Errorf("expander: section %q: %w", section.GetTitle(), err)
		}

		// Populate section slide IDs.
		ids := make([]string, 0, len(slides))
		for _, s := range slides {
			ids = append(ids, s.GetID())
		}
		section.SlideIDs = ids

		allSlides = append(allSlides, slides...)
		slideIndex += len(slides)
	}

	return allSlides, nil
}

func (e *Expander) expandSection(
	ctx context.Context,
	plan *barqv1.DeckPlan,
	section *barqv1.DeckSection,
	parsed *intent.ParsedIntent,
	startIndex int,
) ([]*barqv1.SlideNode, error) {
	userMsg := buildExpandPrompt(plan, section, parsed)

	resp, err := e.provider.GenerateStructured(ctx, llm.GenerateStructuredRequest{
		System: expandSystemPrompt,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: userMsg},
		},
		ToolName:        "output_formatter",
		ToolDescription: "Expand the section into detailed slide content.",
		Schema:          expandSchema,
		MaxTokens:       4096,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM expand call: %w", err)
	}

	var llmResp slideExpandLLMResponse
	if err := resp.UnmarshalInto(&llmResp); err != nil {
		return nil, fmt.Errorf("decoding expand response: %w", err)
	}

	return itemsToSlideNodes(llmResp.Slides, plan.GetRequestID(), startIndex), nil
}

func itemsToSlideNodes(items []slideExpandItem, requestID string, startIndex int) []*barqv1.SlideNode {
	nodes := make([]*barqv1.SlideNode, 0, len(items))

	for i, item := range items {
		node := &barqv1.SlideNode{
			ID:           uuid.NewString(),
			RequestID:    requestID,
			Index:        int32(startIndex + i),
			Role:         stringToSlideRole(item.Role),
			SpeakerNotes: item.SpeakerNotes,
		}

		// Title block.
		node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
			ID:       uuid.NewString(),
			Role:     "heading",
			Text:     item.Title,
			Emphasis: 3,
		})

		// Body bullet points.
		for _, bp := range item.BodyPoints {
			node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
				ID:   uuid.NewString(),
				Role: "body",
				Text: bp,
			})
		}

		// KPI values → body blocks with kpi role.
		for _, kpi := range item.KPIValues {
			node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
				ID:       uuid.NewString(),
				Role:     "kpi_value",
				Text:     fmt.Sprintf("%s|%s|%s", kpi.Label, kpi.Value, kpi.Delta),
				Emphasis: 2,
			})
		}

		// Asset slots.
		if item.NeedsChart && item.ChartQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				ID:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_CHART_BAR,
				Query: item.ChartQuery,
			})
		}

		if item.NeedsDiagram && item.DiagramQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				ID:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_DIAGRAM_FLOWCHART,
				Query: item.DiagramQuery,
			})
		}

		if item.NeedsPhoto && item.PhotoQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				ID:       uuid.NewString(),
				Type:     barqv1.AssetType_ASSET_TYPE_PHOTO,
				Query:    item.PhotoQuery,
				ImageUrl: "", // resolved in Phase 13
			})
		}

		for _, iq := range item.IconQueries {
			if iq == "" {
				continue
			}
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				ID:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_ICON,
				Query: iq,
			})
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func stringToSlideRole(s string) barqv1.SlideRole {
	switch strings.ToLower(s) {
	case "title":
		return barqv1.SlideRole_SLIDE_ROLE_TITLE
	case "section_divider", "divider":
		return barqv1.SlideRole_SLIDE_ROLE_SECTION_DIVIDER
	case "content":
		return barqv1.SlideRole_SLIDE_ROLE_CONTENT
	case "data":
		return barqv1.SlideRole_SLIDE_ROLE_DATA
	case "diagram":
		return barqv1.SlideRole_SLIDE_ROLE_DIAGRAM
	case "image_hero", "hero":
		return barqv1.SlideRole_SLIDE_ROLE_IMAGE_HERO
	case "quote":
		return barqv1.SlideRole_SLIDE_ROLE_QUOTE
	case "kpi":
		return barqv1.SlideRole_SLIDE_ROLE_KPI
	case "timeline":
		return barqv1.SlideRole_SLIDE_ROLE_TIMELINE
	case "comparison":
		return barqv1.SlideRole_SLIDE_ROLE_COMPARISON
	case "closing":
		return barqv1.SlideRole_SLIDE_ROLE_CLOSING
	default:
		return barqv1.SlideRole_SLIDE_ROLE_CONTENT
	}
}

const expandSystemPrompt = `You are an expert content strategist and presentation writer.

Expand the given section into detailed slides. For each slide:
- Write concise, impactful content (max 5 bullet points per content slide).
- Flag if the slide needs a chart (data comparison/trend), diagram (process/relationship), photo (scene/concept), or icons (feature bullets).
- Provide specific, searchable queries for each asset.
- Write speaker notes that elaborate on the talking points.
- KPI slides: extract 2-4 key metrics with labels, values, and trend deltas.

CRITICAL: Never include emoji in any text content. Use only plain text.

Respond using the output_formatter tool.`

func buildExpandPrompt(plan *barqv1.DeckPlan, section *barqv1.DeckSection, parsed *intent.ParsedIntent) string {
	return fmt.Sprintf(`Expand this presentation section into detailed slides:

**Deck title:** %s
**Section:** %s (section %d of %d)
**Audience:** %s | **Tone:** %s | **Domain:** %s
**Topic tags:** %s

Generate content for approximately %d slides in this section.
The section should maintain narrative coherence within the %s arc.`,
		plan.GetTitle(),
		section.GetTitle(),
		section.GetIndex()+1,
		len(plan.GetSections()),
		parsed.Audience,
		parsed.Tone,
		parsed.Domain,
		strings.Join(plan.GetTopicTags(), ", "),
		max(1, plan.GetTotalSlides()/int32(max(1, len(plan.GetSections())))),
		plan.GetArc().String(),
	)
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
