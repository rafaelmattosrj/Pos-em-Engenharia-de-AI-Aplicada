package com.iadeva.gateway;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.ai.chat.messages.AssistantMessage;
import org.springframework.ai.chat.metadata.ChatResponseMetadata;
import org.springframework.ai.chat.model.ChatResponse;
import org.springframework.ai.chat.model.Generation;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class OpenRouterServiceTest {

    @Mock
    private ResilientChatClient fallbackClient;

    private OpenRouterService newService(String systemPrompt) throws Exception {
        OpenRouterService service = new OpenRouterService(fallbackClient);
        var field = OpenRouterService.class.getDeclaredField("systemPrompt");
        field.setAccessible(true);
        field.set(service, systemPrompt);
        return service;
    }

    @Test
    void generate_retornaModeloEConteudoDaResposta() throws Exception {
        ChatResponse response = new ChatResponse(
                List.of(new Generation(new AssistantMessage("resposta"))),
                ChatResponseMetadata.builder().model("model-a").build());
        when(fallbackClient.callForResponse(anyString(), anyString())).thenReturn(response);

        LlmResponse llmResponse = newService("You are a helpful assistant.").generate("pergunta valida");

        assertThat(llmResponse.content()).isEqualTo("resposta");
        assertThat(llmResponse.model()).isEqualTo("model-a");

        ArgumentCaptor<String> systemPromptCaptor = ArgumentCaptor.forClass(String.class);
        verify(fallbackClient).callForResponse(systemPromptCaptor.capture(), eq("pergunta valida"));
        assertThat(systemPromptCaptor.getValue()).isEqualTo("You are a helpful assistant.");
    }

    @Test
    void generate_propagaExcecaoQuandoTodosOsModelosFalham() throws Exception {
        when(fallbackClient.callForResponse(anyString(), anyString()))
                .thenThrow(new RuntimeException("All fallback models failed"));

        OpenRouterService service = newService("sys");

        assertThatThrownBy(() -> service.generate("pergunta valida"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("All fallback models failed");
    }
}
