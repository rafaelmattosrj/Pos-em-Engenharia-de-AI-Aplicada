package com.iadeva.roteamento.graph;

import com.iadeva.roteamento.ResilientChatClient;
import com.iadeva.roteamento.graph.nodes.*;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class WorkflowOrchestratorTest {

    @Mock
    private ResilientChatClient fallbackClient;

    private WorkflowOrchestrator newOrchestrator() {
        return new WorkflowOrchestrator(
                new IdentifyIntentNode(),
                new UpperCaseNode(),
                new LowerCaseNode(),
                new FallbackNode(fallbackClient),
                new ChatResponseNode());
    }

    @Test
    void roteiaParaUppercaseQuandoComandoReconhecido() {
        GraphState state = newOrchestrator().invoke("por favor coloque em UPPER case");

        assertThat(state.output()).isEqualTo("POR FAVOR COLOQUE EM UPPER CASE");
    }

    @Test
    void roteiaParaLowercaseQuandoComandoReconhecido() {
        GraphState state = newOrchestrator().invoke("LOWER isso por favor");

        assertThat(state.output()).isEqualTo("lower isso por favor");
    }

    @Test
    void roteiaParaFallbackLlmQuandoComandoDesconhecido() {
        when(fallbackClient.call(anyString(), anyString())).thenReturn("resposta do LLM");

        GraphState state = newOrchestrator().invoke("qual a capital do brasil?");

        assertThat(state.output()).isEqualTo("resposta do LLM");
    }

    @Test
    void propagaErroQuandoTodosOsModelosDeFallbackFalham() {
        when(fallbackClient.call(anyString(), anyString()))
                .thenThrow(new RuntimeException("All fallback models failed"));

        WorkflowOrchestrator orchestrator = newOrchestrator();

        assertThatThrownBy(() -> orchestrator.invoke("pergunta qualquer"))
                .isInstanceOf(RuntimeException.class);
    }
}
