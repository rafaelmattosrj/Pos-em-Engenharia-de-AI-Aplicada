// Package handler implementa o endpoint HTTP — equivalente à rota
// POST /api/brag de server.ts.
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"brag-bot/model"
	"brag-bot/service"
)

// BragHandler expõe o endpoint de geração de Brag Documents.
type BragHandler struct {
	Service *service.BragService
}

// Generate trata POST /api/brag.
func (h *BragHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req model.BragRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if strings.TrimSpace(req.Definition) == "" {
		writeError(w, http.StatusBadRequest, "Definition is required")
		return
	}

	document, err := h.Service.Generate(r.Context(), req.Definition)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate brag")
		return
	}

	writeJSON(w, http.StatusOK, document)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
