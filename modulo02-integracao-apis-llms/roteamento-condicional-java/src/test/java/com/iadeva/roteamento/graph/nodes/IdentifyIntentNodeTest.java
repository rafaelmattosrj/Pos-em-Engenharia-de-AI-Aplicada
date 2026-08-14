package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class IdentifyIntentNodeTest {

    private final IdentifyIntentNode node = new IdentifyIntentNode();

    @Test
    void detectaComandoUppercase() {
        GraphState state = node.process(GraphState.initial("por favor coloque em UPPER case"));

        assertThat(state.command()).isEqualTo(GraphState.Command.UPPERCASE);
    }

    @Test
    void detectaComandoLowercase() {
        GraphState state = node.process(GraphState.initial("LOWER isso por favor"));

        assertThat(state.command()).isEqualTo(GraphState.Command.LOWERCASE);
    }

    @Test
    void retornaUnknownQuandoNaoReconheceComando() {
        GraphState state = node.process(GraphState.initial("qual a capital do brasil?"));

        assertThat(state.command()).isEqualTo(GraphState.Command.UNKNOWN);
    }
}
