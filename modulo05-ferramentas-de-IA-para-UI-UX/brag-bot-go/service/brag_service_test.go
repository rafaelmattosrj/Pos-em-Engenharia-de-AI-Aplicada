package service

import (
	"context"
	"testing"
)

type stubGeminiClient struct {
	response string
	err      error
}

func (s *stubGeminiClient) GenerateJSON(ctx context.Context, prompt string, temperature float64) (string, error) {
	return s.response, s.err
}

func TestGenerate_ParsesGeminiResponse(t *testing.T) {
	client := &stubGeminiClient{response: `{
		"title": "Reduziu latencia da API em 50%",
		"context": "API de pagamentos com timeouts frequentes",
		"actionTaken": "Refatorou o connection pool e adicionou cache",
		"businessImpact": "Reducao de 80% nos tickets de suporte",
		"metrics": ["50% reduction", "10ms latency"],
		"technologiesUsed": ["Go", "Redis"]
	}`}
	svc := &BragService{Client: client}

	doc, err := svc.Generate(context.Background(), "otimizei a api de pagamentos")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if doc.ID == "" {
		t.Error("esperava id gerado")
	}
	if doc.Title != "Reduziu latencia da API em 50%" {
		t.Errorf("title inesperado: %q", doc.Title)
	}
	if len(doc.Metrics) != 2 {
		t.Errorf("esperava 2 metrics, obteve %d", len(doc.Metrics))
	}
	if len(doc.TechnologiesUsed) != 2 {
		t.Errorf("esperava 2 technologiesUsed, obteve %d", len(doc.TechnologiesUsed))
	}
}

func TestGenerate_PropagatesGeminiError(t *testing.T) {
	client := &stubGeminiClient{err: errFake("falha na api do gemini")}
	svc := &BragService{Client: client}

	_, err := svc.Generate(context.Background(), "algo qualquer")
	if err == nil {
		t.Fatal("esperava erro propagado do cliente Gemini")
	}
}

func TestGenerate_InvalidJSONReturnsError(t *testing.T) {
	client := &stubGeminiClient{response: "isso nao e JSON"}
	svc := &BragService{Client: client}

	_, err := svc.Generate(context.Background(), "algo qualquer")
	if err == nil {
		t.Fatal("esperava erro para resposta que nao e JSON valido")
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }
