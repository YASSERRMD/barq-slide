package llm_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/YASSERRMD/barq-slides/internal/config"
	"github.com/YASSERRMD/barq-slides/internal/llm"
)

// TestFactory_DefaultProvider verifies the factory resolves the configured
// default provider without needing real API keys (using a stub config).
func TestFactory_DefaultProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		wantName llm.ProviderName
		wantErr  bool
	}{
		{
			name:     "anthropic default",
			provider: "anthropic",
			wantErr:  true, // empty API key → construction error
		},
		{
			name:     "unknown provider",
			provider: "unknown_provider",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				LLM: config.LLMConfig{
					DefaultProvider: tc.provider,
				},
			}

			factory := llm.NewFactory(cfg)
			_, err := factory.Default()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMockProvider_GenerateText validates the mock's call recording.
func TestMockProvider_GenerateText(t *testing.T) {
	t.Parallel()

	mock := &llm.MockProvider{
		ProviderNameValue: "mock",
		TextResponses: []llm.TextResponse{
			{Text: "hello world", InputTokens: 10, OutputTokens: 3, Model: "mock-1"},
		},
	}

	resp, err := mock.GenerateText(context.Background(), llm.GenerateTextRequest{
		System:   "you are helpful",
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "hello world", resp.Text)
	assert.Len(t, mock.TextCalls, 1)
	assert.Equal(t, "you are helpful", mock.TextCalls[0].System)
}

// TestMockProvider_GenerateStructured validates structured response unmarshalling.
func TestMockProvider_GenerateStructured(t *testing.T) {
	t.Parallel()

	type planResult struct {
		Title string `json:"title"`
		Slides int   `json:"slides"`
	}

	payload := planResult{Title: "AI in Healthcare", Slides: 12}

	mock := &llm.MockProvider{
		StructuredResponses: []llm.StructuredResponse{
			{Raw: llm.MustMarshal(payload), InputTokens: 50, OutputTokens: 20},
		},
	}

	resp, err := mock.GenerateStructured(context.Background(), llm.GenerateStructuredRequest{
		System:   "you are a planner",
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "plan a deck"}},
		ToolName: "output_formatter",
		Schema:   json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"},"slides":{"type":"integer"}}}`),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	var got planResult
	require.NoError(t, resp.UnmarshalInto(&got))
	assert.Equal(t, "AI in Healthcare", got.Title)
	assert.Equal(t, 12, got.Slides)
}

// TestMockProvider_ExhaustedStubs verifies error on exhausted stub list.
func TestMockProvider_ExhaustedStubs(t *testing.T) {
	t.Parallel()

	mock := &llm.MockProvider{}

	_, err := mock.GenerateText(context.Background(), llm.GenerateTextRequest{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no more TextResponse stubs")
}

// TestMockProvider_MultiProviderRouting verifies the factory can build each
// provider type given valid (non-empty) credentials.
func TestMockProvider_MultiProviderRouting(t *testing.T) {
	t.Parallel()

	providers := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "openai",
			cfg: &config.Config{
				LLM: config.LLMConfig{
					DefaultProvider: "openai",
					OpenAI: config.OpenAIConfig{
						APIKey:  "sk-test",
						Model:   "gpt-4o",
						BaseURL: "https://api.openai.com/v1",
					},
				},
			},
		},
		{
			name: "xai",
			cfg: &config.Config{
				LLM: config.LLMConfig{
					DefaultProvider: "xai",
					XAI: config.XAIConfig{
						APIKey:  "xai-test",
						Model:   "grok-3",
						BaseURL: "https://api.x.ai/v1",
					},
				},
			},
		},
		{
			name: "minimax",
			cfg: &config.Config{
				LLM: config.LLMConfig{
					DefaultProvider: "minimax",
					Minimax: config.MinimaxConfig{
						APIKey:  "minimax-test",
						Model:   "abab6.5s-chat",
						BaseURL: "https://api.minimax.chat/v1",
					},
				},
			},
		},
	}

	for _, tc := range providers {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			factory := llm.NewFactory(tc.cfg)
			p, err := factory.Default()
			require.NoError(t, err)
			assert.Equal(t, llm.ProviderName(tc.name), p.Name())
		})
	}
}
