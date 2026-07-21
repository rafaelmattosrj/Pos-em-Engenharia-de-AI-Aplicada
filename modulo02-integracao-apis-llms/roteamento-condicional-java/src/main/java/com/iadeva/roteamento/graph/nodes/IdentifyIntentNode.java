package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.springframework.stereotype.Component;

// Nó de identificação de intenção — analisa o input e define o comando
// Equivalente ao identifyIntentNode.ts do LangGraph
@Component
public class IdentifyIntentNode {

    public GraphState process(GraphState state) {
        String input = state.messages().getLast().toLowerCase();

        GraphState.Command command;
        if (input.contains("upper")) {
            command = GraphState.Command.UPPERCASE;
        } else if (input.contains("lower")) {
            command = GraphState.Command.LOWERCASE;
        } else {
            command = GraphState.Command.UNKNOWN;
        }

        return state.withCommand(command);
    }
}
