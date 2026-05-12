package com.iadeva.cognitive.agent;

import com.iadeva.cognitive.model.AgentResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

// ReAct (Reasoning + Acting) — equivalente ao react_agent.py da aula 07
// Cada step raciocina antes de agir: o LLM produz Thought → Action → Observation
// antes de decidir o próximo movimento ou encerrar.
@Component
public class ReactAgent {

    private static final Logger log = LoggerFactory.getLogger(ReactAgent.class);

    private final ChatClient chatClient;

    @Value("${agent.react.max-steps:10}")
    private int maxSteps;

    public ReactAgent(ChatClient.Builder chatClientBuilder) {
        this.chatClient = chatClientBuilder.build();
    }

    /**
     * Executa o loop ReAct para a tarefa fornecida.
     * Cada iteração pede ao LLM que produza:
     *   Thought: [raciocínio]
     *   Action: [ação a tomar]
     *   Observation: [resultado observado]
     * O loop para quando o LLM emite "Action: FINAL_ANSWER" ou atinge maxSteps.
     * Proteção anti-loop: se a mesma Action for escolhida 3 vezes seguidas, encerra.
     *
     * @param input tarefa original do usuário
     * @return AgentResponse com resultado final e métricas
     */
    public AgentResponse execute(String input) {
        log.info("[ReAct] Iniciando execução — tarefa: {}", input);

        // Histórico acumulado de Thought/Action/Observation para contexto
        StringBuilder history = new StringBuilder();
        String lastAction = null;
        int sameActionCount = 0;
        int stepCount = 0;
        int estimatedTokens = 0;
        String finalAnswer = null;

        for (int step = 1; step <= maxSteps; step++) {
            stepCount = step;
            log.debug("[ReAct] Step {}/{}", step, maxSteps);

            String prompt = buildReActPrompt(input, history.toString(), step);

            // Chama o LLM para obter o próximo Thought/Action
            String response = chatClient.prompt()
                    .user(prompt)
                    .call()
                    .content();

            // Estimativa simples de tokens: ~4 chars por token
            estimatedTokens += (prompt.length() + response.length()) / 4;

            log.debug("[ReAct] Resposta LLM step {}: {}", step, response);

            // Extrai a Action da resposta do LLM
            String currentAction = extractField(response, "Action");

            // Adiciona o step ao histórico
            history.append("\n--- Step ").append(step).append(" ---\n")
                   .append(response).append("\n");

            // Verifica se o agente chegou à resposta final
            if (currentAction != null && currentAction.startsWith("FINAL_ANSWER")) {
                finalAnswer = extractFinalAnswer(currentAction);
                log.info("[ReAct] Resposta final obtida no step {}", step);
                break;
            }

            // Proteção anti-loop: mesma action 3x consecutivas → encerra
            if (currentAction != null && currentAction.equals(lastAction)) {
                sameActionCount++;
                if (sameActionCount >= 3) {
                    log.warn("[ReAct] Loop detectado (mesma action {} vezes). Encerrando.", sameActionCount);
                    finalAnswer = "Não foi possível determinar uma resposta definitiva após "
                            + step + " iterações. Último raciocínio: " + extractField(response, "Thought");
                    break;
                }
            } else {
                sameActionCount = 0;
                lastAction = currentAction;
            }
        }

        // Se esgotou os steps sem resposta final
        if (finalAnswer == null) {
            finalAnswer = "Limite de " + maxSteps + " steps atingido sem resposta final. "
                    + "Histórico parcial disponível.";
            log.warn("[ReAct] maxSteps ({}) atingido sem FINAL_ANSWER.", maxSteps);
        }

        AgentResponse.AgentMetrics metrics = new AgentResponse.AgentMetrics(stepCount, estimatedTokens, 0);
        return new AgentResponse(finalAnswer, metrics);
    }

    /**
     * Monta o prompt ReAct com contexto histórico e instrução para o próximo step.
     */
    private String buildReActPrompt(String task, String history, int stepNumber) {
        return """
                Você é um agente ReAct. Resolva a tarefa usando o ciclo Thought → Action → Observation.
                
                Regras:
                - Thought: explique seu raciocínio antes de agir
                - Action: indique a ação como "Action: <nome_da_acao>(<parametros>)"
                  ou "Action: FINAL_ANSWER(<sua_resposta_completa>)" quando tiver a resposta
                - Observation: descreva o resultado hipotético da ação
                
                Tarefa: %s
                
                Histórico até agora:
                %s
                
                Step %d — produza apenas o próximo Thought, Action e Observation:
                """.formatted(task, history.isEmpty() ? "(nenhum)" : history, stepNumber);
    }

    /**
     * Extrai o valor de um campo (ex.: "Action:", "Thought:") da resposta do LLM.
     */
    private String extractField(String response, String fieldName) {
        if (response == null) return null;
        String prefix = fieldName + ":";
        int start = response.indexOf(prefix);
        if (start == -1) return null;
        start += prefix.length();
        int end = response.indexOf("\n", start);
        String value = (end == -1 ? response.substring(start) : response.substring(start, end)).trim();
        return value.isEmpty() ? null : value;
    }

    /**
     * Extrai o texto da resposta final de "FINAL_ANSWER(<texto>)".
     */
    private String extractFinalAnswer(String actionValue) {
        if (actionValue.contains("(") && actionValue.contains(")")) {
            int open = actionValue.indexOf('(');
            int close = actionValue.lastIndexOf(')');
            if (open < close) {
                return actionValue.substring(open + 1, close).trim();
            }
        }
        // Fallback: retorna tudo após o prefixo FINAL_ANSWER
        return actionValue.replace("FINAL_ANSWER", "").trim();
    }
}
