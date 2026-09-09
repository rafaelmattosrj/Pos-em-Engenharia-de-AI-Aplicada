// Package httpserver expõe os endpoints HTTP descritos em IDEIA.md —
// equivalente a com.psprouting.controller (RoutingController +
// GlobalExceptionHandler) da versão Java. Como Go não tem um framework web
// "de fábrica" como o Spring MVC, os handlers usam net/http puro (o
// ServeMux com padrões de método, disponível desde Go 1.22), sem
// dependência externa — projeto tem só 2 rotas, não há necessidade de
// chi/gin (ver Passo 2 da skill de portabilidade).
package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"psp-routing-intelligence/domain"
	"psp-routing-intelligence/routing"
)

// New monta o mux com as duas rotas de IDEIA.md, envolvendo com CORS
// liberado em /api/** (equivalente a CorsConfig da versao Java) — o
// frontend React descrito em IDEIA.md nao foi portado (ver README), mas o
// CORS permanece configurado para qualquer cliente local.
func New(routingService *routing.Service, seedService *routing.SeedService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/routing/recommend", handleRecommend(routingService))
	mux.HandleFunc("POST /api/routing/seed", handleSeed(seedService))
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleRecommend(service *routing.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req recommendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error":  "VALIDATION_ERROR",
				"fields": map[string]string{"body": "JSON invalido"},
			})
			return
		}

		if fieldErrors := req.validate(); len(fieldErrors) > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error":  "VALIDATION_ERROR",
				"fields": fieldErrors,
			})
			return
		}

		recommendation, err := service.Recommend(r.Context(), req.toDomain())
		if err != nil {
			var invalidResponse *routing.InvalidLLMResponseError
			if errors.As(err, &invalidResponse) {
				writeJSON(w, http.StatusBadGateway, map[string]any{
					"error":   "INVALID_LLM_RESPONSE",
					"message": err.Error(),
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "INTERNAL_ERROR",
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, recommendation)
	}
}

func handleSeed(service *routing.SeedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seeded, err := service.Seed(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "INTERNAL_ERROR",
				"message": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"seeded": seeded})
	}
}

// recommendRequest e o corpo de POST /api/routing/recommend, no mesmo shape
// do exemplo em IDEIA.md — equivalente a dto.RecommendRequest da versão
// Java (validação manual em vez de Bean Validation).
type recommendRequest struct {
	Amount           *float64 `json:"amount"`
	Method           string   `json:"method"`
	Brand            string   `json:"brand"`
	UserRegion       string   `json:"userRegion"`
	Hour             *int     `json:"hour"`
	MerchantCategory string   `json:"merchantCategory"`
}

func (r recommendRequest) validate() map[string]string {
	fields := map[string]string{}

	if r.Amount == nil || *r.Amount <= 0 {
		fields["amount"] = "deve ser um valor positivo"
	}
	if _, err := domain.ParsePaymentMethod(r.Method); err != nil {
		fields["method"] = "deve ser um de: PIX, CREDIT_CARD, WALLET"
	}
	if strings.TrimSpace(r.UserRegion) == "" {
		fields["userRegion"] = "obrigatorio"
	}
	if r.Hour == nil || *r.Hour < 0 || *r.Hour > 23 {
		fields["hour"] = "deve estar entre 0 e 23"
	}
	if strings.TrimSpace(r.MerchantCategory) == "" {
		fields["merchantCategory"] = "obrigatorio"
	}

	return fields
}

func (r recommendRequest) toDomain() domain.Transaction {
	method, _ := domain.ParsePaymentMethod(r.Method)
	amount := 0.0
	if r.Amount != nil {
		amount = *r.Amount
	}
	hour := 0
	if r.Hour != nil {
		hour = *r.Hour
	}

	return domain.Transaction{
		Amount:           amount,
		Method:           method,
		Brand:            r.Brand,
		UserRegion:       r.UserRegion,
		Hour:             hour,
		MerchantCategory: r.MerchantCategory,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
