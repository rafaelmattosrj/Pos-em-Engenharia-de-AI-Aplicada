package com.iadeva.roteamento.graph.nodes;

import com.iadeva.roteamento.ResilientChatClient;
import com.iadeva.roteamento.graph.GraphState;
import org.springframework.stereotype.Component;

// Fallback: quando o comando é desconhecido, chama o LLM para resposta livre
@Component
public class FallbackNode {

    private final ResilientChatClient fallbackClient;

    public FallbackNode(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public GraphState process(GraphState state) {
        String response = fallbackClient.call(
                "Você é um assistente útil. Responda em português.",
                state.output());

        return state.withOutput(response);
    }
}
