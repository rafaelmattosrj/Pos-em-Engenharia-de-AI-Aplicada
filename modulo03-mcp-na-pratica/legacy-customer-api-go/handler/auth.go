// Package handler implementa os endpoints HTTP da API legada — equivalente
// a AuthController.java e CustomerController.java.
package handler

import (
	"encoding/json"
	"net/http"

	"legacy-customer-api/model"
	"legacy-customer-api/service"
)

// AuthHandler expõe os endpoints de autenticação — equivalente a
// AuthController.java.
type AuthHandler struct {
	AuthService         *service.AuthService
	ExpectedSuperSecret string
}

// Login trata POST /auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	token, ok := h.AuthService.Login(req.Username, req.Password)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Credenciais inválidas")
		return
	}

	role, _ := h.AuthService.ValidateToken(token)
	if role == "" {
		role = "UNKNOWN"
	}

	writeJSON(w, http.StatusOK, model.AuthResponse{Token: token, Role: role})
}

// ServiceToken trata POST /auth/service-token.
func (h *AuthHandler) ServiceToken(w http.ResponseWriter, r *http.Request) {
	superSecret := r.Header.Get("X-Super-Secret")
	if superSecret == "" || superSecret != h.ExpectedSuperSecret {
		writeJSONError(w, http.StatusForbidden, "Header X-Super-Secret ausente ou inválido")
		return
	}

	token := h.AuthService.CreateServiceToken()
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
