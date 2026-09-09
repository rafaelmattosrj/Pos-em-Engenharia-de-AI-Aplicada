package domain

import "fmt"

// IncidentNotFoundError é o porte de Exceptions.IncidentNotFoundException
// (domain/errors.ts / Exceptions.java).
type IncidentNotFoundError struct {
	ID string
}

func (e *IncidentNotFoundError) Error() string {
	return fmt.Sprintf("Incident %q was not found", e.ID)
}

// RunbookNotFoundError é o porte de Exceptions.RunbookNotFoundException.
type RunbookNotFoundError struct {
	Service string
}

func (e *RunbookNotFoundError) Error() string {
	return fmt.Sprintf("Runbook for service %q was not found", e.Service)
}

// ModelUnavailableError é o porte de Exceptions.ModelUnavailableException
// (agents/model.ts: composeResilientRunnable).
type ModelUnavailableError struct {
	Message string
}

func (e *ModelUnavailableError) Error() string {
	return e.Message
}
