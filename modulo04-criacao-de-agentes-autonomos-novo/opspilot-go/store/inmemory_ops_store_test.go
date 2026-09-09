package store

import (
	"errors"
	"testing"

	"opspilot/domain"
)

func TestCreateThenResolveIncident(t *testing.T) {
	s := NewInMemoryOpsStore()
	incident := s.CreateIncident("Checkout fora do ar", "checkout-api", domain.SeverityCritical)

	if incident.Status != domain.IncidentOpen {
		t.Fatalf("expected OPEN, got %s", incident.Status)
	}
	open := domain.IncidentOpen
	if got := len(s.GetIncidents(&open)); got != 1 {
		t.Fatalf("expected 1 open incident, got %d", got)
	}

	summary := "Rollback aplicado"
	resolved, err := s.ResolveIncident(incident.ID, &summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Status != domain.IncidentResolved {
		t.Fatalf("expected RESOLVED, got %s", resolved.Status)
	}
	if got := len(s.GetIncidents(&open)); got != 0 {
		t.Fatalf("expected 0 open incidents, got %d", got)
	}
	resolvedStatus := domain.IncidentResolved
	if got := len(s.GetIncidents(&resolvedStatus)); got != 1 {
		t.Fatalf("expected 1 resolved incident, got %d", got)
	}
}

func TestResolvingUnknownIncidentErrors(t *testing.T) {
	s := NewInMemoryOpsStore()
	_, err := s.ResolveIncident("nao-existe", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var notFound *domain.IncidentNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected IncidentNotFoundError, got %T", err)
	}
}

func TestFiltersAlertsByStatus(t *testing.T) {
	s := NewInMemoryOpsStore()
	s.SeedAlert(domain.Alert{ID: "a1", Service: "svc", Description: "desc", Severity: domain.SeverityHigh, Status: domain.AlertFiring})
	s.SeedAlert(domain.Alert{ID: "a2", Service: "svc", Description: "desc", Severity: domain.SeverityLow, Status: domain.AlertResolved})

	firing := domain.AlertFiring
	if got := len(s.GetAlerts(&firing)); got != 1 {
		t.Fatalf("expected 1 firing alert, got %d", got)
	}
	if got := len(s.GetAlerts(nil)); got != 2 {
		t.Fatalf("expected 2 alerts with no filter, got %d", got)
	}
}

func TestUnknownRunbookErrors(t *testing.T) {
	s := NewInMemoryOpsStore()
	s.SeedRunbook(domain.Runbook{Service: "checkout-api", Content: "conteudo"})

	runbook, err := s.GetRunbook("checkout-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runbook.Content != "conteudo" {
		t.Fatalf("unexpected content: %s", runbook.Content)
	}

	_, err = s.GetRunbook("outro-servico")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFound *domain.RunbookNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected RunbookNotFoundError, got %T", err)
	}
}
