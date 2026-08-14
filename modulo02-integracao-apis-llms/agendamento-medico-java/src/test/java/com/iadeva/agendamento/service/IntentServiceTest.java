package com.iadeva.agendamento.service;

import com.iadeva.agendamento.ResilientChatClient;
import com.iadeva.agendamento.model.IntentResult;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class IntentServiceTest {

    @Mock
    private ResilientChatClient fallbackClient;

    private final AppointmentService appointmentService = new AppointmentService();

    private IntentService newService() {
        return new IntentService(fallbackClient);
    }

    @Test
    void identifyIntent_jsonPuro() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn(
                "{\"intent\":\"schedule\",\"patientName\":\"Rafael\",\"professionalId\":1,"
                        + "\"professionalName\":\"Dr. Alicio\",\"datetime\":\"2030-01-01T10:00:00Z\",\"reason\":\"check-up\"}");

        IntentResult result = newService().identifyIntent("quero agendar", appointmentService.getProfessionals());

        assertThat(result.intent()).isEqualTo("schedule");
        assertThat(result.patientName()).isEqualTo("Rafael");
        assertThat(result.professionalId()).isEqualTo(1);
    }

    @Test
    void identifyIntent_jsonCercadoEmMarkdown() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn(
                "```json\n{\"intent\":\"cancel\",\"patientName\":null,\"professionalId\":null,"
                        + "\"professionalName\":null,\"datetime\":null,\"reason\":null}\n```");

        IntentResult result = newService().identifyIntent("quero cancelar", appointmentService.getProfessionals());

        assertThat(result.intent()).isEqualTo("cancel");
    }

    @Test
    void identifyIntent_respostaInvalidaCaiParaUnknown() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn("isso nao e JSON");

        IntentResult result = newService().identifyIntent("mensagem qualquer", appointmentService.getProfessionals());

        assertThat(result.intent()).isEqualTo("unknown");
    }
}
