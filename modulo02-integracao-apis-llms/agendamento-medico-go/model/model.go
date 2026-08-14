// Package model define os tipos de domínio do agendamento médico —
// equivalente a Appointment.java, Professional.java e IntentResult.java.
package model

import "time"

// Professional é um profissional de saúde disponível para agendamento.
type Professional struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

// Appointment é uma consulta agendada.
type Appointment struct {
	Date           time.Time
	PatientName    string
	Reason         string
	ProfessionalID int
}

// IntentResult é o output estruturado do LLM ao identificar a intenção do
// paciente — equivalente ao IntentSchema (Zod/BeanOutputConverter).
type IntentResult struct {
	Intent           string  `json:"intent"` // "schedule" | "cancel" | "unknown"
	PatientName      *string `json:"patientName"`
	ProfessionalID   *int    `json:"professionalId"`
	ProfessionalName *string `json:"professionalName"`
	Datetime         *string `json:"datetime"` // ISO 8601
	Reason           *string `json:"reason"`
}

// UnknownIntent é o IntentResult retornado quando a identificação de
// intenção falha ou é inconclusiva.
func UnknownIntent() IntentResult {
	return IntentResult{Intent: "unknown"}
}
