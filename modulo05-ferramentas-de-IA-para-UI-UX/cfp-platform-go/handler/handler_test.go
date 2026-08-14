package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cfp-platform/model"
	"cfp-platform/service"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["message"] != "Hello API" {
		t.Errorf("mensagem inesperada: %q", body["message"])
	}
}

func TestEventHandler_CreateAndFindAll(t *testing.T) {
	h := &EventHandler{Service: &service.EventService{}}

	body, _ := json.Marshal(map[string]any{"nome": "DevFest", "endereco": "Centro", "capacidade": 500, "data": "2026-08-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create: esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}
	var created model.Event
	json.NewDecoder(rec.Body).Decode(&created)
	if created.ID == "" || created.Nome != "DevFest" {
		t.Errorf("evento criado inesperado: %+v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	listRec := httptest.NewRecorder()
	h.FindAll(listRec, listReq)

	var events []model.Event
	json.NewDecoder(listRec.Body).Decode(&events)
	if len(events) != 1 {
		t.Fatalf("esperava 1 evento listado, obteve %d", len(events))
	}
}

func TestEventHandler_Create_MissingNomeReturns400(t *testing.T) {
	h := &EventHandler{Service: &service.EventService{}}

	body, _ := json.Marshal(map[string]any{"nome": "", "endereco": "Centro", "capacidade": 100, "data": "2026-08-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestSpeakerHandler_CreateAndFindAll(t *testing.T) {
	h := &SpeakerHandler{Service: &service.SpeakerService{}}

	body, _ := json.Marshal(map[string]any{"name": "Ana Souza", "email": "ana@example.com", "talkTitle": "IA na pratica", "isGDE": true})
	req := httptest.NewRequest(http.MethodPost, "/api/speakers", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create: esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}
	var created model.Speaker
	json.NewDecoder(rec.Body).Decode(&created)
	if created.ID == "" || created.Name != "Ana Souza" || !created.IsGDE {
		t.Errorf("palestrante criado inesperado: %+v", created)
	}
}

func TestSpeakerHandler_Create_InvalidEmailReturns400(t *testing.T) {
	h := &SpeakerHandler{Service: &service.SpeakerService{}}

	body, _ := json.Marshal(map[string]any{"name": "Bruno", "email": "nao-e-email", "talkTitle": "Talk", "isGDE": false})
	req := httptest.NewRequest(http.MethodPost, "/api/speakers", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}
