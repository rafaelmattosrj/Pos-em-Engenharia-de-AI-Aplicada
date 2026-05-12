package com.iadeva.evals.evaluator;

// Equivalente ao tool_selection_eval.py da aula 12 — avalia qualidade das decisões de tool selection

import com.iadeva.evals.model.ToolSelectionCase;
import com.iadeva.evals.model.ToolSelectionReport;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;
import java.util.regex.Pattern;

/**
 * Avalia a qualidade das decisões de seleção de ferramentas do agente.
 * Usa o ChatClient do Spring AI para simular decisões reais do modelo.
 * O agente recebe um contexto e precisa escolher a ferramenta correta.
 */
@Component
public class ToolSelectionEvaluator {

    private final ChatClient chatClient;
    private final double minAccuracy;
    private final double maxUnnecessary;

    // Ferramentas disponíveis no domínio de monitoramento de incidentes
    private static final List<String> AVAILABLE_TOOLS = List.of(
            "getMetrics", "getLogs", "getDeployHistory", "saveIncident", "notifyTeam"
    );

    public ToolSelectionEvaluator(
            ChatClient.Builder chatClientBuilder,
            @Value("${evals.tool-selection.min-accuracy:0.80}") double minAccuracy,
            @Value("${evals.tool-selection.max-unnecessary:0.10}") double maxUnnecessary
    ) {
        this.chatClient = chatClientBuilder.build();
        this.minAccuracy = minAccuracy;
        this.maxUnnecessary = maxUnnecessary;
    }

    /**
     * Executa a avaliação de tool selection sobre os 5 cenários hardcoded.
     * Retorna o relatório com métricas de acurácia e taxa de chamadas desnecessárias.
     */
    public ToolSelectionReport evaluate() {
        List<ToolSelectionCase> cases = buildTestCases();

        int correctTool = 0;
        int correctArgs = 0;
        int unnecessaryCalls = 0;
        int wrongTool = 0;

        for (ToolSelectionCase tc : cases) {
            String response = askModel(tc);
            String selected = extractToolName(response);

            boolean toolCorreto = tc.expectedTool().equalsIgnoreCase(selected);
            boolean ferramenta_proibida = tc.forbiddenTools().stream()
                    .anyMatch(f -> f.equalsIgnoreCase(selected));

            if (toolCorreto) {
                correctTool++;
                // Verifica se os argumentos-chave estão presentes na resposta
                boolean argsOk = tc.expectedArgs().keySet().stream()
                        .allMatch(k -> response.toLowerCase().contains(k.toLowerCase()));
                if (argsOk) correctArgs++;
            } else {
                wrongTool++;
            }

            if (ferramenta_proibida) {
                unnecessaryCalls++;
            }
        }

        int total = cases.size();
        double toolAccuracy = (double) correctTool / total;
        double argAccuracy = (double) correctArgs / total;
        double unnecessaryRate = (double) unnecessaryCalls / total;
        double wrongRate = (double) wrongTool / total;

        // Critério de aprovação: accuracy >= minAccuracy E unnecessaryRate <= maxUnnecessary
        boolean passed = toolAccuracy >= minAccuracy && unnecessaryRate <= maxUnnecessary;

        return new ToolSelectionReport(
                total,
                round(toolAccuracy),
                round(argAccuracy),
                round(unnecessaryRate),
                round(wrongRate),
                passed
        );
    }

    /**
     * Consulta o modelo para decidir qual ferramenta usar dado o contexto do caso de teste.
     */
    private String askModel(ToolSelectionCase tc) {
        String prompt = """
                Você é um agente de monitoramento de sistemas. Dado o contexto abaixo, escolha UMA ferramenta.
                
                Ferramentas disponíveis: %s
                
                Contexto: %s
                Passo do loop: %d
                
                Responda APENAS com o nome da ferramenta e os argumentos em formato JSON.
                Exemplo: {"tool": "getMetrics", "args": {"service": "payment-api"}}
                """.formatted(AVAILABLE_TOOLS, tc.context(), tc.loopStep());

        try {
            return chatClient.prompt()
                    .user(prompt)
                    .call()
                    .content();
        } catch (Exception e) {
            // Em caso de falha na API, retorna resposta vazia para contabilizar como errada
            System.err.println("[ToolSelectionEvaluator] Erro na chamada ao modelo: " + e.getMessage());
            return "{}";
        }
    }

    /**
     * Extrai o nome da ferramenta da resposta do modelo via regex.
     */
    private String extractToolName(String response) {
        Pattern pattern = Pattern.compile("\"tool\"\\s*:\\s*\"([^\"]+)\"");
        var matcher = pattern.matcher(response);
        return matcher.find() ? matcher.group(1) : "unknown";
    }

    /**
     * Define os 5 casos de teste de tool selection.
     * Cada caso representa uma situação real de um loop de agente de monitoramento.
     */
    private List<ToolSelectionCase> buildTestCases() {
        return List.of(
                new ToolSelectionCase(
                        "ts-001",
                        "Latência alta detectada em payment-api. Ainda não coletamos dados de métricas.",
                        1,
                        "getMetrics",
                        Map.of("service", "payment-api", "metric", "latency"),
                        List.of("saveIncident", "notifyTeam")
                ),
                new ToolSelectionCase(
                        "ts-002",
                        "Métricas coletadas mostram CPU em 95%. Preciso entender o que aconteceu antes de abrir incidente.",
                        2,
                        "getLogs",
                        Map.of("service", "payment-api", "level", "ERROR"),
                        List.of("saveIncident", "notifyTeam")
                ),
                new ToolSelectionCase(
                        "ts-003",
                        "Logs mostram OutOfMemoryError às 14h. Houve algum deploy próximo desse horário?",
                        3,
                        "getDeployHistory",
                        Map.of("service", "payment-api", "since", "13:00"),
                        List.of("saveIncident", "notifyTeam")
                ),
                new ToolSelectionCase(
                        "ts-004",
                        "Deploy identificado às 13:45. Causa raiz confirmada: memory leak na versão 2.3.1. Devo registrar o incidente.",
                        4,
                        "saveIncident",
                        Map.of("service", "payment-api", "severity", "high"),
                        List.of("getMetrics", "getLogs")
                ),
                new ToolSelectionCase(
                        "ts-005",
                        "Incidente registrado. Time de on-call precisa ser notificado imediatamente.",
                        5,
                        "notifyTeam",
                        Map.of("channel", "on-call", "severity", "high"),
                        List.of("getMetrics", "getLogs", "getDeployHistory")
                )
        );
    }

    private double round(double value) {
        return Math.round(value * 10000.0) / 10000.0;
    }
}
