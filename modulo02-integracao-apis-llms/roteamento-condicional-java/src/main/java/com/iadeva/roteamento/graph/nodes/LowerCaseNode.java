package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.springframework.stereotype.Component;

@Component
public class LowerCaseNode {
    public GraphState process(GraphState state) {
        return state.withOutput(state.output().toLowerCase());
    }
}
