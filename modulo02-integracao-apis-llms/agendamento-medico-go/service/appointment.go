// Package service implementa a lógica de negócio de agendamento —
// equivalente a AppointmentService.java, IntentService.java e
// MessageGeneratorService.java.
package service

import (
	"strings"
	"sync"
	"time"

	"agendamento-medico/model"
)

// Professionals é a lista fixa de profissionais disponíveis — equivalente a
// AppointmentService.PROFESSIONALS.
var Professionals = []model.Professional{
	{ID: 1, Name: "Dr. Alicio da Silva", Specialty: "Cardiologia"},
	{ID: 2, Name: "Dra. Ana Pereira", Specialty: "Dermatologia"},
	{ID: 3, Name: "Dra. Carol Gomes", Specialty: "Neurologia"},
}

// AppointmentService gerencia agendamentos em memória. Protegido por mutex
// porque, ao contrário do protótipo Java, um servidor Go atende requisições
// HTTP concorrentemente por padrão.
type AppointmentService struct {
	mu           sync.Mutex
	appointments []model.Appointment
}

// NewAppointmentService cria o serviço já populado com os mesmos 2
// agendamentos de exemplo da versão Java.
func NewAppointmentService() *AppointmentService {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	todayAt11 := now.Add(11 * time.Hour)
	tomorrowAt14 := todayAt11.Add(24*time.Hour - 11*time.Hour + 14*time.Hour)

	return &AppointmentService{
		appointments: []model.Appointment{
			{Date: todayAt11, PatientName: "Joao da Silva", Reason: "check-up regular", ProfessionalID: 1},
			{Date: tomorrowAt14, PatientName: "Luana Costa", Reason: "Erupção cutânea", ProfessionalID: 2},
		},
	}
}

// CheckAvailability retorna true se não houver agendamento conflitante para
// o profissional e horário informados.
func (s *AppointmentService) CheckAvailability(professionalID int, date time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.checkAvailabilityLocked(professionalID, date)
}

func (s *AppointmentService) checkAvailabilityLocked(professionalID int, date time.Time) bool {
	for _, a := range s.appointments {
		if a.ProfessionalID == professionalID && a.Date.Equal(date) {
			return false
		}
	}
	return true
}

// ErrSlotUnavailable é retornado quando o horário já está ocupado.
var ErrSlotUnavailable = appointmentError("Horário indisponível para este profissional")

// ErrAppointmentNotFound é retornado quando não há agendamento correspondente
// para cancelar.
var ErrAppointmentNotFound = appointmentError("Agendamento não encontrado para cancelamento")

type appointmentError string

func (e appointmentError) Error() string { return string(e) }

// BookAppointment cria um novo agendamento, se o horário estiver disponível.
func (s *AppointmentService) BookAppointment(professionalID int, date time.Time, patientName, reason string) (model.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.checkAvailabilityLocked(professionalID, date) {
		return model.Appointment{}, ErrSlotUnavailable
	}

	appointment := model.Appointment{Date: date, PatientName: patientName, Reason: reason, ProfessionalID: professionalID}
	s.appointments = append(s.appointments, appointment)
	return appointment, nil
}

// CancelAppointment remove o agendamento correspondente, se existir.
func (s *AppointmentService) CancelAppointment(professionalID int, patientName string, date time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.appointments {
		if a.ProfessionalID == professionalID && a.Date.Equal(date) && strings.EqualFold(a.PatientName, patientName) {
			s.appointments = append(s.appointments[:i], s.appointments[i+1:]...)
			return nil
		}
	}
	return ErrAppointmentNotFound
}

// GetProfessionals retorna a lista de profissionais disponíveis.
func (s *AppointmentService) GetProfessionals() []model.Professional {
	return Professionals
}
