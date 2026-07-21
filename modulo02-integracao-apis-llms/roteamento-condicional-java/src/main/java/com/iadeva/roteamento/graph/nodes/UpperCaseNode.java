package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.springframework.stereotype.Component;

@Component
public class UpperCaseNode {
    public GraphState process(GraphState state) {
        return state.withOutput(state.output().toUpperCase());
    }
}
