package llm

import (
	"fmt"

	"github.com/YASSERRMD/barq-slides/internal/config"
)

// Factory creates Provider instances from application config.
// It centralises provider construction and routes xAI / Minimax through
// the OpenAI-compat adapter by injecting provider-specific BaseURLs.
type Factory struct {
	cfg *config.Config
}

// NewFactory constructs a Factory from application config.
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{cfg: cfg}
}

// Default returns the Provider configured as the default.
func (f *Factory) Default() (Provider, error) {
	return f.Get(ProviderName(f.cfg.LLM.DefaultProvider))
}

// Get returns a Provider by name, constructing it from config.
func (f *Factory) Get(name ProviderName) (Provider, error) {
	switch name {
	case ProviderAnthropic, "":
		return f.anthropic()
	case ProviderGemini:
		return f.gemini()
	case ProviderOpenAI:
		return f.openai()
	case ProviderXAI:
		return f.xai()
	case ProviderMinimax:
		return f.minimax()
	default:
		return nil, fmt.Errorf("llm factory: unknown provider %q", name)
	}
}

// GetDynamic returns a Provider by name, substituting the API key, model and base URL at runtime.
// If provider is "anthropic" but a custom baseUrl is provided, it is treated as an
// OpenAI-compatible endpoint (e.g. z.ai, GLM, etc.) and routed through the OpenAI adapter.
func (f *Factory) GetDynamic(name ProviderName, apiKey, model, baseUrl string) (Provider, error) {
	switch name {
	case ProviderAnthropic, "":
		// If a custom base URL is set, it's an OpenAI-compatible provider — not native Anthropic.
		if baseUrl != "" {
			reqModel := model
			if reqModel == "" { reqModel = "gpt-4o" }
			return NewOpenAICompatAdapter(ProviderOpenAI, Config{
				Provider:    ProviderOpenAI,
				APIKey:      apiKey,
				BaseURL:     baseUrl,
				Model:       reqModel,
				MaxTokens:   defaultOpenAIMaxTokens,
				Temperature: defaultOpenAITemperature,
			})
		}
		reqModel := f.cfg.LLM.Anthropic.Model
		if model != "" { reqModel = model }
		return NewAnthropicAdapter(Config{
			Provider:    ProviderAnthropic,
			APIKey:      apiKey,
			Model:       reqModel,
			MaxTokens:   defaultAnthropicMaxTokens,
			Temperature: defaultAnthropicTemperature,
		})
	case ProviderGemini:
		reqModel := f.cfg.LLM.Gemini.Model
		if model != "" { reqModel = model }
		return NewGeminiAdapter(Config{
			Provider:    ProviderGemini,
			APIKey:      apiKey,
			Model:       reqModel,
			MaxTokens:   defaultGeminiMaxTokens,
			Temperature: defaultGeminiTemperature,
		})
	case ProviderOpenAI:
		reqModel := f.cfg.LLM.OpenAI.Model
		if model != "" { reqModel = model }
		reqBaseUrl := f.cfg.LLM.OpenAI.BaseURL
		if baseUrl != "" { reqBaseUrl = baseUrl }
		return NewOpenAICompatAdapter(ProviderOpenAI, Config{
			Provider:    ProviderOpenAI,
			APIKey:      apiKey,
			BaseURL:     reqBaseUrl,
			Model:       reqModel,
			MaxTokens:   defaultOpenAIMaxTokens,
			Temperature: defaultOpenAITemperature,
		})
	case ProviderXAI:
		reqModel := f.cfg.LLM.XAI.Model
		if model != "" { reqModel = model }
		reqBaseUrl := f.cfg.LLM.XAI.BaseURL
		if baseUrl != "" { reqBaseUrl = baseUrl }
		return NewOpenAICompatAdapter(ProviderXAI, Config{
			Provider:    ProviderXAI,
			APIKey:      apiKey,
			BaseURL:     reqBaseUrl,
			Model:       reqModel,
			MaxTokens:   defaultOpenAIMaxTokens,
			Temperature: defaultOpenAITemperature,
		})
	case ProviderMinimax:
		reqModel := f.cfg.LLM.Minimax.Model
		if model != "" { reqModel = model }
		reqBaseUrl := f.cfg.LLM.Minimax.BaseURL
		if baseUrl != "" { reqBaseUrl = baseUrl }
		return NewOpenAICompatAdapter(ProviderMinimax, Config{
			Provider:    ProviderMinimax,
			APIKey:      apiKey,
			BaseURL:     reqBaseUrl,
			Model:       reqModel,
			MaxTokens:   defaultOpenAIMaxTokens,
			Temperature: defaultOpenAITemperature,
		})
	case ProviderOllama:
		reqModel := "llama3"
		if model != "" { reqModel = model }
		reqBaseUrl := "http://localhost:11434/v1"
		if baseUrl != "" { reqBaseUrl = baseUrl }
		return NewOpenAICompatAdapter(ProviderOllama, Config{
			Provider:    ProviderOllama,
			APIKey:      apiKey, // Ollama technically doesn't need an API key usually, but handles it.
			BaseURL:     reqBaseUrl,
			Model:       reqModel,
			MaxTokens:   defaultOpenAIMaxTokens,
			Temperature: defaultOpenAITemperature,
		})
	default:
		return nil, fmt.Errorf("llm factory (dynamic): unknown provider %q", name)
	}
}

