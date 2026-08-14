package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway-openrouter/openrouter"
)

func TestChatHandler_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"model":   "model-a",
			"choices": []map[string]any{{"message": map[string]string{"content": "resposta"}}},
		})
	}))
	defer upstream.Close()

	client := &ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: upstream.URL, HTTPClient: upstream.Client()},
		FallbackModels: []string{"model-a"},
		SystemPrompt:   "sys",
	}

	body, _ := json.Marshal(map[string]string{"question": "pergunta valida"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ChatHandler(client).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Content != "resposta" {
		t.Errorf("esperava content='resposta', obteve %q", resp.Content)
	}
}

func TestChatHandler_QuestionTooShort(t *testing.T) {
	client := &ResilientClient{FallbackModels: []string{"model-a"}}

	body, _ := json.Marshal(map[string]string{"question": "oi"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ChatHandler(client).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestChatHandler_MethodNotAllowed(t *testing.T) {
	client := &ResilientClient{FallbackModels: []string{"model-a"}}

	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	rec := httptest.NewRecorder()

	ChatHandler(client).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, obteve %d", rec.Code)
	}
}
