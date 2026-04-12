package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// LLMConfigStore holds the active LLM configuration set via the UI.
// It is safe for concurrent use.
type LLMConfigStore struct {
	mu       sync.RWMutex
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

var globalLLMConfig = &LLMConfigStore{}

// LLMConfigPayload is the JSON body for GET/POST /api/llm-config.
type LLMConfigPayload struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseUrl"`
}

// Get returns a snapshot of the current config.
func (s *LLMConfigStore) Get() LLMConfigPayload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return LLMConfigPayload{
		Provider: s.Provider,
		APIKey:   s.APIKey,
		Model:    s.Model,
		BaseURL:  s.BaseURL,
	}
}

// Set overwrites the stored config.
func (s *LLMConfigStore) Set(p LLMConfigPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Provider = p.Provider
	s.APIKey = p.APIKey
	s.Model = p.Model
	s.BaseURL = p.BaseURL
}

// HandleGetLLMConfig serves GET /api/llm-config.
func HandleGetLLMConfig(w http.ResponseWriter, r *http.Request) {
	payload := globalLLMConfig.Get()
	// Never return the actual key to the browser — just whether one is set.
	safe := LLMConfigPayload{
		Provider: payload.Provider,
		APIKey:   maskKey(payload.APIKey),
		Model:    payload.Model,
		BaseURL:  payload.BaseURL,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(safe)
}

// HandleSetLLMConfig serves POST /api/llm-config.
func HandleSetLLMConfig(w http.ResponseWriter, r *http.Request) {
	var p LLMConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	globalLLMConfig.Set(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func maskKey(k string) string {
	if len(k) <= 8 {
		return "***"
	}
	return k[:4] + "****"
}
