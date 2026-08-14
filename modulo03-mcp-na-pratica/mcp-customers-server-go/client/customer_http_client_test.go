package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListAll_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header inesperado: %q", r.Header.Get("Authorization"))
		}
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "1", "name": "Alice", "phone": "111"},
		})
	}))
	defer server.Close()

	client := NewCustomerHttpClient(server.URL, "test-token")
	customers, err := client.ListAll(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(customers) != 1 || customers[0].Name != "Alice" {
		t.Errorf("resultado inesperado: %+v", customers)
	}
}

func TestCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "new-id", "name": "Novo", "phone": "999"})
	}))
	defer server.Close()

	client := NewCustomerHttpClient(server.URL, "test-token")
	result := client.Create(context.Background(), "Novo", "999")
	if !result.Success {
		t.Fatalf("esperava sucesso, obteve %+v", result)
	}
	if result.Customer == nil || result.Customer.ID != "new-id" {
		t.Errorf("customer inesperado: %+v", result.Customer)
	}
}

func TestUpdate_NoFieldsProvided(t *testing.T) {
	client := NewCustomerHttpClient("http://unused", "test-token")
	result := client.Update(context.Background(), "1", "", "")
	if result.Success {
		t.Error("esperava falha quando nenhum campo e fornecido")
	}
	if result.Message != "Nenhum campo fornecido para atualização" {
		t.Errorf("mensagem inesperada: %q", result.Message)
	}
}

func TestDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewCustomerHttpClient(server.URL, "test-token")
	result := client.Delete(context.Background(), "1")
	if !result.Success {
		t.Fatalf("esperava sucesso, obteve %+v", result)
	}
}

func TestDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"nao encontrado"}`))
	}))
	defer server.Close()

	client := NewCustomerHttpClient(server.URL, "test-token")
	result := client.Delete(context.Background(), "id-inexistente")
	if result.Success {
		t.Error("esperava falha para status 404")
	}
}
