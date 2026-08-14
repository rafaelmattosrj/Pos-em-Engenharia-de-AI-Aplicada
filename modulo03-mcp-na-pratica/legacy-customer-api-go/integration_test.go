package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"legacy-customer-api/service"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	authService := service.NewAuthService("test-secret-com-pelo-menos-32-caracteres", time.Hour)
	customerService := service.NewCustomerService()
	return buildServer(authService, customerService, "test-super-secret")
}

func login(t *testing.T, server http.Handler, username, password string) (string, int) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return "", rec.Code
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp["token"], rec.Code
}

// Cenário 1: Login com credenciais válidas retorna JWT.
func TestLogin_ValidCredentials_ReturnsJWT(t *testing.T) {
	server := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("esperava token nao vazio")
	}
	if resp["role"] != "ADMIN" {
		t.Errorf("esperava role=ADMIN, obteve %q", resp["role"])
	}
}

func TestLogin_InvalidCredentials_Returns401(t *testing.T) {
	server := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "senhaErrada"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obteve %d", rec.Code)
	}
}

// Cenário 2: Acesso sem token retorna 401.
func TestGetCustomers_WithoutToken_Returns401(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obteve %d", rec.Code)
	}
}

// Cenário 3: MEMBER não pode fazer POST /customers (403 Forbidden).
func TestCreateCustomer_AsMember_Returns403(t *testing.T) {
	server := newTestServer(t)

	token, loginStatus := login(t, server, "member", "pass456")
	if loginStatus != http.StatusOK {
		t.Fatalf("login como member falhou: status %d", loginStatus)
	}

	body, _ := json.Marshal(map[string]string{"name": "Teste MEMBER", "phone": "(00) 00000-0000"})
	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obteve %d", rec.Code)
	}
}

// Bônus: ADMIN com token válido pode listar clientes.
func TestListCustomers_AsAdmin_Succeeds(t *testing.T) {
	server := newTestServer(t)

	token, loginStatus := login(t, server, "admin", "password123")
	if loginStatus != http.StatusOK {
		t.Fatalf("login como admin falhou: status %d", loginStatus)
	}

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}

	var customers []map[string]any
	json.NewDecoder(rec.Body).Decode(&customers)
	if len(customers) < 5 {
		t.Errorf("esperava pelo menos 5 clientes, obteve %d", len(customers))
	}
}

func TestServiceToken_ValidSuperSecret_ReturnsToken(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/service-token", nil)
	req.Header.Set("X-Super-Secret", "test-super-secret")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
}

func TestServiceToken_InvalidSuperSecret_Returns403(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/service-token", nil)
	req.Header.Set("X-Super-Secret", "segredo-errado")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obteve %d", rec.Code)
	}
}

func TestHealth_Public(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
}

func TestCustomerCRUD_AsAdmin(t *testing.T) {
	server := newTestServer(t)
	token, _ := login(t, server, "admin", "password123")

	// Create
	body, _ := json.Marshal(map[string]string{"name": "Fulano", "phone": "123"})
	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: esperava 201, obteve %d", rec.Code)
	}
	var created map[string]any
	json.NewDecoder(rec.Body).Decode(&created)
	id := created["id"].(string)

	// Get by ID
	req = httptest.NewRequest(http.MethodGet, "/customers/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: esperava 200, obteve %d", rec.Code)
	}

	// Update
	updateBody, _ := json.Marshal(map[string]string{"name": "Fulano Atualizado"})
	req = httptest.NewRequest(http.MethodPut, "/customers/"+id, bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: esperava 200, obteve %d", rec.Code)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/customers/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: esperava 204, obteve %d", rec.Code)
	}

	// Get by ID again → 404
	req = httptest.NewRequest(http.MethodGet, "/customers/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get apos delete: esperava 404, obteve %d", rec.Code)
	}
}
