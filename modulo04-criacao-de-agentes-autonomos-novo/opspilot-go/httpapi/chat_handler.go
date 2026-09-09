// Package httpapi é o porte simplificado de http/server.ts -- só a rota
// POST /chat, roteando pro nome de estratégia pedido no corpo
// ({"message": "...", "strategy": "team"}). Sem CORS/streaming/auditoria de
// request (fora do escopo deste porte, ver README).
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"opspilot/domain"
)

// ChatHandler é o porte de ChatHttpHandler.java.
type ChatHandler struct {
	Strategies      map[string]domain.ReasoningStrategy
	DefaultStrategy string
}

func NewChatHandler(strategies map[string]domain.ReasoningStrategy, defaultStrategy string) *ChatHandler {
	return &ChatHandler{Strategies: strategies, DefaultStrategy: defaultStrategy}
}

var _ http.Handler = (*ChatHandler)(nil)

type chatRequestBody struct {
	Message  string `json:"message"`
	Strategy string `json:"strategy"`
}

type traceEventDTO struct {
	Type    string `json:"type"`
	Node    string `json:"node"`
	Content string `json:"content"`
	To      string `json:"to,omitempty"`
}

type metricsDTO struct {
	LLMCalls  int   `json:"llmCalls"`
	LatencyMs int64 `json:"latencyMs"`
}

type chatResponseBody struct {
	Answer   string          `json:"answer"`
	Strategy string          `json:"strategy"`
	Metrics  metricsDTO      `json:"metrics"`
	Trace    []traceEventDTO `json:"trace"`
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/chat" {
		writeJSON(w, http.StatusNotFound, errorBody("Route not found"))
		return
	}

	var body chatRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("Invalid JSON body"))
		return
	}

	if strings.TrimSpace(body.Message) == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("Validation failed: message is required"))
		return
	}

	strategyName := body.Strategy
	if strategyName == "" {
		strategyName = h.DefaultStrategy
	}
	strategy, ok := h.Strategies[strategyName]
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorBody(fmt.Sprintf("Unknown strategy %q", strategyName)))
		return
	}

	result, err := strategy.Run(domain.NewStrategyRunInput(body.Message))
	if err != nil {
		// Diferente do porte Java (que deixa uma exceção não tratada
		// estourar até o servidor), aqui o erro (ex.: ModelUnavailableError)
		// vira uma resposta 500 explícita -- idiomático em Go.
		writeJSON(w, http.StatusInternalServerError, errorBody(err.Error()))
		return
	}

	trace := make([]traceEventDTO, 0, len(result.Trace))
	for _, event := range result.Trace {
		trace = append(trace, traceEventDTO{
			Type:    string(event.Type),
			Node:    event.Node,
			Content: event.Content,
			To:      event.To,
		})
	}

	writeJSON(w, http.StatusOK, chatResponseBody{
		Answer:   result.Answer,
		Strategy: strategy.Name(),
		Metrics:  metricsDTO{LLMCalls: result.Metrics.LLMCalls, LatencyMs: result.Metrics.LatencyMs},
		Trace:    trace,
	})
}

func errorBody(message string) map[string]string {
	return map[string]string{"error": message}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
