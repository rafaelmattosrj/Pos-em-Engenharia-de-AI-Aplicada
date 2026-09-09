package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"opspilot/domain"
)

// InMemoryOpsStore é o porte de InMemoryStore (store/in-memory-store.ts) --
// o próprio original usa essa implementação em memória em testes; aqui ela é
// a implementação default (ver README). Protegida por mutex porque, ao
// contrário do event-loop single-threaded do Node, o net/http do Go atende
// requisições concorrentemente por padrão.
type InMemoryOpsStore struct {
	mu        sync.Mutex
	alerts    []domain.Alert
	incidents []domain.Incident
	runbooks  map[string]domain.Runbook
}

// NewInMemoryOpsStore cria um InMemoryOpsStore vazio, pronto para uso.
func NewInMemoryOpsStore() *InMemoryOpsStore {
	return &InMemoryOpsStore{
		runbooks: make(map[string]domain.Runbook),
	}
}

var _ OpsStore = (*InMemoryOpsStore)(nil)

func (s *InMemoryOpsStore) GetAlerts(status *domain.AlertStatus) []domain.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]domain.Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		if status == nil || alert.Status == *status {
			result = append(result, alert)
		}
	}
	return result
}

func (s *InMemoryOpsStore) GetIncidents(status *domain.IncidentStatus) []domain.Incident {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]domain.Incident, 0, len(s.incidents))
	for _, incident := range s.incidents {
		if status == nil || incident.Status == *status {
			result = append(result, incident)
		}
	}
	// Já inserido em ordem de criação (createIncident faz append), igual ao
	// .sort por createdAt do original -- sem necessidade de reordenar.
	return result
}

func (s *InMemoryOpsStore) CreateIncident(title, service string, severity domain.Severity) domain.Incident {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident := domain.Incident{
		ID:        fmt.Sprintf("inc-%d-%s", time.Now().UnixMilli(), randomHex(2)),
		Title:     title,
		Service:   service,
		Severity:  severity,
		Status:    domain.IncidentOpen,
		CreatedAt: time.Now().UnixMilli(),
	}
	s.incidents = append(s.incidents, incident)
	return incident
}

func (s *InMemoryOpsStore) ResolveIncident(id string, summary *string) (domain.Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, incident := range s.incidents {
		if incident.ID == id {
			summaryValue := ""
			if summary != nil {
				summaryValue = *summary
			}
			resolved := incident.Resolved(time.Now().UnixMilli(), summaryValue)
			s.incidents[i] = resolved
			return resolved, nil
		}
	}
	return domain.Incident{}, &domain.IncidentNotFoundError{ID: id}
}

func (s *InMemoryOpsStore) GetRunbook(service string) (domain.Runbook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	runbook, ok := s.runbooks[service]
	if !ok {
		return domain.Runbook{}, &domain.RunbookNotFoundError{Service: service}
	}
	return runbook, nil
}

func (s *InMemoryOpsStore) SeedAlert(alert domain.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, alert)
}

func (s *InMemoryOpsStore) SeedRunbook(runbook domain.Runbook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runbooks[runbook.Service] = runbook
}

// randomHex é o equivalente a UUID.randomUUID().toString().substring(0, 4) do
// porte Java: só precisa de baixa colisão para um id legível, não de um UUID
// completo.
func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
