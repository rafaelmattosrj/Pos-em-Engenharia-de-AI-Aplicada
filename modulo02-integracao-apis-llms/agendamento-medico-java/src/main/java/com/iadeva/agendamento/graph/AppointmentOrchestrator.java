package com.iadeva.agendamento.graph;

import com.iadeva.agendamento.model.IntentResult;
import com.iadeva.agendamento.service.AppointmentService;
import com.iadeva.agendamento.service.IntentService;
import com.iadeva.agendamento.service.MessageGeneratorService;
import org.springframework.stereotype.Component;

import java.time.Instant;

// Orquestrador do fluxo — substitui o StateGraph do LangGraph
// Fluxo: identifyIntent → schedule/cancel/message → END
@Component
public class AppointmentOrchestrator {

    private final IntentService intentService;
    private final AppointmentService appointmentService;
    private final MessageGeneratorService messageGenerator;

    public AppointmentOrchestrator(
            IntentService intentService,
            AppointmentService appointmentService,
            MessageGeneratorService messageGenerator) {
        this.intentService = intentService;
        this.appointmentService = appointmentService;
        this.messageGenerator = messageGenerator;
    }

    public String process(String userMessage) {
        // Nó 1: identifica intenção
        IntentResult intent = intentService.identifyIntent(
                userMessage,
                appointmentService.getProfessionals()
        );

        System.out.printf("➡️ Intenção identificada: %s%n", intent.intent());

        // Aresta condicional baseada na intenção
        return switch (intent.intent()) {
            case "schedule" -> handleSchedule(intent);
            case "cancel" -> handleCancel(intent);
            default -> messageGenerator.generateUnknownIntentMessage(userMessage);
        };
    }

    private String handleSchedule(IntentResult intent) {
        try {
            Instant date = Instant.parse(intent.datetime());
            var appointment = appointmentService.bookAppointment(
                    intent.professionalId(),
                    date,
                    intent.patientName(),
                    intent.reason()
            );
            return messageGenerator.generateSuccessMessage(
                    "agendamento",
                    intent.patientName(),
                    "consulta com %s em %s".formatted(intent.professionalName(), intent.datetime())
            );
        } catch (Exception e) {
            return messageGenerator.generateErrorMessage(e.getMessage());
        }
    }

    private String handleCancel(IntentResult intent) {
        try {
            Instant date = Instant.parse(intent.datetime());
            appointmentService.cancelAppointment(intent.professionalId(), intent.patientName(), date);
            return messageGenerator.generateSuccessMessage(
                    "cancelamento",
                    intent.patientName(),
                    "consulta com %s".formatted(intent.professionalName())
            );
        } catch (Exception e) {
            return messageGenerator.generateErrorMessage(e.getMessage());
        }
    }
}
