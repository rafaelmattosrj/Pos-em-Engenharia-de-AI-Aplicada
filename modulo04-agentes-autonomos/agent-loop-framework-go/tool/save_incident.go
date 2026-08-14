package tool

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SaveIncidentTool persiste incidentes em memória durante a execução da
// aplicação — equivalente a SaveIncidentTool.java. Em produção persistiria
// no banco de dados de incidentes (PagerDuty, Jira, ServiceNow).
type SaveIncidentTool struct {
	mu    sync.Mutex
	store []map[string]any
}

func (*SaveIncidentTool) Name() string { return "saveIncident" }

func (t *SaveIncidentTool) Execute(args map[string]any) string {
	id := "INC-" + strings.ToUpper(uuid.New().String()[:8])
	title := stringArg(args, "title", "Incidente sem título")
	severity := stringArg(args, "severity", "medium")
	description := stringArg(args, "description", "Sem descrição")

	incident := map[string]any{
		"id":          id,
		"title":       title,
		"severity":    severity,
		"description": description,
		"status":      "OPEN",
		"created_at":  time.Now().Format(time.RFC3339),
		"created_by":  "DevOps Agent",
	}

	t.mu.Lock()
	t.store = append(t.store, incident)
	t.mu.Unlock()

	log.Printf("[SaveIncidentTool] Incidente criado: %s — %s", id, title)

	out, err := json.Marshal(map[string]any{
		"success":     true,
		"incident_id": id,
		"message":     "Incidente registrado com sucesso",
		"incident":    incident,
	})
	if err != nil {
		return `{"success": false, "error": "Falha ao serializar confirmação"}`
	}
	return string(out)
}

// AllIncidents retorna todos os incidentes salvos em memória (útil para
// testes e auditoria).
func (t *SaveIncidentTool) AllIncidents() []map[string]any {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]map[string]any, len(t.store))
	copy(result, t.store)
	return result
}
