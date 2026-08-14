package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"brag-bot/model"
	"brag-bot/service"
)

type fakeGeminiClient struct {
	response string
	err      error
}

func (f *fakeGeminiClient) GenerateJSON(ctx context.Context, prompt string, temperature float64) (string, error) {
	return f.response, f.err
}

func TestGenerate_MissingDefinitionReturns400(t *testing.T) {
	h := &BragHandler{Service: &service.BragService{Client: &fakeGeminiClient{}}}

	req := httptest.NewRequest(http.MethodPost, "/api/brag", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "Definition is required" {
		t.Errorf("mensagem de erro inesperada: %q", body["error"])
	}
}

func TestGenerate_ValidDefinitionReturnsBragDocument(t *testing.T) {
	geminiJSON := `{
		"title": "Reduziu latencia da API em 50%",
		"context": "API de pagamentos com timeouts frequentes",
		"actionTaken": "Refatorou o connection pool e adicionou cache",
		"businessImpact": "Reducao de 80% nos tickets de suporte",
		"metrics": ["50% reduction", "10ms latency"],
		"technologiesUsed": ["Go", "Redis"]
	}`
	h := &BragHandler{Service: &service.BragService{Client: &fakeGeminiClient{response: geminiJSON}}}

	req := httptest.NewRequest(http.MethodPost, "/api/brag", bytes.NewReader([]byte(`{"definition":"otimizei a api de pagamentos"}`)))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}
	var doc model.BragDocument
	json.NewDecoder(rec.Body).Decode(&doc)
	if doc.ID == "" {
		t.Error("esperava id gerado")
	}
	if doc.Title != "Reduziu latencia da API em 50%" {
		t.Errorf("title inesperado: %q", doc.Title)
	}
	if len(doc.Metrics) != 2 {
		t.Errorf("esperava 2 metrics, obteve %d", len(doc.Metrics))
	}
}

func TestGenerate_GeminiErrorReturns500(t *testing.T) {
	h := &BragHandler{Service: &service.BragService{Client: &fakeGeminiClient{err: errors.New("falha na api do gemini")}}}

	req := httptest.NewRequest(http.MethodPost, "/api/brag", bytes.NewReader([]byte(`{"definition":"algo qualquer"}`)))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperava 500, obteve %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "Failed to generate brag" {
		t.Errorf("mensagem de erro inesperada: %q", body["error"])
	}
}

func TestGenerate_InvalidJSONFromGeminiReturns500(t *testing.T) {
	h := &BragHandler{Service: &service.BragService{Client: &fakeGeminiClient{response: "isso nao e JSON"}}}

	req := httptest.NewRequest(http.MethodPost, "/api/brag", bytes.NewReader([]byte(`{"definition":"algo qualquer"}`)))
	rec := httptest.NewRecorder()

	h.Generate(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperava 500, obteve %d", rec.Code)
	}
}
