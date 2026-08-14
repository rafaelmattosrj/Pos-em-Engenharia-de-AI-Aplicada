package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.ResilientChatClient;
import com.iadeva.roteamento.graph.GraphState;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class FallbackNodeTest {

    @Mock
    private ResilientChatClient fallbackClient;

    @Test
    void chamaLlmComSystemPromptEmPortuguesERetornaResposta() {
        when(fallbackClient.call(eq("Você é um assistente útil. Responda em português."), eq("qual a capital do brasil?")))
                .thenReturn("resposta do LLM");

        FallbackNode node = new FallbackNode(fallbackClient);
        GraphState state = node.process(GraphState.initial("qual a capital do brasil?"));

        assertThat(state.output()).isEqualTo("resposta do LLM");
        verify(fallbackClient).call("Você é um assistente útil. Responda em português.", "qual a capital do brasil?");
    }

    @Test
    void propagaErroQuandoTodosOsModelosFalham() {
        when(fallbackClient.call(eq("Você é um assistente útil. Responda em português."), eq("pergunta qualquer")))
                .thenThrow(new RuntimeException("All fallback models failed"));

        FallbackNode node = new FallbackNode(fallbackClient);

        assertThatThrownBy(() -> node.process(GraphState.initial("pergunta qualquer")))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("All fallback models failed");
    }
}
