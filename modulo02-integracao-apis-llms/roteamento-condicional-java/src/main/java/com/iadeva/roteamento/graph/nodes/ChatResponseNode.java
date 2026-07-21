package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.graph.GraphState;
import org.springframework.stereotype.Component;

// Nó final — apenas retorna o estado com o output processado
@Component
public class ChatResponseNode {
    public GraphState process(GraphState state) {
        return state; // output já foi transformado pelos nós anteriores
    }
}
