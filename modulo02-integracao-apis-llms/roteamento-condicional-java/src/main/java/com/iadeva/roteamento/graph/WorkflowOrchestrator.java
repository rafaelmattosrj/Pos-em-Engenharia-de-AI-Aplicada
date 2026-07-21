package com.iadeva.roteamento.graph;

import com.iadeva.roteamento.graph.nodes.*;
import org.springframework.stereotype.Component;

// Orquestrador do fluxo — substitui o StateGraph do LangGraph
// Implementa o mesmo padrão de nós + arestas condicionais
@Component
public class WorkflowOrchestrator {

    private final IdentifyIntentNode identifyIntent;
    private final UpperCaseNode upperCase;
    private final LowerCaseNode lowerCase;
    private final FallbackNode fallback;
    private final ChatResponseNode chatResponse;

    public WorkflowOrchestrator(
            IdentifyIntentNode identifyIntent,
            UpperCaseNode upperCase,
            LowerCaseNode lowerCase,
            FallbackNode fallback,
            ChatResponseNode chatResponse) {
        this.identifyIntent = identifyIntent;
        this.upperCase = upperCase;
        this.lowerCase = lowerCase;
        this.fallback = fallback;
        this.chatResponse = chatResponse;
    }

    // Fluxo: START → identifyIntent → (conditional) → uppercase/lowercase/fallback → chatResponse → END
    public GraphState invoke(String input) {
        GraphState state = GraphState.initial(input);

        // Nó 1: identifica a intenção
        state = identifyIntent.process(state);

        // Aresta condicional baseada no comando identificado
        state = switch (state.command()) {
            case UPPERCASE -> upperCase.process(state);
            case LOWERCASE -> lowerCase.process(state);
            case UNKNOWN -> fallback.process(state);
        };

        // Nó final
        return chatResponse.process(state);
    }
}
