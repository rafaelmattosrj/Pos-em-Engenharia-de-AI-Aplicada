// Package store contém o contrato de persistência (porte de OpsStore,
// domain/types.ts) e a implementação em memória usada por padrão neste porte
// (o original usa SQLite em produção via SqliteOpsStore -- fora de escopo,
// ver README).
package store

import "opspilot/domain"

// OpsStore é o porte da interface OpsStore (domain/types.ts). Um ponteiro nil
// em *domain.AlertStatus / *domain.IncidentStatus representa "sem filtro"
// (equivalente a status?: AlertStatus | undefined no original / null em Java).
type OpsStore interface {
	GetAlerts(status *domain.AlertStatus) []domain.Alert
	GetIncidents(status *domain.IncidentStatus) []domain.Incident
	CreateIncident(title, service string, severity domain.Severity) domain.Incident
	ResolveIncident(id string, summary *string) (domain.Incident, error)
	GetRunbook(service string) (domain.Runbook, error)
	SeedAlert(alert domain.Alert)
	SeedRunbook(runbook domain.Runbook)
}
