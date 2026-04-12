package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

const (
	defaultOpenAIModel       = "gpt-4o"
	defaultOpenAIMaxTokens   = 8192
	defaultOpenAITemperature = 0.7
)

// OpenAICompatAdapter implements Provider using the OpenAI API (or any
// OpenAI-compatible endpoint). It forces structured output via
// ResponseFormatJSONSchema.
//
// This adapter routes to OpenAI, xAI (Grok), and Minimax by injecting
// the appropriate BaseURL in the Config.
type OpenAICompatAdapter struct {
	client   *openai.Client
	model    string
	cfg      Config
	provider ProviderName
}

// NewOpenAICompatAdapter constructs an OpenAICompatAdapter.
// The provider name is stored for observability; the actual routing
// is controlled by cfg.BaseURL.
func NewOpenAICompatAdapter(provider ProviderName, cfg Config) (*OpenAICompatAdapter, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%s: api key is required", provider)
	}

	clientCfg := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}

	model := cfg.Model
	if model == "" {
		switch provider {
		case ProviderXAI:
			model = "grok-3"
		case ProviderMinimax:
			model = "abab6.5s-chat"
		default:
			model = defaultOpenAIModel
		}
	}

	return &OpenAICompatAdapter{
		client:   openai.NewClientWithConfig(clientCfg),
		model:    model,
		cfg:      cfg,
		provider: provider,
	}, nil
}

// Name returns the provider identifier.
func (o *OpenAICompatAdapter) Name() ProviderName { return o.provider }

// GenerateText produces a plain-text completion.
func (o *OpenAICompatAdapter) GenerateText(ctx context.Context, req GenerateTextRequest) (*TextResponse, error) {
	maxTokens := o.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = defaultOpenAIMaxTokens
	}

	temp := float32(o.cfg.Temperature)
	if req.Temperature != nil {
		temp = float32(*req.Temperature)
	}
	if temp == 0 {
		temp = defaultOpenAITemperature
	}

	msgs := toOpenAIMessages(req.System, req.Messages)

	chatReq := openai.ChatCompletionRequest{
		Model:       o.model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: temp,
	}

	resp, err := o.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("%s GenerateText: %w", o.provider, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("%s: empty choices in response", o.provider)
	}

	return &TextResponse{
		Text:         resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		Model:        resp.Model,
	}, nil
}

// GenerateStructured forces JSON output via ResponseFormatJSONSchema.
// The schema is embedded in a named JSON schema object so the model knows
// exactly what structure to produce.
func (o *OpenAICompatAdapter) GenerateStructured(ctx context.Context, req GenerateStructuredRequest) (*StructuredResponse, error) {
	maxTokens := o.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = defaultOpenAIMaxTokens
	}

	toolName := req.ToolName
	if toolName == "" {
		toolName = outputFormatterTool
	}

	msgs := toOpenAIMessages(req.System, req.Messages)

	// Append a final user message instructing strict JSON output.
	if req.ToolDescription != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.ToolDescription,
		})
	}

	// Build a strict JSON schema for the response format.
	schemaObj := map[string]interface{}{
		"type":   "object",
		"strict": true,
	}
	if len(req.Schema) > 0 {
		var s map[string]interface{}
		if err := json.Unmarshal(req.Schema, &s); err == nil {
			schemaObj = s
			schemaObj["strict"] = true
		}
	}

	schemaBytes, err := json.Marshal(schemaObj)
	if err != nil {
		return nil, fmt.Errorf("%s: marshalling schema: %w", o.provider, err)
	}

	chatReq := openai.ChatCompletionRequest{
		Model:     o.model,
		Messages:  msgs,
		MaxTokens: maxTokens,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   toolName,
				Schema: json.RawMessage(schemaBytes),
				Strict: true,
			},
		},
	}

	resp, err := o.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("%s GenerateStructured: %w", o.provider, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("%s: empty choices in structured response", o.provider)
	}

	raw := json.RawMessage(resp.Choices[0].Message.Content)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%s: structured response is not valid JSON", o.provider)
	}

	return &StructuredResponse{
		Raw:          raw,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		Model:        resp.Model,
	}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func toOpenAIMessages(system string, msgs []Message) []openai.ChatCompletionMessage {
	out := make([]openai.ChatCompletionMessage, 0, len(msgs)+1)

	if system != "" {
		out = append(out, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		})
	}

	for _, m := range msgs {
		var role string
		switch m.Role {
		case RoleUser:
			role = openai.ChatMessageRoleUser
		case RoleAssistant:
			role = openai.ChatMessageRoleAssistant
		case RoleSystem:
			role = openai.ChatMessageRoleSystem
		default:
			role = openai.ChatMessageRoleUser
		}
		out = append(out, openai.ChatCompletionMessage{
			Role:    role,
			Content: m.Content,
		})
	}
	return out
}
