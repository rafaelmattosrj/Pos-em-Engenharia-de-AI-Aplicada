package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"legacy-customer-api/service"
)

type contextKey string

const roleContextKey contextKey = "role"

// Authenticate popula o contexto da requisição com a role do usuário, caso
// o token Bearer seja válido — equivalente a JwtFilter.java. Não rejeita a
// requisição por si só: a decisão de autorização fica a cargo de
// RequireRole, assim como o SecurityConfig faz depois do filtro em Java.
func Authenticate(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := bearerToken(r); ok {
				if role, valid := authService.ValidateToken(token); valid {
					ctx := context.WithValue(r.Context(), roleContextKey, role)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RoleFromContext retorna a role autenticada da requisição, se houver.
func RoleFromContext(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(roleContextKey).(string)
	return role, ok
}

// RequireRole envolve um handler exigindo que a requisição esteja
// autenticada com uma das roles permitidas — equivalente às regras
// .hasAnyRole/.hasRole do SecurityConfig.java. Retorna 401 se não
// autenticado, 403 se autenticado mas sem a role exigida.
func RequireRole(next http.HandlerFunc, allowedRoles ...string) http.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := RoleFromContext(r)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "Não autenticado")
			return
		}
		if !allowed[role] {
			writeJSONError(w, http.StatusForbidden, "Acesso negado")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
