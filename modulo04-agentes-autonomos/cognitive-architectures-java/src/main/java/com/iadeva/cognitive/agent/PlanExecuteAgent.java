package com.iadeva.cognitive.agent;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.cognitive.model.AgentResponse;
import com.iadeva.cognitive.model.ExecutionPlan;
import com.iadeva.cognitive.model.ExecutionPlan.PlanStep;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

// Plan-and-Execute — equivalente ao plan_execute_agent.py da aula 08
// Estratégia: planeja tudo primeiro com uma única chamada LLM,
// depois executa deterministicamente sem consultar o LLM novamente.
@Component
public class PlanExecuteAgent {

    private static final Logger log = LoggerFactory.getLogger(PlanExecuteAgent.class);

    private final ChatClient chatClient;
    private final ObjectMapper objectMapper;

    public PlanExecuteAgent(ChatClient.Builder chatClientBuilder, ObjectMapper objectMapper) {
        this.chatClient = chatClientBuilder.build();
        this.objectMapper = objectMapper;
    }

    /**
     * Executa a arquitetura Plan-and-Execute:
     * 1. Chama LLM uma única vez para gerar o plano completo (JSON estruturado)
     * 2. Executa cada step deterministicamente, simulando a ferramenta
     * 3. Consolida os resultados em uma resposta final
     *
     * @param input tarefa a ser resolvida
     * @return AgentResponse com resultado e métricas de execução
     */
    public AgentResponse execute(String input) {
        log.info("[Plan-Execute] Iniciando — tarefa: {}", input);

        // Fase 1: Planejamento — única chamada ao LLM
        String planPrompt = buildPlanningPrompt(input);
        String planJson = chatClient.prompt()
                .user(planPrompt)
                .call()
                .content();

        int estimatedTokens = (planPrompt.length() + planJson.length()) / 4;
        log.debug("[Plan-Execute] Plano recebido: {}", planJson);

        ExecutionPlan plan = parsePlan(planJson);
        log.info("[Plan-Execute] Plano gerado com {} steps", plan.steps().size());

        // Fase 2: Execução determinística — sem mais chamadas LLM
        StringBuilder executionLog = new StringBuilder();
        executionLog.append("# Resultado da Execução do Plano\n\n");
        executionLog.append("**Tarefa:** ").append(input).append("\n\n");

        for (PlanStep step : plan.steps()) {
            log.debug("[Plan-Execute] Executando step {}: {}", step.stepNumber(), step.description());
            String stepResult = simulateToolExecution(step);
            executionLog.append("## Step ").append(step.stepNumber())
                        .append(": ").append(step.description()).append("\n");
            executionLog.append("- **Ferramenta:** ").append(step.tool()).append("\n");
            executionLog.append("- **Resultado:** ").append(stepResult).append("\n");
            executionLog.append("- **Critério de aceite:** ").append(step.successCriteria()).append("\n\n");
        }

        executionLog.append("## Conclusão\n");
        executionLog.append("Todos os ").append(plan.steps().size())
                    .append(" steps do plano foram executados com sucesso.");

        AgentResponse.AgentMetrics metrics = new AgentResponse.AgentMetrics(
                plan.steps().size(), estimatedTokens, 0);
        return new AgentResponse(executionLog.toString(), metrics);
    }

    /**
     * Monta o prompt de planejamento que instrui o LLM a retornar um JSON válido
     * com a estrutura de ExecutionPlan.
     */
    private String buildPlanningPrompt(String task) {
        return """
                Você é um agente planejador. Analise a tarefa abaixo e crie um plano de execução detalhado.
                
                Retorne APENAS um JSON válido (sem markdown, sem explicações) com a seguinte estrutura:
                {
                  "steps": [
                    {
                      "stepNumber": 1,
                      "description": "descrição do que será feito",
                      "tool": "nome_da_ferramenta",
                      "args": {"param1": "valor1"},
                      "successCriteria": "critério para considerar o step concluído"
                    }
                  ]
                }
                
                Ferramentas disponíveis: search, calculate, summarize, validate, format, store
                
                Tarefa: %s
                """.formatted(task);
    }

    /**
     * Faz parse do JSON retornado pelo LLM para ExecutionPlan.
     * Em caso de falha, retorna um plano de fallback com step genérico.
     */
    private ExecutionPlan parsePlan(String planJson) {
        try {
            // Extrai apenas o bloco JSON caso o LLM tenha incluído texto extra
            String json = extractJson(planJson);
            Map<String, Object> raw = objectMapper.readValue(json, new TypeReference<>() {});
            List<?> rawSteps = (List<?>) raw.get("steps");

            List<PlanStep> steps = rawSteps.stream()
                    .map(s -> {
                        @SuppressWarnings("unchecked")
                        Map<String, Object> stepMap = (Map<String, Object>) s;
                        return new PlanStep(
                                ((Number) stepMap.getOrDefault("stepNumber", 1)).intValue(),
                                (String) stepMap.getOrDefault("description", "step sem descrição"),
                                (String) stepMap.getOrDefault("tool", "generic"),
                                castArgs(stepMap.get("args")),
                                (String) stepMap.getOrDefault("successCriteria", "step concluído")
                        );
                    })
                    .toList();
            return new ExecutionPlan(steps);

        } catch (Exception e) {
            log.warn("[Plan-Execute] Falha ao fazer parse do plano JSON. Usando fallback. Erro: {}", e.getMessage());
            // Plano de fallback: um único step genérico
            PlanStep fallback = new PlanStep(1, "Executar tarefa diretamente",
                    "generic", Map.of("task", "executar"), "tarefa concluída");
            return new ExecutionPlan(List.of(fallback));
        }
    }

    /**
     * Extrai o bloco JSON de uma string que pode conter texto extra do LLM.
     */
    private String extractJson(String text) {
        int start = text.indexOf('{');
        int end = text.lastIndexOf('}');
        if (start != -1 && end != -1 && end > start) {
            return text.substring(start, end + 1);
        }
        return text;
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> castArgs(Object args) {
        if (args instanceof Map<?, ?>) {
            return (Map<String, Object>) args;
        }
        return Map.of();
    }

    /**
     * Simula a execução de um step de ferramenta sem chamar APIs reais.
     * Em produção, aqui seria feito o dispatch para ferramentas reais via Function Calling.
     */
    private String simulateToolExecution(PlanStep step) {
        return switch (step.tool()) {
            case "search"    -> "Resultados de busca obtidos para: " + step.args();
            case "calculate" -> "Cálculo executado com resultado numérico baseado em: " + step.args();
            case "summarize" -> "Resumo gerado com sucesso para o conteúdo fornecido.";
            case "validate"  -> "Validação concluída — dados consistentes.";
            case "format"    -> "Conteúdo formatado conforme template especificado.";
            case "store"     -> "Dados persistidos com sucesso.";
            default          -> "Step '" + step.tool() + "' executado (simulação).";
        };
    }
}
