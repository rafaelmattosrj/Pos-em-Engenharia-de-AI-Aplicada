package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"legacy-customer-api/service"
)

// CustomerHandler expõe os endpoints de clientes e o health check —
// equivalente a CustomerController.java.
type CustomerHandler struct {
	CustomerService *service.CustomerService
}

// Health trata GET /health.
func (h *CustomerHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "legacy-customer-api-go",
	})
}

// ListAll trata GET /customers.
func (h *CustomerHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.CustomerService.FindAll())
}

// GetByID trata GET /customers/{id}.
func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "UUID inválido: "+r.PathValue("id"))
		return
	}

	customer, ok := h.CustomerService.FindByID(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "Cliente não encontrado: "+id.String())
		return
	}
	writeJSON(w, http.StatusOK, customer)
}

// Create trata POST /customers.
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	name := strings.TrimSpace(body["name"])
	phone := strings.TrimSpace(body["phone"])

	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "Campo 'name' é obrigatório")
		return
	}
	if phone == "" {
		writeJSONError(w, http.StatusBadRequest, "Campo 'phone' é obrigatório")
		return
	}

	created := h.CustomerService.Create(name, phone)
	writeJSON(w, http.StatusCreated, created)
}

// Update trata PUT /customers/{id}.
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "UUID inválido: "+r.PathValue("id"))
		return
	}

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	updated, ok := h.CustomerService.Update(id, body["name"], body["phone"])
	if !ok {
		writeJSONError(w, http.StatusNotFound, "Cliente não encontrado: "+id.String())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete trata DELETE /customers/{id}.
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "UUID inválido: "+r.PathValue("id"))
		return
	}

	if !h.CustomerService.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "Cliente não encontrado: "+id.String())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
