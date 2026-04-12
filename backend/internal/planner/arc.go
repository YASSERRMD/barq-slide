// Package planner generates the high-level DeckPlan from a ParsedIntent.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/intent"
	"github.com/YASSERRMD/barq-slides/internal/llm"
)

// arcLLMResponse is the JSON the LLM fills for narrative arc planning.
type arcLLMResponse struct {
	Title    string       `json:"title"`
	Subtitle string       `json:"subtitle"`
	Arc      string       `json:"arc"`
	Sections []arcSection `json:"sections"`
}

type arcSection struct {
	Title      string   `json:"title"`
	SlideCount int      `json:"slide_count"`
	SlideRoles []string `json:"slide_roles"`
}

var arcSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "title":    {"type": "string"},
    "subtitle": {"type": "string"},
    "arc":      {"type": "string", "enum": ["problem_solution","before_after","data_story","pitch","tutorial","report"]},
    "sections": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title":       {"type": "string"},
          "slide_count": {"type": "integer", "minimum": 1, "maximum": 8},
          "slide_roles": {"type": "array", "items": {"type": "string"}}
        },
        "required": ["title","slide_count","slide_roles"]
      },
      "minItems": 2,
      "maxItems": 6
    }
  },
  "required": ["title","subtitle","arc","sections"]
}`)

// ArcPlanner generates the high-level narrative arc and section structure.
type ArcPlanner struct {
	provider llm.Provider
}

// NewArcPlanner creates an ArcPlanner backed by the given LLM provider.
func NewArcPlanner(provider llm.Provider) *ArcPlanner {
	return &ArcPlanner{provider: provider}
}

// Plan generates the DeckPlan (title, subtitle, arc, sections) from ParsedIntent.
func (a *ArcPlanner) Plan(ctx context.Context, parsed *intent.ParsedIntent) (*barqv1.DeckPlan, error) {
	userMsg := buildArcPrompt(parsed)

	resp, err := a.provider.GenerateStructured(ctx, llm.GenerateStructuredRequest{
		System: arcSystemPrompt,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: userMsg},
		},
		ToolName:        "output_formatter",
		ToolDescription: "Generate the deck plan with narrative arc and sections.",
		Schema:          arcSchema,
		MaxTokens:       2048,
	})
	if err != nil {
		return nil, fmt.Errorf("arc planner: LLM call failed: %w", err)
	}

	var llmResp arcLLMResponse
	if err := resp.UnmarshalInto(&llmResp); err != nil {
		return nil, fmt.Errorf("arc planner: decoding response: %w", err)
	}

	return buildDeckPlan(parsed, &llmResp), nil
}

func buildDeckPlan(parsed *intent.ParsedIntent, r *arcLLMResponse) *barqv1.DeckPlan {
	arc := stringToNarrativeArc(r.Arc)

	sections := make([]*barqv1.DeckSection, 0, len(r.Sections))
	totalSlides := 0

	for i, sec := range r.Sections {
		sections = append(sections, &barqv1.DeckSection{
			Index: int32(i),
			Title: sec.Title,
		})
		totalSlides += sec.SlideCount
	}

	return &barqv1.DeckPlan{
		RequestId:   parsed.RequestID,
		Title:       r.Title,
		Subtitle:    r.Subtitle,
		Arc:         arc,
		Sections:    sections,
		TotalSlides: int32(totalSlides),
		TopicTags:   parsed.TopicTags,
	}
}

func stringToNarrativeArc(s string) barqv1.NarrativeArc {
	switch strings.ToLower(s) {
	case "problem_solution":
		return barqv1.NarrativeArc_NARRATIVE_ARC_PROBLEM_SOLUTION
	case "before_after":
		return barqv1.NarrativeArc_NARRATIVE_ARC_BEFORE_AFTER
	case "data_story":
		return barqv1.NarrativeArc_NARRATIVE_ARC_DATA_STORY
	case "pitch":
		return barqv1.NarrativeArc_NARRATIVE_ARC_PITCH
	case "tutorial":
		return barqv1.NarrativeArc_NARRATIVE_ARC_TUTORIAL
	case "report":
		return barqv1.NarrativeArc_NARRATIVE_ARC_REPORT
	default:
		return barqv1.NarrativeArc_NARRATIVE_ARC_REPORT
	}
}

const arcSystemPrompt = `You are a world-class presentation strategist and information architect.

Your task is to create the high-level narrative plan for a presentation deck.

Guidelines:
- Choose the narrative arc that best serves the goal and audience.
- Create 2–6 logical sections that flow naturally.
- Each section should have 1–8 slides.
- Total slides should match the requested count.
- The title should be punchy and memorable (max 60 chars).
- The subtitle should expand the concept (max 120 chars).

Always respond using the output_formatter tool.`

func buildArcPrompt(parsed *intent.ParsedIntent) string {
	return fmt.Sprintf(`Create a presentation plan for the following request:

**Prompt:** %s
**Audience:** %s
**Tone:** %s
**Domain:** %s
**Target slide count:** %d
**Topic tags:** %s

Design a narrative arc that will resonate with this specific audience and achieve the presentation's goal.`,
		parsed.Prompt,
		parsed.Audience,
		parsed.Tone,
		parsed.Domain,
		parsed.SlideCount,
		strings.Join(parsed.TopicTags, ", "),
	)
}
