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
