package com.iadeva.evals.evaluator;

// Equivalente ao benchmark_runner.py da aula 09 — avalia comparativamente as 3 arquiteturas cognitivas

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.iadeva.evals.dataset.EvalDataset;
import com.iadeva.evals.model.BenchmarkMetrics;
import com.iadeva.evals.model.BenchmarkReport;
import com.iadeva.evals.model.EvalScenario;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;

/**
 * Executa benchmarks comparativos entre as 3 arquiteturas cognitivas:
 * ReactAgent, PlanExecuteAgent e ReflectionAgent.
 * Os resultados são simulados com variações realistas para fins de avaliação do framework.
 */
@Component
public class BenchmarkRunner {

    private final EvalDataset evalDataset;
    private final ObjectMapper objectMapper;
    private final String reportsPath;

    // Semente fixa para reprodutibilidade dos resultados simulados
    private final Random random = new Random(42);

    public BenchmarkRunner(
            EvalDataset evalDataset,
            ObjectMapper objectMapper,
            @Value("${evals.reports.path:./reports}") String reportsPath
    ) {
        this.evalDataset = evalDataset;
        // Configura ObjectMapper com indentação e suporte a Instant
        this.objectMapper = objectMapper.copy()
                .enable(SerializationFeature.INDENT_OUTPUT)
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
                .findAndRegisterModules();
        this.reportsPath = reportsPath;
    }

    /**
     * Executa o benchmark completo e retorna o relatório comparativo.
     * O relatório também é salvo em disco em ./reports/.
     */
    public BenchmarkReport run() {
        List<EvalScenario> scenarios = evalDataset.getScenarios();

        // Simula execução dos 3 agentes em cada cenário
        BenchmarkMetrics react = simulateArchitecture("ReactAgent", scenarios, 0.82, 4.2, 1800);
        BenchmarkMetrics planExecute = simulateArchitecture("PlanExecuteAgent", scenarios, 0.88, 3.1, 2400);
        BenchmarkMetrics reflection = simulateArchitecture("ReflectionAgent", scenarios, 0.76, 5.8, 3200);

        List<BenchmarkMetrics> results = List.of(react, planExecute, reflection);

        // Gera veredicto comparativo por critério
        Map<String, String> verdict = gerarVeredicto(results);

        BenchmarkReport report = new BenchmarkReport(
                Instant.now(),
                scenarios.size(),
                results,
                verdict
        );

        salvarRelatorio(report);
        return report;
    }

    /**
     * Simula métricas realistas para uma arquitetura com pequenas variações aleatórias.
     *
     * @param architecture nome da arquitetura
     * @param scenarios    cenários do dataset
     * @param baseCompletion taxa base de conclusão
     * @param baseSteps      média base de steps
     * @param baseTokens     estimativa base de tokens
     */
    private BenchmarkMetrics simulateArchitecture(
            String architecture,
            List<EvalScenario> scenarios,
            double baseCompletion,
            double baseSteps,
            int baseTokens
    ) {
        // Aplica variação de ±5% sobre os valores base para simular ruído real
        double completionRate = clamp(baseCompletion + (random.nextDouble() - 0.5) * 0.10, 0, 1);
        double avgSteps = Math.max(1.0, baseSteps + (random.nextDouble() - 0.5) * 1.0);
        int avgTokens = (int) (baseTokens + (random.nextDouble() - 0.5) * 400);
        double toolSuccessRate = clamp(completionRate + 0.05 + random.nextDouble() * 0.05, 0, 1);

        // Calcula cobertura de ferramentas: percentual de expectedTools que seriam chamadas
        double totalExpected = scenarios.stream()
                .mapToInt(s -> s.expectedTools().size())
                .sum();
        double covered = totalExpected * clamp(completionRate + random.nextDouble() * 0.08, 0, 1);
        double toolCoverage = covered / totalExpected;

        return new BenchmarkMetrics(
                architecture,
                round(completionRate),
                round(avgSteps),
                avgTokens,
                round(toolSuccessRate),
                round(toolCoverage)
        );
    }

    /**
     * Gera veredicto identificando a arquitetura vencedora em cada critério.
     */
    private Map<String, String> gerarVeredicto(List<BenchmarkMetrics> results) {
        Map<String, String> verdict = new HashMap<>();

        // Melhor taxa de conclusão
        results.stream()
                .max((a, b) -> Double.compare(a.completionRate(), b.completionRate()))
                .ifPresent(m -> verdict.put("melhor_conclusao", m.architecture()));

        // Menor número de steps (mais eficiente)
        results.stream()
                .min((a, b) -> Double.compare(a.avgSteps(), b.avgSteps()))
                .ifPresent(m -> verdict.put("mais_eficiente", m.architecture()));

        // Menor consumo de tokens
        results.stream()
                .min((a, b) -> Integer.compare(a.avgTokensEstimated(), b.avgTokensEstimated()))
                .ifPresent(m -> verdict.put("menor_custo", m.architecture()));

        // Melhor cobertura de ferramentas
        results.stream()
                .max((a, b) -> Double.compare(a.toolCoverage(), b.toolCoverage()))
                .ifPresent(m -> verdict.put("melhor_cobertura_ferramentas", m.architecture()));

        // Recomendação geral: arquitetura com maior score ponderado
        String recomendado = results.stream()
                .max((a, b) -> Double.compare(scoreGeral(a), scoreGeral(b)))
                .map(BenchmarkMetrics::architecture)
                .orElse("indeterminado");
        verdict.put("recomendado_geral", recomendado);

        return verdict;
    }

    /**
     * Score ponderado para comparação geral: 40% conclusão + 30% cobertura + 30% toolSuccess.
     */
    private double scoreGeral(BenchmarkMetrics m) {
        return 0.40 * m.completionRate()
                + 0.30 * m.toolCoverage()
                + 0.30 * m.toolSuccessRate();
    }

    /**
     * Persiste o relatório em JSON no diretório configurado.
     */
    private void salvarRelatorio(BenchmarkReport report) {
        try {
            File dir = new File(reportsPath);
            if (!dir.exists()) {
                dir.mkdirs();
            }
            String timestamp = DateTimeFormatter.ofPattern("yyyyMMdd_HHmmss")
                    .withZone(java.time.ZoneId.systemDefault())
                    .format(report.generatedAt());
            File file = new File(dir, "benchmark_" + timestamp + ".json");
            objectMapper.writeValue(file, report);
        } catch (IOException e) {
            // Logar mas não falhar — o relatório em memória ainda é retornado
            System.err.println("[BenchmarkRunner] Falha ao salvar relatório: " + e.getMessage());
        }
    }

    private double clamp(double value, double min, double max) {
        return Math.max(min, Math.min(max, value));
    }

    private double round(double value) {
        return Math.round(value * 10000.0) / 10000.0;
    }
}
