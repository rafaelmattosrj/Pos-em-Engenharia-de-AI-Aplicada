package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class LowerCaseNodeTest {

    private final LowerCaseNode node = new LowerCaseNode();

    @Test
    void convertOutputParaLowercase() {
        GraphState state = node.process(GraphState.initial("CONVERTE PARA LOWER"));

        assertThat(state.output()).isEqualTo("converte para lower");
    }
}
