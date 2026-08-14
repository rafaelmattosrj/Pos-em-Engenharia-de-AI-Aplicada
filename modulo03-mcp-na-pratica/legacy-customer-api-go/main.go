// API legada de clientes com autenticação JWT, service tokens, RBAC e rate
// limiting — porte Go do projeto Spring Boot/Spring Security
// legacy-customer-api-java.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"legacy-customer-api/handler"
	"legacy-customer-api/middleware"
	"legacy-customer-api/service"
)

// buildServer monta o roteador completo (rotas + middlewares) — equivalente
// a SecurityConfig.filterChain + o registro dos @RestController.
func buildServer(authService *service.AuthService, customerService *service.CustomerService, superSecret string) http.Handler {
	authHandler := &handler.AuthHandler{AuthService: authService, ExpectedSuperSecret: superSecret}
	customerHandler := &handler.CustomerHandler{CustomerService: customerService}

	mux := http.NewServeMux()

	// Público — equivalente a .requestMatchers("/auth/**", "/health").permitAll()
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/service-token", authHandler.ServiceToken)
	mux.HandleFunc("GET /health", customerHandler.Health)

	// Leitura: MEMBER ou ADMIN — equivalente a .hasAnyRole("MEMBER", "ADMIN")
	mux.HandleFunc("GET /customers", middleware.RequireRole(customerHandler.ListAll, "MEMBER", "ADMIN"))
	mux.HandleFunc("GET /customers/{id}", middleware.RequireRole(customerHandler.GetByID, "MEMBER", "ADMIN"))

	// Escrita: somente ADMIN — equivalente a .hasRole("ADMIN")
	mux.HandleFunc("POST /customers", middleware.RequireRole(customerHandler.Create, "ADMIN"))
	mux.HandleFunc("PUT /customers/{id}", middleware.RequireRole(customerHandler.Update, "ADMIN"))
	mux.HandleFunc("DELETE /customers/{id}", middleware.RequireRole(customerHandler.Delete, "ADMIN"))

	// Cadeia de filtros: rate limit → autenticação (popula role) → rotas.
	// Mesma ordem de SecurityConfig.filterChain (addFilterBefore rateLimitFilter, jwtFilter).
	rateLimiter := middleware.NewRateLimiter()
	var h http.Handler = mux
	h = middleware.Authenticate(authService)(h)
	h = rateLimiter.Middleware(h)
	return h
}

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	jwtSecret := getEnv("JWT_SECRET", "minha-chave-secreta-muito-longa-para-hs256-pelo-menos-32-chars")
	jwtExpiration := time.Duration(getEnvInt("JWT_EXPIRATION_MS", 86400000)) * time.Millisecond
	superSecret := getEnv("SUPER_SECRET", "superSecret123")

	authService := service.NewAuthService(jwtSecret, jwtExpiration)
	customerService := service.NewCustomerService()

	h := buildServer(authService, customerService, superSecret)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("legacy-customer-api ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}
