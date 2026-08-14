// Package graph orquestra o fluxo de agendamento médico — equivalente a
// AppointmentOrchestrator.java (que por sua vez substitui o StateGraph do
// LangGraph). Fluxo: identifyIntent → schedule/cancel/mensagem → END.
package graph

import (
	"context"
	"fmt"
	"log"
	"time"

	"agendamento-medico/model"
	"agendamento-medico/service"
)

// Orchestrator conduz o pipeline de agendamento.
type Orchestrator struct {
	IntentService      *service.IntentService
	AppointmentService *service.AppointmentService
	MessageGenerator   *service.MessageGeneratorService
}

// Process identifica a intenção da mensagem do usuário e a roteia para o
// tratamento correspondente — equivalente a AppointmentOrchestrator.process.
func (o *Orchestrator) Process(ctx context.Context, userMessage string) (string, error) {
	intent := o.IntentService.IdentifyIntent(ctx, userMessage, o.AppointmentService.GetProfessionals())

	log.Printf("➡️ Intenção identificada: %s\n", intent.Intent)

	switch intent.Intent {
	case "schedule":
		return o.handleSchedule(ctx, intent)
	case "cancel":
		return o.handleCancel(ctx, intent)
	default:
		return o.MessageGenerator.GenerateUnknownIntentMessage(ctx, userMessage)
	}
}

func (o *Orchestrator) handleSchedule(ctx context.Context, intent model.IntentResult) (string, error) {
	message, err := func() (string, error) {
		date, err := parseDatetime(intent.Datetime)
		if err != nil {
			return "", err
		}

		if _, err := o.AppointmentService.BookAppointment(
			intPtrValue(intent.ProfessionalID), date, strPtrValue(intent.PatientName), strPtrValue(intent.Reason),
		); err != nil {
			return "", err
		}

		details := fmt.Sprintf("consulta com %s em %s", strPtrValue(intent.ProfessionalName), strPtrValue(intent.Datetime))
		return o.MessageGenerator.GenerateSuccessMessage(ctx, "agendamento", strPtrValue(intent.PatientName), details)
	}()

	if err != nil {
		return o.MessageGenerator.GenerateErrorMessage(ctx, err.Error())
	}
	return message, nil
}

func (o *Orchestrator) handleCancel(ctx context.Context, intent model.IntentResult) (string, error) {
	message, err := func() (string, error) {
		date, err := parseDatetime(intent.Datetime)
		if err != nil {
			return "", err
		}

		if err := o.AppointmentService.CancelAppointment(
			intPtrValue(intent.ProfessionalID), strPtrValue(intent.PatientName), date,
		); err != nil {
			return "", err
		}

		details := fmt.Sprintf("consulta com %s", strPtrValue(intent.ProfessionalName))
		return o.MessageGenerator.GenerateSuccessMessage(ctx, "cancelamento", strPtrValue(intent.PatientName), details)
	}()

	if err != nil {
		return o.MessageGenerator.GenerateErrorMessage(ctx, err.Error())
	}
	return message, nil
}

func parseDatetime(datetime *string) (time.Time, error) {
	if datetime == nil {
		return time.Time{}, fmt.Errorf("datetime ausente na intenção identificada")
	}
	return time.Parse(time.RFC3339, *datetime)
}

func intPtrValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func strPtrValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
