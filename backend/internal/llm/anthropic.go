package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultAnthropicMaxTokens   = 8192
	defaultAnthropicTemperature = 0.7
	// outputFormatterTool is the canonical tool name that forces structured output.
	outputFormatterTool = "output_formatter"
)

// AnthropicAdapter implements Provider using the Anthropic Claude API.
// Structured output is achieved by forcing tool_use with a single tool whose
// input_schema mirrors the requested JSON Schema.
type AnthropicAdapter struct {
	client *anthropic.Client
	model  string
	cfg    Config
}

// NewAnthropicAdapter constructs an AnthropicAdapter from Config.
func NewAnthropicAdapter(cfg Config) (*AnthropicAdapter, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic: api key is required")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	model := cfg.Model
	if model == "" {
		model = "claude-opus-4-6"
	}

	return &AnthropicAdapter{client: &client, model: model, cfg: cfg}, nil
}

// Name returns the provider identifier.
func (a *AnthropicAdapter) Name() ProviderName { return ProviderAnthropic }

// GenerateText produces a plain-text completion.
func (a *AnthropicAdapter) GenerateText(ctx context.Context, req GenerateTextRequest) (*TextResponse, error) {
	maxTokens := a.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = defaultAnthropicMaxTokens
	}

	temp := a.cfg.Temperature
	if req.Temperature != nil {
		temp = *req.Temperature
	}
	if temp == 0 {
		temp = defaultAnthropicTemperature
	}

	messages := toAnthropicMessages(req.Messages)

	params := anthropic.MessageNewParams{
		Model:       anthropic.F(anthropic.Model(a.model)),
		MaxTokens:   anthropic.F(int64(maxTokens)),
		Temperature: anthropic.F(temp),
		Messages:    anthropic.F(messages),
	}

	if req.System != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			{Type: anthropic.F(anthropic.TextBlockParamTypeText), Text: anthropic.F(req.System)},
		})
	}

	start := time.Now()
	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic GenerateText: %w", err)
	}
	_ = time.Since(start)

	text := extractAnthropicText(resp)

	return &TextResponse{
		Text:         text,
		InputTokens:  int(resp.Usage.InputTokens),
		OutputTokens: int(resp.Usage.OutputTokens),
		Model:        string(resp.Model),
	}, nil
}

// GenerateStructured forces structured JSON output by binding the model to
// the `output_formatter` tool and setting tool_choice = {"type":"tool","name":"output_formatter"}.
func (a *AnthropicAdapter) GenerateStructured(ctx context.Context, req GenerateStructuredRequest) (*StructuredResponse, error) {
	maxTokens := a.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = defaultAnthropicMaxTokens
	}

	toolName := req.ToolName
	if toolName == "" {
		toolName = outputFormatterTool
	}

	toolDesc := req.ToolDescription
	if toolDesc == "" {
		toolDesc = "Format the response as structured JSON conforming to the provided schema."
	}

	// Build the tool with the caller's JSON schema as input_schema.
	var inputSchema interface{}
	if len(req.Schema) > 0 {
		if err := json.Unmarshal(req.Schema, &inputSchema); err != nil {
			return nil, fmt.Errorf("anthropic: invalid schema: %w", err)
		}
	} else {
		inputSchema = map[string]interface{}{"type": "object"}
	}

	tool := anthropic.ToolParam{
		Name:        anthropic.F(toolName),
		Description: anthropic.F(toolDesc),
		InputSchema: anthropic.F(anthropic.ToolInputSchemaParam{
			Type:       anthropic.F(anthropic.ToolInputSchemaTypeObject),
			Properties: anthropic.F(inputSchema),
		}),
	}

	messages := toAnthropicMessages(req.Messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.Model(a.model)),
		MaxTokens: anthropic.F(int64(maxTokens)),
		Messages:  anthropic.F(messages),
		Tools:     anthropic.F([]anthropic.ToolParam{tool}),
		ToolChoice: anthropic.F[anthropic.ToolChoiceUnionParam](anthropic.ToolChoiceToolParam{
			Type: anthropic.F(anthropic.ToolChoiceToolTypeToolChoice),
			Name: anthropic.F(toolName),
		}),
	}

	if req.System != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			{Type: anthropic.F(anthropic.TextBlockParamTypeText), Text: anthropic.F(req.System)},
		})
	}

	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic GenerateStructured: %w", err)
	}

	raw, err := extractAnthropicToolInput(resp, toolName)
	if err != nil {
		return nil, fmt.Errorf("anthropic: extracting tool input: %w", err)
	}

	return &StructuredResponse{
		Raw:          raw,
		InputTokens:  int(resp.Usage.InputTokens),
		OutputTokens: int(resp.Usage.OutputTokens),
		Model:        string(resp.Model),
	}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func toAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case RoleAssistant:
			out = append(out, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
		// System role is handled separately via params.System.
	}
	return out
}

func extractAnthropicText(resp *anthropic.Message) string {
	for _, block := range resp.Content {
		if block.Type == anthropic.ContentBlockTypeText {
			return block.Text
		}
	}
	return ""
}

func extractAnthropicToolInput(resp *anthropic.Message, toolName string) (json.RawMessage, error) {
	for _, block := range resp.Content {
		if block.Type == anthropic.ContentBlockTypeToolUse && block.Name == toolName {
			raw, err := json.Marshal(block.Input)
			if err != nil {
				return nil, fmt.Errorf("marshalling tool input: %w", err)
			}
			return raw, nil
		}
	}
	return nil, fmt.Errorf("tool %q not found in response (stop_reason=%s)", toolName, resp.StopReason)
}
