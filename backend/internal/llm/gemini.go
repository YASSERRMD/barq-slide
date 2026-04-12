package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

const (
	defaultGeminiModel       = "gemini-2.0-flash"
	defaultGeminiMaxTokens   = 8192
	defaultGeminiTemperature = 0.7
)

// GeminiAdapter implements Provider using the Google Gemini API.
// Structured output is forced via GenerationConfig.ResponseSchema +
// ResponseMimeType = "application/json".
type GeminiAdapter struct {
	client *genai.Client
	model  string
	cfg    Config
}

// NewGeminiAdapter constructs a GeminiAdapter from Config.
func NewGeminiAdapter(cfg Config) (*GeminiAdapter, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini: api key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: creating client: %w", err)
	}

	model := cfg.Model
	if model == "" {
		model = defaultGeminiModel
	}

	return &GeminiAdapter{client: client, model: model, cfg: cfg}, nil
}

// Name returns the provider identifier.
func (g *GeminiAdapter) Name() ProviderName { return ProviderGemini }

// GenerateText produces a plain-text completion.
func (g *GeminiAdapter) GenerateText(ctx context.Context, req GenerateTextRequest) (*TextResponse, error) {
	maxTokens := int32(g.cfg.MaxTokens)
	if req.MaxTokens > 0 {
		maxTokens = int32(req.MaxTokens)
	}
	if maxTokens == 0 {
		maxTokens = defaultGeminiMaxTokens
	}

	temp := float32(g.cfg.Temperature)
	if req.Temperature != nil {
		temp = float32(*req.Temperature)
	}
	if temp == 0 {
		temp = defaultGeminiTemperature
	}

	contents := toGeminiContents(req.Messages)

	config := &genai.GenerateContentConfig{
		MaxOutputTokens: maxTokens,
		Temperature:     &temp,
	}

	if req.System != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.System}},
		}
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini GenerateText: %w", err)
	}

	text := extractGeminiText(result)

	return &TextResponse{
		Text:         text,
		InputTokens:  int(result.UsageMetadata.PromptTokenCount),
		OutputTokens: int(result.UsageMetadata.CandidatesTokenCount),
		Model:        g.model,
	}, nil
}

// GenerateStructured forces JSON output via ResponseSchema + ResponseMimeType.
func (g *GeminiAdapter) GenerateStructured(ctx context.Context, req GenerateStructuredRequest) (*StructuredResponse, error) {
	maxTokens := int32(g.cfg.MaxTokens)
	if req.MaxTokens > 0 {
		maxTokens = int32(req.MaxTokens)
	}
	if maxTokens == 0 {
		maxTokens = defaultGeminiMaxTokens
	}

	// Convert the caller's JSON Schema to a genai.Schema.
	schema, err := jsonSchemaToGemini(req.Schema)
	if err != nil {
		return nil, fmt.Errorf("gemini: converting schema: %w", err)
	}

	temp := float32(defaultGeminiTemperature)
	config := &genai.GenerateContentConfig{
		MaxOutputTokens:  maxTokens,
		Temperature:      &temp,
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	if req.System != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.System}},
		}
	}

	contents := toGeminiContents(req.Messages)

	result, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini GenerateStructured: %w", err)
	}

	raw := json.RawMessage(extractGeminiText(result))
	if !json.Valid(raw) {
		return nil, fmt.Errorf("gemini: response is not valid JSON: %s", string(raw))
	}

	return &StructuredResponse{
		Raw:          raw,
		InputTokens:  int(result.UsageMetadata.PromptTokenCount),
		OutputTokens: int(result.UsageMetadata.CandidatesTokenCount),
		Model:        g.model,
	}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func toGeminiContents(msgs []Message) []*genai.Content {
	out := make([]*genai.Content, 0, len(msgs))
	for _, m := range msgs {
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		if m.Role == RoleSystem {
			continue // handled via SystemInstruction
		}
		out = append(out, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: m.Content}},
		})
	}
	return out
}

func extractGeminiText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}
	cand := resp.Candidates[0]
	if cand.Content == nil || len(cand.Content.Parts) == 0 {
		return ""
	}
	return cand.Content.Parts[0].Text
}

// jsonSchemaToGemini converts a JSON Schema (as json.RawMessage) into a
// *genai.Schema. Only the essential fields (type, properties, required,
// items, enum) are mapped.
func jsonSchemaToGemini(raw json.RawMessage) (*genai.Schema, error) {
	if len(raw) == 0 {
		return &genai.Schema{Type: genai.TypeObject}, nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing JSON schema: %w", err)
	}

	return mapToGeminiSchema(m), nil
}

func mapToGeminiSchema(m map[string]interface{}) *genai.Schema {
	s := &genai.Schema{}

	if t, ok := m["type"].(string); ok {
		switch t {
		case "object":
			s.Type = genai.TypeObject
		case "array":
			s.Type = genai.TypeArray
		case "string":
			s.Type = genai.TypeString
		case "number", "integer":
			s.Type = genai.TypeNumber
		case "boolean":
			s.Type = genai.TypeBoolean
		}
	}

	if desc, ok := m["description"].(string); ok {
		s.Description = desc
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		s.Properties = make(map[string]*genai.Schema, len(props))
		for k, v := range props {
			if vm, ok := v.(map[string]interface{}); ok {
				s.Properties[k] = mapToGeminiSchema(vm)
			}
		}
	}

	if req, ok := m["required"].([]interface{}); ok {
		for _, r := range req {
			if rs, ok := r.(string); ok {
				s.Required = append(s.Required, rs)
			}
		}
	}

	if items, ok := m["items"].(map[string]interface{}); ok {
		s.Items = mapToGeminiSchema(items)
	}

	if enums, ok := m["enum"].([]interface{}); ok {
		for _, e := range enums {
			if es, ok := e.(string); ok {
				s.Enum = append(s.Enum, es)
			}
		}
	}

	return s
}
