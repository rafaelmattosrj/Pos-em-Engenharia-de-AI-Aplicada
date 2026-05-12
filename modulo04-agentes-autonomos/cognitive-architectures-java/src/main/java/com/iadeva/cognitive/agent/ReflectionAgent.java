package com.iadeva.cognitive.agent;

import com.iadeva.cognitive.model.AgentResponse;
import com.iadeva.cognitive.model.CritiqueResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

// Reflection — equivalente ao reflection_agent.py da aula 08
// Melhora iterativamente via auto-avaliação: gera output, critica, regenera com feedback.
// O ciclo continua até o CritiqueEvaluator aprovar ou atingir maxCycles.
@Component
public class ReflectionAgent {

    private static final Logger log = LoggerFactory.getLogger(ReflectionAgent.class);

    private final ChatClient chatClient;
    private final CritiqueEvaluator critiqueEvaluator;

    @Value("${agent.reflection.max-cycles:3}")
    private int maxCycles;

    @Value("${agent.reflection.threshold:0.7}")
    private double threshold;

    public ReflectionAgent(ChatClient.Builder chatClientBuilder, CritiqueEvaluator critiqueEvaluator) {
        this.chatClient = chatClientBuilder.build();
        this.critiqueEvaluator = critiqueEvaluator;
    }

    /**
     * Executa o ciclo de reflexão:
     * 1. Gera um output inicial para a tarefa
     * 2. Avalia com CritiqueEvaluator
     * 3. Se score < threshold, regenera incorporando o feedback
     * 4. Repete até passar ou atingir maxCycles
     *
     * @param input tarefa a ser resolvida
     * @return AgentResponse com melhor output obtido e número de reflexões realizadas
     */
    public AgentResponse execute(String input) {
        log.info("[Reflection] Iniciando ciclo de reflexão — tarefa: {}", input);

        String currentOutput = null;
        CritiqueResult lastCritique = null;
        int reflectionCount = 0;
        int estimatedTokens = 0;

        for (int cycle = 1; cycle <= maxCycles; cycle++) {
            log.info("[Reflection] Ciclo {}/{}", cycle, maxCycles);

            // Gera (ou regenera) o output
            String prompt = buildGenerationPrompt(input, currentOutput, lastCritique);
            currentOutput = chatClient.prompt()
                    .user(prompt)
                    .call()
                    .content();
            estimatedTokens += (prompt.length() + currentOutput.length()) / 4;

            log.debug("[Reflection] Output ciclo {}: {}", cycle, currentOutput);

            // Avalia o output produzido
            lastCritique = critiqueEvaluator.evaluate(input, currentOutput);
            log.info("[Reflection] Score ciclo {}: {:.2f} — {}",
                    cycle, lastCritique.score(), lastCritique.passed() ? "APROVADO" : "REPROVADO");

            if (lastCritique.passed()) {
                log.info("[Reflection] Output aprovado no ciclo {}. Encerrando.", cycle);
                break;
            }

            // Se não passou e ainda há ciclos, incrementa o contador de reflexões
            if (cycle < maxCycles) {
                reflectionCount++;
                log.info("[Reflection] Score abaixo do threshold ({}). Aplicando feedback e regenerando.", threshold);
            } else {
                log.warn("[Reflection] maxCycles ({}) atingido. Retornando melhor output disponível.", maxCycles);
            }
        }

        // Formata o resultado final com informações de qualidade
        String finalResult = formatResult(currentOutput, lastCritique, reflectionCount);

        AgentResponse.AgentMetrics metrics = new AgentResponse.AgentMetrics(
                reflectionCount + 1, estimatedTokens, reflectionCount);
        return new AgentResponse(finalResult, metrics);
    }

    /**
     * Monta o prompt de geração.
     * Na primeira iteração (sem output anterior): geração inicial.
     * Nas iterações seguintes: inclui o output anterior e o feedback do crítico.
     */
    private String buildGenerationPrompt(String task, String previousOutput, CritiqueResult critique) {
        if (previousOutput == null || critique == null) {
            // Primeira geração — sem contexto de reflexão
            return """
                    Responda à seguinte tarefa de forma completa, precisa e bem estruturada:
                    
                    Tarefa: %s
                    """.formatted(task);
        }

        // Geração com reflexão — incorpora o feedback do CritiqueEvaluator
        return """
                Você recebeu feedback sobre sua resposta anterior. Melhore-a incorporando as sugestões.
                
                Tarefa original: %s
                
                Sua resposta anterior:
                %s
                
                Feedback do avaliador (score: %.2f/1.0):
                - Correção: %.2f
                - Completude: %.2f
                - Qualidade: %.2f
                - Observações: %s
                
                Gere uma resposta MELHORADA que corrija os pontos levantados:
                """.formatted(
                task,
                previousOutput,
                critique.score(),
                critique.correctness(),
                critique.completeness(),
                critique.quality(),
                critique.feedback()
        );
    }

    /**
     * Formata o resultado final com metadados de qualidade para transparência.
     */
    private String formatResult(String output, CritiqueResult critique, int reflections) {
        StringBuilder sb = new StringBuilder();
        sb.append(output);
        sb.append("\n\n---\n");
        sb.append("*Qualidade: ").append(String.format("%.0f%%", critique.score() * 100));
        if (reflections > 0) {
            sb.append(" após ").append(reflections).append(" reflexão(ões)");
        }
        sb.append(" | ").append(critique.passed() ? "Aprovado" : "Melhor disponível");
        sb.append("*");
        return sb.toString();
    }
}
