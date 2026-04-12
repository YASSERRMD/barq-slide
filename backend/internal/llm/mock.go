package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// MockProvider is a test double for the Provider interface.
// It records calls and returns canned responses.
type MockProvider struct {
	ProviderNameValue ProviderName

	TextResponses       []TextResponse
	StructuredResponses []StructuredResponse

	TextCalls       []GenerateTextRequest
	StructuredCalls []GenerateStructuredRequest

	TextErr       error
	StructuredErr error

	textIdx       int
	structuredIdx int
}

// Name returns the configured provider name (defaults to "mock").
func (m *MockProvider) Name() ProviderName {
	if m.ProviderNameValue == "" {
		return "mock"
	}
	return m.ProviderNameValue
}

// GenerateText returns the next canned TextResponse or an error.
func (m *MockProvider) GenerateText(_ context.Context, req GenerateTextRequest) (*TextResponse, error) {
	m.TextCalls = append(m.TextCalls, req)

	if m.TextErr != nil {
		return nil, m.TextErr
	}

	if m.textIdx >= len(m.TextResponses) {
		return nil, fmt.Errorf("mock: no more TextResponse stubs (call %d)", m.textIdx)
	}

	resp := m.TextResponses[m.textIdx]
	m.textIdx++
	return &resp, nil
}

// GenerateStructured returns the next canned StructuredResponse or an error.
func (m *MockProvider) GenerateStructured(_ context.Context, req GenerateStructuredRequest) (*StructuredResponse, error) {
	m.StructuredCalls = append(m.StructuredCalls, req)

	if m.StructuredErr != nil {
		return nil, m.StructuredErr
	}

	if m.structuredIdx >= len(m.StructuredResponses) {
		return nil, fmt.Errorf("mock: no more StructuredResponse stubs (call %d)", m.structuredIdx)
	}

	resp := m.StructuredResponses[m.structuredIdx]
	m.structuredIdx++
	return &resp, nil
}

// MustMarshal is a test helper that JSON-marshals v and panics on error.
func MustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("llm.MustMarshal: %v", err))
	}
	return b
}
