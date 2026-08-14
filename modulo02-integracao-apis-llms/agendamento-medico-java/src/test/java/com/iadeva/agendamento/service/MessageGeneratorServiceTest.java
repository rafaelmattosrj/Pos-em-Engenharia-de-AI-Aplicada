package com.iadeva.agendamento.service;

import com.iadeva.agendamento.ResilientChatClient;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class MessageGeneratorServiceTest {

    @Mock
    private ResilientChatClient fallbackClient;

    private MessageGeneratorService newService() {
        return new MessageGeneratorService(fallbackClient);
    }

    @Test
    void generateSuccessMessage_repassaDetalhesParaOLlmERetornaResposta() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn("Sua consulta foi agendada com sucesso!");

        String message = newService().generateSuccessMessage("agendamento", "Rafael Souza", "consulta com Dr. Alicio em 2030-01-01T10:00:00Z");

        assertThat(message).isEqualTo("Sua consulta foi agendada com sucesso!");

        ArgumentCaptor<String> userPrompt = ArgumentCaptor.forClass(String.class);
        verify(fallbackClient).call(anyString(), userPrompt.capture());
        assertThat(userPrompt.getValue()).contains("agendamento").contains("Rafael Souza");
    }

    @Test
    void generateErrorMessage_repassaOErroParaOLlm() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn("Desculpe, houve um problema.");

        String message = newService().generateErrorMessage("Horário indisponível para este profissional");

        assertThat(message).isEqualTo("Desculpe, houve um problema.");

        ArgumentCaptor<String> userPrompt = ArgumentCaptor.forClass(String.class);
        verify(fallbackClient).call(anyString(), userPrompt.capture());
        assertThat(userPrompt.getValue()).contains("Horário indisponível para este profissional");
    }

    @Test
    void generateUnknownIntentMessage_repassaAMensagemOriginal() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn("Olá! Em que posso ajudar?");

        String message = newService().generateUnknownIntentMessage("oi, tudo bem?");

        assertThat(message).isEqualTo("Olá! Em que posso ajudar?");
        verify(fallbackClient).call(anyString(), eq("oi, tudo bem?"));
    }
}
