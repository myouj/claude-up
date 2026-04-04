package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ----- OpenAIProvider Tests -----

func TestOpenAIProvider_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		resp := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "OpenAI response text",
					},
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 42,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider with test server URL
	provider := &OpenAIProvider{baseURL: server.URL + "/chat/completions"}

	messages := []map[string]string{
		{"role": "user", "content": "Hello"},
	}
	response, tokens, err := provider.Call(messages, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "OpenAI response text" {
		t.Errorf("response: got %q, want %q", response, "OpenAI response text")
	}
	if tokens != 42 {
		t.Errorf("tokens: got %d, want %d", tokens, 42)
	}
}

func TestOpenAIProvider_Call_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	provider := &OpenAIProvider{baseURL: server.URL}

	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gpt-4o")
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestOpenAIProvider_Call_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json{{{"))
	}))
	defer server.Close()

	provider := &OpenAIProvider{baseURL: server.URL}

	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gpt-4o")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestOpenAIProvider_Call_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"usage": map[string]interface{}{"total_tokens": 10},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OpenAIProvider{baseURL: server.URL}

	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gpt-4o")
	if err == nil {
		t.Fatal("expected error for unexpected response format")
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := &OpenAIProvider{}
	if p.Name() != "openai" {
		t.Errorf("Name(): got %s, want openai", p.Name())
	}
}

func TestOpenAIProvider_BaseURL(t *testing.T) {
	p := &OpenAIProvider{baseURL: "https://custom.openai.com/v1"}
	if p.BaseURL() != "https://custom.openai.com/v1" {
		t.Errorf("BaseURL(): got %s", p.BaseURL())
	}
}

// ----- ClaudeProvider Tests -----

func TestClaudeProvider_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version header, got %s", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("anthropic-dangerous-direct-browser-access") != "true" {
			t.Errorf("expected dangerous-browser header")
		}

		resp := map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Claude response text",
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  10,
				"output_tokens": 5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &ClaudeProvider{baseURL: server.URL}
	messages := []map[string]string{
		{"role": "user", "content": "Hello"},
	}
	response, tokens, err := provider.Call(messages, "claude-3-5-sonnet-20241022")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "Claude response text" {
		t.Errorf("response: got %q, want %q", response, "Claude response text")
	}
	if tokens != 15 {
		t.Errorf("tokens: got %d, want %d", tokens, 15)
	}
}

func TestClaudeProvider_Call_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := &ClaudeProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "claude-3-5-sonnet-20241022")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClaudeProvider_Call_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{broken"))
	}))
	defer server.Close()

	provider := &ClaudeProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "claude")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClaudeProvider_Call_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"usage": map[string]interface{}{"input_tokens": 10},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &ClaudeProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "claude")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClaudeProvider_Name(t *testing.T) {
	p := &ClaudeProvider{}
	if p.Name() != "claude" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

// ----- GeminiProvider Tests -----

func TestGeminiProvider_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"text": "Gemini response text",
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &GeminiProvider{baseURL: server.URL + "/v1beta/models/test:generateContent"}
	messages := []map[string]string{
		{"role": "user", "content": "Hello"},
	}
	response, tokens, err := provider.Call(messages, "gemini-2.0-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "Gemini response text" {
		t.Errorf("response: got %q, want %q", response, "Gemini response text")
	}
	if tokens != 0 {
		t.Errorf("tokens: got %d, want 0", tokens)
	}
}

func TestGeminiProvider_Call_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	provider := &GeminiProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gemini")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGeminiProvider_Call_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := &GeminiProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gemini")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGeminiProvider_Call_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"foo": "bar",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &GeminiProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "gemini")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGeminiProvider_Name(t *testing.T) {
	p := &GeminiProvider{}
	if p.Name() != "gemini" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

func TestGeminiProvider_Call_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{"text": "ok"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &GeminiProvider{baseURL: server.URL + "/v1beta/models/gemini-2.0-flash:generateContent"}
	// Empty model triggers default
	_, _, err := provider.Call([]map[string]string{{"role": "user", "content": "test"}}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ----- MiniMaxProvider Tests -----

func TestMiniMaxProvider_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		resp := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"messages": []interface{}{
						map[string]interface{}{
							"text": "MiniMax response text",
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &MiniMaxProvider{baseURL: server.URL + "/v1/text/chatcompletion_v2"}
	messages := []map[string]string{
		{"role": "USER", "text": "Hello"},
	}
	response, tokens, err := provider.Call(messages, "MiniMax-M2.7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "MiniMax response text" {
		t.Errorf("response: got %q, want %q", response, "MiniMax response text")
	}
	if tokens != 0 {
		t.Errorf("tokens: got %d, want 0", tokens)
	}
}

func TestMiniMaxProvider_Call_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := &MiniMaxProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "USER", "text": "test"}}, "MiniMax-M2.7")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMiniMaxProvider_Call_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid"))
	}))
	defer server.Close()

	provider := &MiniMaxProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "USER", "text": "test"}}, "MiniMax-M2.7")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMiniMaxProvider_Call_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &MiniMaxProvider{baseURL: server.URL}
	_, _, err := provider.Call([]map[string]string{{"role": "USER", "text": "test"}}, "MiniMax-M2.7")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMiniMaxProvider_Name(t *testing.T) {
	p := &MiniMaxProvider{}
	if p.Name() != "minimax" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

// ----- Constructor Tests -----

func TestNewOpenAIProvider(t *testing.T) {
	p := NewOpenAIProvider()
	if p.Name() != "openai" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

func TestNewClaudeProvider(t *testing.T) {
	p := NewClaudeProvider()
	if p.Name() != "claude" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

func TestNewGeminiProvider(t *testing.T) {
	p := NewGeminiProvider()
	if p.Name() != "gemini" {
		t.Errorf("Name(): got %s", p.Name())
	}
}

func TestNewMiniMaxProvider(t *testing.T) {
	p := NewMiniMaxProvider()
	if p.Name() != "minimax" {
		t.Errorf("Name(): got %s", p.Name())
	}
}
