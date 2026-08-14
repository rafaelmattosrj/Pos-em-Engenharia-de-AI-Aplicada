package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class UpperCaseNodeTest {

    private final UpperCaseNode node = new UpperCaseNode();

    @Test
    void convertOutputParaUppercase() {
        GraphState state = node.process(GraphState.initial("converte para upper"));

        assertThat(state.output()).isEqualTo("CONVERTE PARA UPPER");
    }
}
