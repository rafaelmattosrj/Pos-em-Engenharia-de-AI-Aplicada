// Package handler implementa os endpoints HTTP — equivalente a
// app.controller.ts, event.controller.ts e speaker.controller.ts.
package handler

import (
	"encoding/json"
	"net/http"

	"cfp-platform/dto"
	"cfp-platform/service"
)

// Health trata GET /api — equivalente a AppController.getData.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Hello API"})
}

// EventHandler expõe os endpoints de eventos — equivalente a
// EventController.
type EventHandler struct {
	Service *service.EventService
}

// Create trata POST /api/events.
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d dto.CreateEventDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}
	if err := d.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.Service.Create(d))
}

// FindAll trata GET /api/events.
func (h *EventHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Service.FindAll())
}

// SpeakerHandler expõe os endpoints de palestrantes — equivalente a
// SpeakerController.
type SpeakerHandler struct {
	Service *service.SpeakerService
}

// Create trata POST /api/speakers.
func (h *SpeakerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d dto.CreateSpeakerDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}
	if err := d.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.Service.Create(d))
}

// FindAll trata GET /api/speakers.
func (h *SpeakerHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Service.FindAll())
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
