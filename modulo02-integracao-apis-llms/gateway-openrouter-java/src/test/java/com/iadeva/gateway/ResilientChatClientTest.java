package com.iadeva.gateway;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.ai.chat.messages.AssistantMessage;
import org.springframework.ai.chat.metadata.ChatResponseMetadata;
import org.springframework.ai.chat.model.ChatModel;
import org.springframework.ai.chat.model.ChatResponse;
import org.springframework.ai.chat.model.Generation;
import org.springframework.ai.chat.prompt.Prompt;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ResilientChatClientTest {

    @Mock
    private ChatModel chatModel;

    @Test
    void callForResponse_caiParaOProximoModeloQuandoOPrimeiroFalha() {
        ChatResponse okResponse = new ChatResponse(
                List.of(new Generation(new AssistantMessage("resposta do fallback"))),
                ChatResponseMetadata.builder().model("model-b").build());

        when(chatModel.call(any(Prompt.class)))
                .thenThrow(new RuntimeException("modelo indisponivel"))
                .thenReturn(okResponse);

        ResilientChatClient client = new ResilientChatClient(chatModel, "model-a,model-b");

        ChatResponse response = client.callForResponse("sys", "pergunta valida");

        assertThat(response.getResult().getOutput().getText()).isEqualTo("resposta do fallback");
        assertThat(response.getMetadata().getModel()).isEqualTo("model-b");
    }

    @Test
    void callForResponse_lancaExcecaoQuandoTodosOsModelosFalham() {
        when(chatModel.call(any(Prompt.class))).thenThrow(new RuntimeException("modelo indisponivel"));

        ResilientChatClient client = new ResilientChatClient(chatModel, "model-a,model-b");

        assertThatThrownBy(() -> client.callForResponse("sys", "pergunta valida"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("All fallback models failed");
    }
}
