package planner_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/intent"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/YASSERRMD/barq-slides/internal/planner"
)

// TestArcPlanner_Plan validates that a canned LLM response is correctly
// translated into a DeckPlan with the right arc and section count.
func TestArcPlanner_Plan(t *testing.T) {
	t.Parallel()

	arcPayload := map[string]interface{}{
		"title":    "AI Revolution in Healthcare",
		"subtitle": "How machine learning is transforming patient outcomes",
		"arc":      "data_story",
		"sections": []map[string]interface{}{
			{"title": "The Problem", "slide_count": 2, "slide_roles": []string{"content", "data"}},
			{"title": "AI Solutions", "slide_count": 3, "slide_roles": []string{"content", "diagram", "kpi"}},
			{"title": "Roadmap", "slide_count": 2, "slide_roles": []string{"timeline", "closing"}},
		},
	}

	mock := &llm.MockProvider{
		StructuredResponses: []llm.StructuredResponse{
			{Raw: llm.MustMarshal(arcPayload)},
		},
	}

	arcPlanner := planner.NewArcPlanner(mock)
	parsed := &intent.ParsedIntent{
		RequestId:  "req-001",
		Prompt:     "Create a presentation about AI in healthcare",
		Audience:   "Medical executives",
		Tone:       "formal",
		Domain:     "healthcare",
		SlideCount: 7,
		TopicTags:  []string{"AI", "healthcare", "ML", "patient outcomes"},
	}

	plan, err := arcPlanner.Plan(context.Background(), parsed)
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, "AI Revolution in Healthcare", plan.GetTitle())
	assert.Equal(t, barqv1.NarrativeArc_NARRATIVE_ARC_DATA_STORY, plan.GetArc())
	assert.Len(t, plan.GetSections(), 3)
	assert.Equal(t, "The Problem", plan.GetSections()[0].GetTitle())
	assert.Equal(t, int32(7), plan.GetTotalSlides())
}

// TestExpander_Expand verifies that slide nodes are created with correct
// asset slots when the LLM flags needs_chart and needs_diagram.
func TestExpander_Expand(t *testing.T) {
	t.Parallel()

	expandPayload := map[string]interface{}{
		"slides": []map[string]interface{}{
			{
				"role":          "title",
				"title":         "The Problem",
				"body_points":   []string{"Healthcare costs are rising", "Diagnosis delays cost lives"},
				"speaker_notes": "Open with a compelling statistic.",
				"needs_chart":   false,
				"needs_diagram": false,
				"needs_photo":   true,
				"needs_icons":   false,
				"photo_query":   "hospital corridor patient care",
			},
			{
				"role":          "data",
				"title":         "The Numbers",
				"body_points":   []string{"30% diagnostic error rate", "$250B annual waste"},
				"speaker_notes": "Let the data speak.",
				"needs_chart":   true,
				"needs_diagram": false,
				"needs_photo":   false,
				"needs_icons":   false,
				"chart_query":   "diagnostic error rates by specialty 2018-2024",
			},
		},
	}

	mock := &llm.MockProvider{
		StructuredResponses: []llm.StructuredResponse{
			{Raw: llm.MustMarshal(expandPayload)},
		},
	}

	expander := planner.NewExpander(mock)

	plan := &barqv1.DeckPlan{
		RequestId:   "req-001",
		Title:       "AI in Healthcare",
		TotalSlides: 2,
		Sections: []*barqv1.DeckSection{
			{Index: 0, Title: "The Problem"},
		},
	}

	parsed := &intent.ParsedIntent{
		Audience:  "Medical executives",
		Tone:      "formal",
		Domain:    "healthcare",
		TopicTags: []string{"AI", "healthcare"},
	}

	slides, err := expander.Expand(context.Background(), plan, parsed)
	require.NoError(t, err)
	require.Len(t, slides, 2)

	// First slide: title role, has photo asset.
	assert.Equal(t, barqv1.SlideRole_SLIDE_ROLE_TITLE, slides[0].GetRole())
	require.Len(t, slides[0].GetAssets(), 1)
	assert.Equal(t, barqv1.AssetType_ASSET_TYPE_PHOTO, slides[0].GetAssets()[0].GetType())

	// Second slide: data role, has chart asset.
	assert.Equal(t, barqv1.SlideRole_SLIDE_ROLE_DATA, slides[1].GetRole())
	require.Len(t, slides[1].GetAssets(), 1)
	assert.Equal(t, barqv1.AssetType_ASSET_TYPE_CHART_BAR, slides[1].GetAssets()[0].GetType())
	assert.Equal(t, "diagnostic error rates by specialty 2018-2024", slides[1].GetAssets()[0].GetQuery())
}

// TestExpander_SectionIDsPopulated confirms that section.SlideIDs is populated
// after expansion.
func TestExpander_SectionIDsPopulated(t *testing.T) {
	t.Parallel()

	expandPayload := map[string]interface{}{
		"slides": []map[string]interface{}{
			{
				"role": "content", "title": "Slide 1",
				"body_points": []string{"point"}, "speaker_notes": "",
			},
		},
	}

	mock := &llm.MockProvider{
		StructuredResponses: []llm.StructuredResponse{
			{Raw: llm.MustMarshal(expandPayload)},
		},
	}

	section := &barqv1.DeckSection{Index: 0, Title: "Intro"}
	plan := &barqv1.DeckPlan{
		RequestId:   "req-002",
		TotalSlides: 1,
		Sections:    []*barqv1.DeckSection{section},
	}

	expander := planner.NewExpander(mock)
	slides, err := expander.Expand(context.Background(), plan, &intent.ParsedIntent{})
	require.NoError(t, err)
	require.Len(t, slides, 1)

	assert.Len(t, section.GetSlideIDs(), 1)
	assert.Equal(t, slides[0].GetId(), section.GetSlideIDs()[0])
}

// goldenPrompt_ArcSystem is a snapshot of the arc system prompt for regression testing.
func TestGoldenPrompt_ArcSystem(t *testing.T) {
	t.Parallel()

	// Verify the arc planner sends a system prompt containing key strategy terms.
	mock := &llm.MockProvider{
		StructuredResponses: []llm.StructuredResponse{
			{Raw: llm.MustMarshal(map[string]interface{}{
				"title": "T", "subtitle": "S", "arc": "report",
				"sections": []map[string]interface{}{
					{"title": "A", "slide_count": 2, "slide_roles": []string{"content", "content"}},
				},
			})},
		},
	}

	arcPlanner := planner.NewArcPlanner(mock)
	_, err := arcPlanner.Plan(context.Background(), &intent.ParsedIntent{
		Prompt: "test", Audience: "dev", Tone: "casual", Domain: "tech",
		SlideCount: 2, TopicTags: []string{"tech"},
	})
	require.NoError(t, err)
	require.Len(t, mock.StructuredCalls, 1)

	call := mock.StructuredCalls[0]
	assert.Contains(t, call.System, "narrative arc")
	assert.Equal(t, "output_formatter", call.ToolName)

	// Verify schema is valid JSON.
	var schema interface{}
	require.NoError(t, json.Unmarshal(call.Schema, &schema))
}
