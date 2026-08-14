package service

import (
	"testing"
	"time"
)

func TestBookAppointment_Success(t *testing.T) {
	s := NewAppointmentService()
	date := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)

	appt, err := s.BookAppointment(1, date, "Rafael Souza", "consulta de rotina")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if appt.PatientName != "Rafael Souza" {
		t.Errorf("nome do paciente inesperado: %q", appt.PatientName)
	}
	if s.CheckAvailability(1, date) {
		t.Error("horario deveria estar ocupado apos o agendamento")
	}
}

func TestBookAppointment_SlotUnavailable(t *testing.T) {
	s := NewAppointmentService()
	date := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)

	if _, err := s.BookAppointment(1, date, "Paciente A", "motivo"); err != nil {
		t.Fatalf("primeiro agendamento deveria ter sucesso: %v", err)
	}

	_, err := s.BookAppointment(1, date, "Paciente B", "outro motivo")
	if err != ErrSlotUnavailable {
		t.Fatalf("esperava ErrSlotUnavailable, obteve %v", err)
	}
}

func TestCancelAppointment_Success(t *testing.T) {
	s := NewAppointmentService()
	date := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)
	s.BookAppointment(1, date, "Rafael Souza", "consulta")

	if err := s.CancelAppointment(1, "rafael souza", date); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !s.CheckAvailability(1, date) {
		t.Error("horario deveria estar disponivel apos o cancelamento")
	}
}

func TestCancelAppointment_NotFound(t *testing.T) {
	s := NewAppointmentService()
	date := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)

	err := s.CancelAppointment(1, "ninguem", date)
	if err != ErrAppointmentNotFound {
		t.Fatalf("esperava ErrAppointmentNotFound, obteve %v", err)
	}
}
