package com.iadeva.agendamento;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.ai.chat.messages.AssistantMessage;
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
    void call_caiParaOProximoModeloQuandoOPrimeiroFalha() {
        ChatResponse okResponse = new ChatResponse(List.of(new Generation(new AssistantMessage("ok"))));
        when(chatModel.call(any(Prompt.class)))
                .thenThrow(new RuntimeException("modelo indisponivel"))
                .thenReturn(okResponse);

        ResilientChatClient client = new ResilientChatClient(chatModel, "model-a,model-b");

        String content = client.call("sys", "user");

        assertThat(content).isEqualTo("ok");
    }

    @Test
    void call_lancaExcecaoQuandoTodosOsModelosFalham() {
        when(chatModel.call(any(Prompt.class))).thenThrow(new RuntimeException("modelo indisponivel"));

        ResilientChatClient client = new ResilientChatClient(chatModel, "model-a,model-b");

        assertThatThrownBy(() -> client.call("sys", "user"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("All fallback models failed");
    }
}
