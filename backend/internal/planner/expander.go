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
	Role         string     `json:"role"`
	LayoutType   string     `json:"layout_type"`
	Title        string     `json:"title"`
	Subtitle     string     `json:"subtitle"`
	BodyPoints   []string   `json:"body_points"`
	LeftPoints   []string   `json:"left_points"`
	RightPoints  []string   `json:"right_points"`
	QuoteText    string     `json:"quote_text"`
	QuoteAuthor  string     `json:"quote_author"`
	SpeakerNotes string     `json:"speaker_notes"`
	NeedsChart   bool       `json:"needs_chart"`
	NeedsDiagram bool       `json:"needs_diagram"`
	NeedsPhoto   bool       `json:"needs_photo"`
	NeedsIcons   bool       `json:"needs_icons"`
	ChartQuery   string     `json:"chart_query"`
	DiagramQuery string     `json:"diagram_query"`
	PhotoQuery   string     `json:"photo_query"`
	IconQueries  []string   `json:"icon_queries"`
	KPIValues    []kpiValue `json:"kpi_values"`
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
          "role":          {"type": "string", "enum": ["title","content","section_divider","quote","kpi","comparison","closing"]},
          "layout_type":   {"type": "string", "enum": ["title_center","bullet_list","two_column","quote_full","kpi_cards","section_divider","closing"]},
          "title":         {"type": "string"},
          "subtitle":      {"type": "string"},
          "body_points":   {"type": "array", "items": {"type": "string"}, "maxItems": 6},
          "left_points":   {"type": "array", "items": {"type": "string"}, "maxItems": 4},
          "right_points":  {"type": "array", "items": {"type": "string"}, "maxItems": 4},
          "quote_text":    {"type": "string"},
          "quote_author":  {"type": "string"},
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
        "required": ["role","layout_type","title","body_points","speaker_notes"]
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
			ids = append(ids, s.GetId())
		}
		section.SlideIds = ids

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

	return itemsToSlideNodes(llmResp.Slides, plan.GetRequestId(), startIndex), nil
}

