package com.iadeva.agendamento.graph;

import com.iadeva.agendamento.model.IntentResult;
import com.iadeva.agendamento.service.AppointmentService;
import com.iadeva.agendamento.service.IntentService;
import com.iadeva.agendamento.service.MessageGeneratorService;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.Instant;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AppointmentOrchestratorTest {

    @Mock
    private IntentService intentService;

    @Mock
    private MessageGeneratorService messageGenerator;

    private final AppointmentService appointmentService = new AppointmentService();

    private AppointmentOrchestrator newOrchestrator() {
        return new AppointmentOrchestrator(intentService, appointmentService, messageGenerator);
    }

    @Test
    void process_agendamentoComSucesso() {
        when(intentService.identifyIntent(anyString(), anyList())).thenReturn(new IntentResult(
                "schedule", "Rafael Souza", 3, "Dra. Carol Gomes", "2030-01-01T10:00:00Z", "dor de cabeca"));
        when(messageGenerator.generateSuccessMessage(anyString(), anyString(), anyString()))
                .thenReturn("Sua consulta foi agendada com sucesso!");

        String reply = newOrchestrator().process("quero marcar consulta com a Dra Carol");

        assertThat(reply).isEqualTo("Sua consulta foi agendada com sucesso!");
        assertThat(appointmentService.checkAvailability(3, Instant.parse("2030-01-01T10:00:00Z"))).isFalse();
    }

    @Test
    void process_conflitoDeHorarioGeraMensagemDeErro() {
        // Mesmo profissional e horario de um agendamento ja existente causa conflito real
        Instant conflictDate = Instant.parse("2030-06-15T09:00:00Z");
        appointmentService.bookAppointment(1, conflictDate, "Outro Paciente", "motivo");

        when(intentService.identifyIntent(anyString(), anyList())).thenReturn(new IntentResult(
                "schedule", "Rafael Souza", 1, "Dr. Alicio da Silva", "2030-06-15T09:00:00Z", "check-up"));
        when(messageGenerator.generateErrorMessage(anyString()))
                .thenReturn("Desculpe, esse horario nao esta mais disponivel.");

        String reply = newOrchestrator().process("quero marcar consulta com o Dr Alicio");

        assertThat(reply).isEqualTo("Desculpe, esse horario nao esta mais disponivel.");
    }

    @Test
    void process_cancelamentoComSucesso() {
        Instant date = Instant.parse("2030-02-02T11:00:00Z");
        appointmentService.bookAppointment(2, date, "Luana Costa", "consulta");

        when(intentService.identifyIntent(anyString(), anyList())).thenReturn(new IntentResult(
                "cancel", "Luana Costa", 2, "Dra. Ana Pereira", "2030-02-02T11:00:00Z", null));
        when(messageGenerator.generateSuccessMessage(anyString(), anyString(), anyString()))
                .thenReturn("Seu cancelamento foi confirmado.");

        String reply = newOrchestrator().process("quero cancelar minha consulta");

        assertThat(reply).isEqualTo("Seu cancelamento foi confirmado.");
        assertThat(appointmentService.checkAvailability(2, date)).isTrue();
    }

    @Test
    void process_cancelamentoNaoEncontradoGeraMensagemDeErro() {
        when(intentService.identifyIntent(anyString(), anyList())).thenReturn(new IntentResult(
                "cancel", "Ninguem", 1, "Dr. Alicio da Silva", "2030-03-03T10:00:00Z", null));
        when(messageGenerator.generateErrorMessage(anyString()))
                .thenReturn("Nao encontramos esse agendamento.");

        String reply = newOrchestrator().process("quero cancelar uma consulta que nao existe");

        assertThat(reply).isEqualTo("Nao encontramos esse agendamento.");
    }

    @Test
    void process_intencaoDesconhecidaUsaMensagemDeFallback() {
        when(intentService.identifyIntent(anyString(), anyList())).thenReturn(
                new IntentResult("unknown", null, null, null, null, null));
        when(messageGenerator.generateUnknownIntentMessage("oi, tudo bem?"))
                .thenReturn("Ola! Em que posso ajudar?");

        String reply = newOrchestrator().process("oi, tudo bem?");

        assertThat(reply).isEqualTo("Ola! Em que posso ajudar?");
    }
}
