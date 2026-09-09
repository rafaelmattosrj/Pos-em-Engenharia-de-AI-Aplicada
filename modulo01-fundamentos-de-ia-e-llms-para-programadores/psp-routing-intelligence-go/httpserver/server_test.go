package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"psp-routing-intelligence/domain"
	"psp-routing-intelligence/routing"
)

func fakeRoutingService() *routing.Service {
	return &routing.Service{
		Embed: func(ctx context.Context, text string) ([]float64, error) {
			return []float64{0.1, 0.2, 0.3}, nil
		},
		FindSimilar: func(ctx context.Context, embedding []float64, topK int) ([]domain.SimilarCase, error) {
			return []domain.SimilarCase{
				{PSP: domain.Adyen, Status: domain.Success, Amount: 1320.00, Method: domain.CreditCard, Similarity: 0.94},
			}, nil
		},
		Chat: func(ctx context.Context, prompt string) (string, error) {
			return `{"primary":"ADYEN","confidence":0.87,"reasoning":"Alta aprovacao no Adyen.","fallback":["BRASPAG"]}`, nil
		},
	}
}

func fakeSeedService(seeded *int) *routing.SeedService {
	return &routing.SeedService{
		Clear: func(ctx context.Context) error { return nil },
		Embed: func(ctx context.Context, text string) ([]float64, error) { return []float64{0.1}, nil },
		Store: func(ctx context.Context, tx domain.HistoricalTransaction, text string, embedding []float64) error {
			*seeded++
			return nil
		},
	}
}

func TestRecommend_RequisicaoValidaRetorna200(t *testing.T) {
	handler := New(fakeRoutingService(), fakeSeedService(new(int)))

	body := `{
		"amount": 1500.00,
		"method": "CREDIT_CARD",
		"brand": "VISA",
		"userRegion": "SP",
		"hour": 20,
		"merchantCategory": "STREAMING"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/routing/recommend", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var got domain.RoutingRecommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("resposta nao decodificou: %v", err)
	}
	if got.Primary != domain.Adyen {
		t.Errorf("esperava primary=ADYEN, obteve %s", got.Primary)
	}
	if len(got.SimilarCases) != 1 {
		t.Errorf("esperava 1 similarCase, obteve %d", len(got.SimilarCases))
	}
}

func TestRecommend_CorpoInvalidoRetorna400(t *testing.T) {
	handler := New(fakeRoutingService(), fakeSeedService(new(int)))

	req := httptest.NewRequest(http.MethodPost, "/api/routing/recommend", strings.NewReader("nao e json"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestRecommend_CamposObrigatoriosAusentesRetorna400ComFieldErrors(t *testing.T) {
	handler := New(fakeRoutingService(), fakeSeedService(new(int)))

	req := httptest.NewRequest(http.MethodPost, "/api/routing/recommend", strings.NewReader(`{"userRegion":"SP"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "VALIDATION_ERROR" {
		t.Errorf("esperava error=VALIDATION_ERROR, obteve %v", body["error"])
	}
	fields, ok := body["fields"].(map[string]any)
	if !ok || fields["amount"] == nil || fields["method"] == nil || fields["merchantCategory"] == nil {
		t.Errorf("esperava erros de amount, method e merchantCategory, obteve %v", body["fields"])
	}
}

func TestRecommend_HoraForaDoIntervaloRetorna400(t *testing.T) {
	handler := New(fakeRoutingService(), fakeSeedService(new(int)))

	body := `{"amount": 10, "method": "PIX", "userRegion": "SP", "hour": 25, "merchantCategory": "STREAMING"}`
	req := httptest.NewRequest(http.MethodPost, "/api/routing/recommend", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestRecommend_RespostaInvalidaDoLlmRetorna502(t *testing.T) {
	service := fakeRoutingService()
	service.Chat = func(ctx context.Context, prompt string) (string, error) {
		return "isso nao e json", nil
	}
	handler := New(service, fakeSeedService(new(int)))

	body := `{"amount": 10, "method": "PIX", "userRegion": "SP", "hour": 10, "merchantCategory": "STREAMING"}`
	req := httptest.NewRequest(http.MethodPost, "/api/routing/recommend", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("esperava 502, obteve %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSeed_Retorna200ComQuantidadeCarregada(t *testing.T) {
	seeded := new(int)
	handler := New(fakeRoutingService(), fakeSeedService(seeded))

	req := httptest.NewRequest(http.MethodPost, "/api/routing/seed", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	if int(body["seeded"].(float64)) != 50 {
		t.Errorf("esperava seeded=50, obteve %v", body["seeded"])
	}
	if *seeded != 50 {
		t.Errorf("esperava 50 chamadas de Store, obteve %d", *seeded)
	}
}

func TestRecommend_MetodoErradoRetorna405(t *testing.T) {
	handler := New(fakeRoutingService(), fakeSeedService(new(int)))

	req := httptest.NewRequest(http.MethodGet, "/api/routing/recommend", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, obteve %d", rec.Code)
	}
}
