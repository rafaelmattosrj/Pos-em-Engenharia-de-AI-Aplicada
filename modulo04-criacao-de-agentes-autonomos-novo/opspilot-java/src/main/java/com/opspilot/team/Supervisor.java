package com.opspilot.team;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.opspilot.llm.ChatMessage;
import com.opspilot.llm.ChatModel;

import java.util.List;

/**
 * Porte de createDecideNext (team/supervisor.ts): pede ao modelo uma decisao
 * estruturada {next, brief} via prompt (equivalente idiomatico ao
 * withStructuredOutput do LangChain, que nao tem par direto fora do
 * ecossistema JS/Python) e faz o parse manual da resposta.
 */
public final class Supervisor {

    public static final String SYSTEM_PROMPT = String.join("\n",
            "Voce e o SUPERVISOR do plantao de incidentes. Decida qual papel age a seguir:",
            "analista (diagnostico factual, so leitura), planejador (transforma fatos em plano,",
            "sem ferramentas) ou executor (executa acoes de incidente). Responda 'done' quando o",
            "pedido do plantonista estiver atendido.",
            "Responda SOMENTE com um JSON no formato {\"next\": \"analista|planejador|executor|done\", \"brief\": \"...\"}.");

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private Supervisor() {
    }

    public static DecideNextFn createDecideNext(ChatModel model) {
        return (message, blackboard, handoffCount) -> {
            String userMessage = String.join("\n",
                    "Pedido do plantonista: " + message,
                    "Delegacoes ja usadas: " + handoffCount,
                    "",
                    "Blackboard:",
                    Blackboard.render(blackboard));

            var response = model.invoke(List.of(
                    ChatMessage.system(SYSTEM_PROMPT),
                    ChatMessage.user(userMessage)));

            return parseDecision(response.content());
        };
    }

    public static SupervisorDecision parseDecision(String rawJson) {
        JsonNode node;
        try {
            node = MAPPER.readTree(rawJson);
        } catch (Exception e) {
            throw new IllegalArgumentException("decisao invalida do supervisor: JSON malformado: " + e.getMessage(), e);
        }

        JsonNode nextNode = node.get("next");
        JsonNode briefNode = node.get("brief");
        if (nextNode == null || briefNode == null) {
            throw new IllegalArgumentException("decisao invalida do supervisor: campos 'next'/'brief' ausentes");
        }

        String next = nextNode.asText();
        String brief = briefNode.asText();
        if ("done".equals(next)) {
            return SupervisorDecision.done(brief);
        }
        try {
            return new SupervisorDecision(TeamRole.valueOf(next.toUpperCase()), brief);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("decisao invalida do supervisor: papel desconhecido \"" + next + "\"", e);
        }
    }
}
