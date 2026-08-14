package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req generateRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.GenerationConfig.ResponseMIMEType != "application/json" {
			t.Errorf("esperava responseMimeType=application/json, obteve %q", req.GenerationConfig.ResponseMIMEType)
		}

		json.NewEncoder(w).Encode(generateResponse{
			Candidates: []candidate{
				{Content: content{Parts: []part{{Text: `{"title":"ok"}`}}}},
			},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gemini-2.5-flash", BaseURL: server.URL, HTTPClient: server.Client()}
	text, err := client.GenerateJSON(context.Background(), "prompt", 0.8)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if text != `{"title":"ok"}` {
		t.Errorf("texto inesperado: %q", text)
	}
}

func TestGenerateJSON_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(generateResponse{Error: &apiError{Message: "erro interno"}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gemini-2.5-flash", BaseURL: server.URL, HTTPClient: server.Client()}
	_, err := client.GenerateJSON(context.Background(), "prompt", 0.8)
	if err == nil {
		t.Fatal("esperava erro para status 500")
	}
}
