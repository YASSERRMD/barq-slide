// Package config provides koanf-based configuration loading for barq-slides.
package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// Config is the top-level application configuration.
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	LLM      LLMConfig      `koanf:"llm"`
	Unsplash UnsplashConfig `koanf:"unsplash"`
	Log      LogConfig      `koanf:"log"`
}

// ServerConfig configures the ConnectRPC HTTP server.
type ServerConfig struct {
	Host            string `koanf:"host"`
	Port            int    `koanf:"port"`
	CORSAllowOrigin string `koanf:"cors_allow_origin"`
	TLSCertFile     string `koanf:"tls_cert_file"`
	TLSKeyFile      string `koanf:"tls_key_file"`
}

// LLMConfig holds credentials and defaults for all LLM providers.
type LLMConfig struct {
	DefaultProvider string          `koanf:"default_provider"`
	Anthropic       AnthropicConfig `koanf:"anthropic"`
	Gemini          GeminiConfig    `koanf:"gemini"`
	OpenAI          OpenAIConfig    `koanf:"openai"`
	XAI             XAIConfig       `koanf:"xai"`
	Minimax         MinimaxConfig   `koanf:"minimax"`
}

// AnthropicConfig holds Anthropic API configuration.
type AnthropicConfig struct {
	APIKey string `koanf:"api_key"`
	Model  string `koanf:"model"`
}

// GeminiConfig holds Google Gemini API configuration.
type GeminiConfig struct {
	APIKey string `koanf:"api_key"`
	Model  string `koanf:"model"`
}

// OpenAIConfig holds OpenAI API configuration.
type OpenAIConfig struct {
	APIKey  string `koanf:"api_key"`
	BaseURL string `koanf:"base_url"`
	Model   string `koanf:"model"`
}

// XAIConfig holds xAI (Grok) configuration. Routes via OpenAI-compat adapter.
type XAIConfig struct {
	APIKey  string `koanf:"api_key"`
	BaseURL string `koanf:"base_url"`
	Model   string `koanf:"model"`
}

// MinimaxConfig holds Minimax configuration. Routes via OpenAI-compat adapter.
type MinimaxConfig struct {
	APIKey  string `koanf:"api_key"`
	BaseURL string `koanf:"base_url"`
	Model   string `koanf:"model"`
}

// UnsplashConfig holds Unsplash API configuration.
type UnsplashConfig struct {
	AccessKey string `koanf:"access_key"`
}

// LogConfig configures the structured logger.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"` // "json" | "text"
}

// Load builds a Config from environment variables (prefixed BARQ_).
// All keys are lowercased with __ as the nesting delimiter.
// Example: BARQ_SERVER__PORT=8080 maps to Config.Server.Port.
func Load() (*Config, error) {
	k := koanf.New(".")

	defaults(k)

	if err := k.Load(env.Provider("BARQ_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "BARQ_")),
			"__", ".",
		)
	}), nil); err != nil {
		return nil, fmt.Errorf("loading env config: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}

// defaults populates sensible defaults into the koanf instance.
func defaults(k *koanf.Koanf) {
	_ = k.Load(rawProvider(map[string]interface{}{
		"server.host":              "0.0.0.0",
		"server.port":              8080,
		"server.cors_allow_origin": "*",
		"llm.default_provider":     "anthropic",
		"llm.anthropic.model":      "claude-opus-4-6",
		"llm.gemini.model":         "gemini-2.0-flash",
		"llm.openai.model":         "gpt-4o",
		"llm.openai.base_url":      "https://api.openai.com/v1",
		"llm.xai.base_url":         "https://api.x.ai/v1",
		"llm.xai.model":            "grok-3",
		"llm.minimax.base_url":     "https://api.minimax.chat/v1",
		"llm.minimax.model":        "abab6.5s-chat",
		"log.level":                "info",
		"log.format":               "json",
	}), nil)
}

// rawProvider is a minimal koanf provider for static map defaults.
type rawProvider map[string]interface{}

func (r rawProvider) Load() (map[string]interface{}, error) { return r, nil }
func (r rawProvider) Watch(_ func(event interface{}, err error)) error {
	return nil
}