func itemsToSlideNodes(items []slideExpandItem, requestID string, startIndex int) []*barqv1.SlideNode {
	nodes := make([]*barqv1.SlideNode, 0, len(items))

	for i, item := range items {
		layoutID := layoutIDFromType(item.LayoutType, item.Role)

		node := &barqv1.SlideNode{
			Id:           uuid.NewString(),
			RequestId:    requestID,
			Index:        int32(startIndex + i),
			Role:         stringToSlideRole(item.Role),
			SpeakerNotes: item.SpeakerNotes,
			Layout: &barqv1.SlideLayout{
				LayoutId: layoutID,
			},
		}

		// Title block.
		node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
			Id:       uuid.NewString(),
			Role:     "heading",
			Text:     item.Title,
			Emphasis: 3,
		})

		// Subtitle block (if present).
		if item.Subtitle != "" {
			node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
				Id:   uuid.NewString(),
				Role: "subheading",
				Text: item.Subtitle,
			})
		}

		switch layoutID {
		case barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN:
			// Left column points.
			if len(item.LeftPoints) > 0 {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:   uuid.NewString(),
					Role: "bullet",
					Text: strings.Join(item.LeftPoints, "\n"),
				})
			} else if len(item.BodyPoints) > 0 {
				half := (len(item.BodyPoints) + 1) / 2
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:   uuid.NewString(),
					Role: "bullet",
					Text: strings.Join(item.BodyPoints[:half], "\n"),
				})
				if half < len(item.BodyPoints) {
					node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
						Id:   uuid.NewString(),
						Role: "bullet",
						Text: strings.Join(item.BodyPoints[half:], "\n"),
					})
				}
			}
			// Right column points.
			if len(item.RightPoints) > 0 {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:   uuid.NewString(),
					Role: "bullet",
					Text: strings.Join(item.RightPoints, "\n"),
				})
			}

		case barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL:
			qt := item.QuoteText
			if qt == "" && len(item.BodyPoints) > 0 {
				qt = item.BodyPoints[0]
			}
			if qt != "" {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:       uuid.NewString(),
					Role:     "quote",
					Text:     qt,
					Emphasis: 2,
				})
			}
			attr := item.QuoteAuthor
			if attr == "" && item.Subtitle != "" {
				attr = item.Subtitle
			}
			if attr != "" {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:   uuid.NewString(),
					Role: "attribution",
					Text: attr,
				})
			}

		case barqv1.LayoutID_LAYOUT_ID_KPI_CARDS:
			for _, kpi := range item.KPIValues {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:       uuid.NewString(),
					Role:     "kpi_value",
					Text:     fmt.Sprintf("%s|%s|%s", kpi.Label, kpi.Value, kpi.Delta),
					Emphasis: 2,
				})
			}
			// Fallback: if no KPI values, use body points as metric cards.
			if len(item.KPIValues) == 0 {
				for _, bp := range item.BodyPoints {
					node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
						Id:       uuid.NewString(),
						Role:     "kpi_value",
						Text:     bp,
						Emphasis: 2,
					})
				}
			}

		default:
			// bullet_list, section_divider, title_center, closing.
			for _, bp := range item.BodyPoints {
				node.Blocks = append(node.Blocks, &barqv1.ContentBlock{
					Id:   uuid.NewString(),
					Role: "bullet",
					Text: bp,
				})
			}
		}

		// Asset slots.
		if item.NeedsChart && item.ChartQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				Id:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_CHART_BAR,
				Query: item.ChartQuery,
			})
		}

		if item.NeedsDiagram && item.DiagramQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				Id:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_DIAGRAM_FLOWCHART,
				Query: item.DiagramQuery,
			})
		}

		if item.NeedsPhoto && item.PhotoQuery != "" {
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				Id:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_PHOTO,
				Query: item.PhotoQuery,
			})
		}

		for _, iq := range item.IconQueries {
			if iq == "" {
				continue
			}
			node.Assets = append(node.Assets, &barqv1.AssetSlot{
				Id:    uuid.NewString(),
				Type:  barqv1.AssetType_ASSET_TYPE_ICON,
				Query: iq,
			})
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// layoutIDFromType maps a layout_type string + role fallback to a proto LayoutID.
func layoutIDFromType(layoutType, role string) barqv1.LayoutID {
	switch strings.ToLower(layoutType) {
	case "title_center":
		return barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER
	case "two_column", "comparison":
		return barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN
	case "quote_full", "quote":
		return barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL
	case "kpi_cards", "kpi", "big_number":
		return barqv1.LayoutID_LAYOUT_ID_KPI_CARDS
	case "section_divider", "section":
		return barqv1.LayoutID_LAYOUT_ID_SECTION_DIVIDER
	case "closing":
		return barqv1.LayoutID_LAYOUT_ID_CLOSING
	case "bullet_list":
		return barqv1.LayoutID_LAYOUT_ID_BULLET_LIST
	}
	// Fallback: derive from role.
	switch strings.ToLower(role) {
	case "title":
		return barqv1.LayoutID_LAYOUT_ID_TITLE_CENTER
	case "section_divider", "divider":
		return barqv1.LayoutID_LAYOUT_ID_SECTION_DIVIDER
	case "quote":
		return barqv1.LayoutID_LAYOUT_ID_QUOTE_FULL
	case "kpi":
		return barqv1.LayoutID_LAYOUT_ID_KPI_CARDS
	case "comparison":
		return barqv1.LayoutID_LAYOUT_ID_TWO_COLUMN
	case "closing":
		return barqv1.LayoutID_LAYOUT_ID_CLOSING
	default:
		return barqv1.LayoutID_LAYOUT_ID_BULLET_LIST
	}
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

const expandSystemPrompt = `You are an expert content strategist and presentation designer.

Expand the given section into detailed slides. CRITICAL RULES:

LAYOUT VARIETY — every slide MUST use a different layout_type from its neighbors.
Distribute layouts across the deck:
- "title_center": only for the very first slide (deck title/cover)
- "section_divider": use for section-opening slides (large title, subtitle, no bullets)
- "bullet_list": standard content slide with 3-5 concise bullet points
- "two_column": use for comparisons, pros/cons, before/after, feature lists — populate left_points AND right_points
- "quote_full": use for impactful quotes, testimonials, key insights — populate quote_text and quote_author
- "kpi_cards": use for metrics, statistics, results — populate kpi_values with label/value/delta
- "closing": only for the final slide

CONTENT RULES:
- bullet_list: write 3-5 punchy bullets in body_points (each ≤12 words). Set needs_icons=true for feature slides.
- two_column: write 3-4 bullets per side in left_points / right_points. Give each column a contrasting angle.
- quote_full: one powerful sentence in quote_text (≤25 words), attributed in quote_author.
- kpi_cards: 2-4 metrics in kpi_values, each with a label, a big number/value, and a delta (e.g., "+23%", "↑ 2x").
- section_divider: use subtitle for a one-line teaser of what's coming.
- Always write speaker_notes (2-3 sentences elaborating on the slide).
- Flag needs_chart/needs_diagram/needs_photo/needs_icons with specific searchable queries.

NEVER include emoji. Use only plain text.

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
		max(int32(1), plan.GetTotalSlides()/max(int32(1), int32(len(plan.GetSections())))),
		plan.GetArc().String(),
	)
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