func (f *Factory) anthropic() (Provider, error) {
	return NewAnthropicAdapter(Config{
		Provider:    ProviderAnthropic,
		APIKey:      f.cfg.LLM.Anthropic.APIKey,
		Model:       f.cfg.LLM.Anthropic.Model,
		MaxTokens:   defaultAnthropicMaxTokens,
		Temperature: defaultAnthropicTemperature,
	})
}

func (f *Factory) gemini() (Provider, error) {
	return NewGeminiAdapter(Config{
		Provider:    ProviderGemini,
		APIKey:      f.cfg.LLM.Gemini.APIKey,
		Model:       f.cfg.LLM.Gemini.Model,
		MaxTokens:   defaultGeminiMaxTokens,
		Temperature: defaultGeminiTemperature,
	})
}

func (f *Factory) openai() (Provider, error) {
	return NewOpenAICompatAdapter(ProviderOpenAI, Config{
		Provider:    ProviderOpenAI,
		APIKey:      f.cfg.LLM.OpenAI.APIKey,
		BaseURL:     f.cfg.LLM.OpenAI.BaseURL,
		Model:       f.cfg.LLM.OpenAI.Model,
		MaxTokens:   defaultOpenAIMaxTokens,
		Temperature: defaultOpenAITemperature,
	})
}

// xai routes through OpenAI-compat adapter with the xAI BaseURL.
func (f *Factory) xai() (Provider, error) {
	return NewOpenAICompatAdapter(ProviderXAI, Config{
		Provider:    ProviderXAI,
		APIKey:      f.cfg.LLM.XAI.APIKey,
		BaseURL:     f.cfg.LLM.XAI.BaseURL,
		Model:       f.cfg.LLM.XAI.Model,
		MaxTokens:   defaultOpenAIMaxTokens,
		Temperature: defaultOpenAITemperature,
	})
}

// minimax routes through OpenAI-compat adapter with the Minimax BaseURL.
func (f *Factory) minimax() (Provider, error) {
	return NewOpenAICompatAdapter(ProviderMinimax, Config{
		Provider:    ProviderMinimax,
		APIKey:      f.cfg.LLM.Minimax.APIKey,
		BaseURL:     f.cfg.LLM.Minimax.BaseURL,
		Model:       f.cfg.LLM.Minimax.Model,
		MaxTokens:   defaultOpenAIMaxTokens,
		Temperature: defaultOpenAITemperature,
	})
}
