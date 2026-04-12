// Package llm provides a unified multi-provider LLM abstraction for barq-slides.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ProviderName enumerates supported LLM backends.
type ProviderName string

const (
	ProviderAnthropic ProviderName = "anthropic"
	ProviderGemini    ProviderName = "gemini"
	ProviderOpenAI    ProviderName = "openai"
	ProviderXAI       ProviderName = "xai"
	ProviderMinimax   ProviderName = "minimax"
	ProviderOllama    ProviderName = "ollama"
)

// Config holds provider-level configuration.
type Config struct {
	// Provider identifies which backend to use.
	Provider ProviderName

	// APIKey is the provider's API key.
	APIKey string

	// BaseURL overrides the provider's default API endpoint.
	// Used for xAI, Minimax, and custom OpenAI-compatible deployments.
	BaseURL string

	// Model is the model ID/name to use.
	Model string

	// MaxTokens caps the response length.
	MaxTokens int

	// Temperature controls sampling randomness (0.0 – 2.0).
	Temperature float64

	// Timeout overrides the default request timeout.
	Timeout time.Duration
}

// Message represents a single turn in a conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Role identifies the author of a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// GenerateTextRequest is a plain-text generation request.
type GenerateTextRequest struct {
	// System prompt sent before the user messages.
	System string

	// Messages in the conversation.
	Messages []Message

	// MaxTokens overrides Config.MaxTokens for this request.
	MaxTokens int

	// Temperature overrides Config.Temperature for this request.
	Temperature *float64
}

// GenerateStructuredRequest forces the provider to return valid JSON
// conforming to the given schema via tool use / function calling.
type GenerateStructuredRequest struct {
	// System prompt sent before the user messages.
	System string

	// Messages in the conversation.
	Messages []Message

	// ToolName is the name of the tool/function the model must call.
	ToolName string

	// ToolDescription describes the tool to the model.
	ToolDescription string

	// Schema is the JSON Schema the response object must conform to.
	// Passed as a json.RawMessage to avoid double-encoding.
	Schema json.RawMessage

	// MaxTokens overrides Config.MaxTokens for this request.
	MaxTokens int
}

// TextResponse is a plain-text generation response.
type TextResponse struct {
	// Text is the generated content.
	Text string

	// InputTokens is the number of prompt tokens used.
	InputTokens int

	// OutputTokens is the number of generated tokens.
	OutputTokens int

	// Model is the actual model that served the request.
	Model string
}

// StructuredResponse is a structured (JSON tool-use) generation response.
type StructuredResponse struct {
	// Raw is the raw JSON bytes of the tool call arguments.
	Raw json.RawMessage

	// InputTokens is the number of prompt tokens used.
	InputTokens int

	// OutputTokens is the number of generated tokens.
	OutputTokens int

	// Model is the actual model that served the request.
	Model string
}

// UnmarshalInto decodes the raw JSON into dst.
func (r *StructuredResponse) UnmarshalInto(dst interface{}) error {
	if err := json.Unmarshal(r.Raw, dst); err != nil {
		return fmt.Errorf("unmarshalling structured response: %w", err)
	}
	return nil
}

// Provider is the unified interface all LLM adapters must satisfy.
type Provider interface {
	// Name returns the provider identifier.
	Name() ProviderName

	// GenerateText produces a plain-text completion.
	GenerateText(ctx context.Context, req GenerateTextRequest) (*TextResponse, error)

	// GenerateStructured forces the model to return a JSON object matching
	// the provided JSON Schema, using tool use / function calling as appropriate
	// for the backend.
	GenerateStructured(ctx context.Context, req GenerateStructuredRequest) (*StructuredResponse, error)
}

// UsageStats aggregates token usage for observability.
type UsageStats struct {
	Provider     ProviderName
	Model        string
	InputTokens  int
	OutputTokens int
	Latency      time.Duration
}
