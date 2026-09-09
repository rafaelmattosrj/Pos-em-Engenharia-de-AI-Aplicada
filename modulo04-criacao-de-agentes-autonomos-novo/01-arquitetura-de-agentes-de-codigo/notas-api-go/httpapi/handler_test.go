package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notas-api/httpapi"
	"notas-api/service"
	"notas-api/store"
)

func newTestServer() *httptest.Server {
	svc := service.New(store.NewInMemoryTaskStore())
	mux := http.NewServeMux()
	handler := httpapi.NewHandler(svc)
	mux.Handle("/tasks", handler)
	mux.Handle("/tasks/", handler)
	return httptest.NewServer(mux)
}

func TestCreateThenListThenComplete(t *testing.T) {
	server := newTestServer()
	defer server.Close()
	client := server.Client()

	created, err := client.Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":"Ler livro"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", created.StatusCode)
	}
	var task struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(created.Body).Decode(&task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != "open" {
		t.Errorf("expected status open, got %s", task.Status)
	}

	listed, err := client.Get(server.URL + "/tasks?status=open")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listed.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", listed.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodPatch, server.URL+"/tasks/"+task.ID+"/complete", nil)
	completed, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if completed.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", completed.StatusCode)
	}

	delReq, _ := http.NewRequest(http.MethodDelete, server.URL+"/tasks/"+task.ID, nil)
	removed, err := client.Do(delReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", removed.StatusCode)
	}
}

func TestBlankTitleReturns400(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, err := server.Client().Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":"   "}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCompletingUnknownIdReturns404(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPatch, server.URL+"/tasks/does-not-exist/complete", nil)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
