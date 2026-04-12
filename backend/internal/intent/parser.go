// Package intent parses a raw user prompt into a structured IntentSpec
// with enriched audience, tone, domain, and slide-count signals.
package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/google/uuid"
)

const (
	maxPromptRunes   = 4000
	defaultSlideCount = 10
)

// ParsedIntent is the enriched output of the intent parsing step.
type ParsedIntent struct {
	RequestID  string
	Prompt     string
	Audience   string
	Tone       string
	Domain     string
	SlideCount int
	Locale     string
	TopicTags  []string
}

// intentLLMResponse mirrors the JSON schema the LLM must fill.
type intentLLMResponse struct {
	Audience   string   `json:"audience"`
	Tone       string   `json:"tone"`
	Domain     string   `json:"domain"`
	SlideCount int      `json:"slide_count"`
	TopicTags  []string `json:"topic_tags"`
}

var intentSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "audience":    {"type": "string", "description": "Target audience, e.g. 'C-suite executives'"},
    "tone":        {"type": "string", "enum": ["formal","casual","inspirational","technical","academic","storytelling"]},
    "domain":      {"type": "string", "description": "Domain or industry, e.g. 'fintech', 'healthcare'"},
    "slide_count": {"type": "integer", "minimum": 5, "maximum": 30},
    "topic_tags":  {"type": "array", "items": {"type": "string"}, "maxItems": 8}
  },
  "required": ["audience","tone","domain","slide_count","topic_tags"]
}`)

// Parser extracts structured intent from a free-form user prompt.
type Parser struct {
	provider llm.Provider
}

// NewParser creates a Parser backed by the given LLM provider.
func NewParser(provider llm.Provider) *Parser {
	return &Parser{provider: provider}
}

// Parse enriches the raw IntentSpec with audience, tone, domain, and tags.
// If the spec already has non-empty values, they override the LLM inference.
func (p *Parser) Parse(ctx context.Context, spec *barqv1.IntentSpec) (*ParsedIntent, error) {
	if spec == nil {
		return nil, fmt.Errorf("intent parser: nil IntentSpec")
	}

	prompt := sanitizePrompt(spec.GetPrompt())
	if prompt == "" {
		return nil, fmt.Errorf("intent parser: empty prompt")
	}

	requestID := spec.GetRequestID()
	if requestID == "" {
		requestID = uuid.NewString()
	}

	locale := spec.GetLocale()
	if locale == "" {
		locale = "en-US"
	}

	// If all fields are already set, skip LLM call.
	if spec.GetAudience() != "" && spec.GetTone() != "" && spec.GetDomain() != "" {
		slideCount := int(spec.GetSlideCount())
		if slideCount == 0 {
			slideCount = defaultSlideCount
		}
		return &ParsedIntent{
			RequestID:  requestID,
			Prompt:     prompt,
			Audience:   spec.GetAudience(),
			Tone:       spec.GetTone(),
			Domain:     spec.GetDomain(),
			SlideCount: slideCount,
			Locale:     locale,
		}, nil
	}

	// Build a context-enriched user message for the LLM.
	userMsg := buildIntentPrompt(prompt, spec)

	resp, err := p.provider.GenerateStructured(ctx, llm.GenerateStructuredRequest{
		System: systemPrompt,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: userMsg},
		},
		ToolName:        "output_formatter",
		ToolDescription: "Extract the intent metadata from the user's presentation request.",
		Schema:          intentSchema,
		MaxTokens:       1024,
	})
	if err != nil {
		return nil, fmt.Errorf("intent parser: LLM call failed: %w", err)
	}

	var llmResp intentLLMResponse
	if err := resp.UnmarshalInto(&llmResp); err != nil {
		return nil, fmt.Errorf("intent parser: decoding response: %w", err)
	}

	// Merge: explicit spec fields take priority over LLM inference.
	audience := coalesce(spec.GetAudience(), llmResp.Audience, "general audience")
	tone := coalesce(spec.GetTone(), llmResp.Tone, "formal")
	domain := coalesce(spec.GetDomain(), llmResp.Domain, "general")

	slideCount := int(spec.GetSlideCount())
	if slideCount == 0 {
		slideCount = llmResp.SlideCount
	}
	if slideCount == 0 {
		slideCount = defaultSlideCount
	}

	return &ParsedIntent{
		RequestID:  requestID,
		Prompt:     prompt,
		Audience:   audience,
		Tone:       tone,
		Domain:     domain,
		SlideCount: slideCount,
		Locale:     locale,
		TopicTags:  llmResp.TopicTags,
	}, nil
}

const systemPrompt = `You are an expert presentation strategist.
Given a user's request for a presentation, extract the key metadata:
- audience: who will view this deck
- tone: presentation style
- domain: subject matter domain
- slide_count: appropriate number of slides (5–30)
- topic_tags: 3–8 keyword tags for theme classification

Always respond using the output_formatter tool.`

func buildIntentPrompt(prompt string, spec *barqv1.IntentSpec) string {
	var sb strings.Builder
	sb.WriteString("Presentation request:\n")
	sb.WriteString(prompt)

	if spec.GetAudience() != "" {
		fmt.Fprintf(&sb, "\n\nSpecified audience: %s", spec.GetAudience())
	}
	if spec.GetTone() != "" {
		fmt.Fprintf(&sb, "\nSpecified tone: %s", spec.GetTone())
	}
	if spec.GetDomain() != "" {
		fmt.Fprintf(&sb, "\nSpecified domain: %s", spec.GetDomain())
	}
	if spec.GetSlideCount() > 0 {
		fmt.Fprintf(&sb, "\nRequested slides: %d", spec.GetSlideCount())
	}
	return sb.String()
}

func sanitizePrompt(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxPromptRunes {
		runes := []rune(s)
		s = string(runes[:maxPromptRunes])
	}
	return s
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
