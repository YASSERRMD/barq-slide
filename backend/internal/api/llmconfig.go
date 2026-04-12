package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// LLMConfigStore holds the active LLM configuration set via the UI.
// It is persisted to a JSON file so it survives container restarts.
type LLMConfigStore struct {
	mu       sync.RWMutex
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
	filePath string
}

var globalLLMConfig = newLLMConfigStore("/app/data/llm-config.json")

// LLMConfigPayload is the JSON body for GET/POST /api/llm-config.
type LLMConfigPayload struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseUrl"`
}

func newLLMConfigStore(path string) *LLMConfigStore {
	s := &LLMConfigStore{filePath: path}
	s.load() // best-effort load on startup
	return s
}

// load reads the persisted config from disk (best-effort).
func (s *LLMConfigStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return // file doesn't exist yet — fine
	}
	var p LLMConfigPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return
	}
	s.Provider = p.Provider
	s.APIKey = p.APIKey
	s.Model = p.Model
	s.BaseURL = p.BaseURL
}

// save writes the current config to disk (best-effort).
func (s *LLMConfigStore) save() {
	p := LLMConfigPayload{
		Provider: s.Provider,
		APIKey:   s.APIKey,
		Model:    s.Model,
		BaseURL:  s.BaseURL,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(s.filePath), 0700)
	_ = os.WriteFile(s.filePath, data, 0600)
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

// Set overwrites the stored config and persists it.
func (s *LLMConfigStore) Set(p LLMConfigPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Provider = p.Provider
	s.APIKey = p.APIKey
	s.Model = p.Model
	s.BaseURL = p.BaseURL
	s.save()
}

// HandleGetLLMConfig serves GET /api/llm-config.
func HandleGetLLMConfig(w http.ResponseWriter, r *http.Request) {
	payload := globalLLMConfig.Get()
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
