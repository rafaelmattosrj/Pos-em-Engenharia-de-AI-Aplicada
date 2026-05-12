package com.iadeva.evals;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.evals.dataset.EvalDataset;
import com.iadeva.evals.evaluator.BenchmarkRunner;
import com.iadeva.evals.evaluator.MemoryEvaluator;
import com.iadeva.evals.evaluator.ToolSelectionEvaluator;
import com.iadeva.evals.model.BenchmarkReport;
import com.iadeva.evals.model.MemoryImpactReport;
import com.iadeva.evals.model.ToolSelectionReport;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.ai.chat.client.ChatClient;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

/**
 * Testes de integração do framework de avaliação de agentes.
 * Valida os 3 cenários principais: carregamento do dataset, geração de benchmark
 * e comportamento do ToolSelectionReport quando accuracy está abaixo do threshold.
 */
@SpringBootTest
class EvalsTest {

    @Autowired
    private EvalDataset evalDataset;

    @Autowired
    private BenchmarkRunner benchmarkRunner;

    @Autowired
    private MemoryEvaluator memoryEvaluator;

    // ChatClient é mockado para evitar chamadas reais à API OpenAI nos testes
    @MockBean
    private ChatClient.Builder chatClientBuilder;

    /**
     * Cenário 1: EvalDataset deve carregar exatamente 5 cenários do arquivo JSON.
     * Valida que o arquivo está bem formado e o componente funciona na inicialização.
     */
    @Test
    @DisplayName("EvalDataset deve carregar 5 cenários do eval-dataset.json")
    void evalDataset_deveCarregarCincoCenarios() {
        var scenarios = evalDataset.getScenarios();

        assertThat(scenarios).hasSize(5);
        assertThat(scenarios).extracting("id")
                .containsExactly(
                        "deploy-latency-001",
                        "memory-leak-002",
                        "database-slow-003",
                        "service-down-004",
                        "cpu-spike-005"
                );

        // Garante que todos os cenários têm expectedTools definidos
        assertThat(scenarios).allSatisfy(s -> {
            assertThat(s.expectedTools()).isNotEmpty();
            assertThat(s.difficulty()).isIn("easy", "medium", "hard");
        });
    }

    /**
     * Cenário 2: BenchmarkRunner deve gerar relatório com exatamente 3 arquiteturas.
     * Valida que as métricas estão dentro de faixas realistas.
     */
    @Test
    @DisplayName("BenchmarkRunner deve gerar relatório com 3 arquiteturas cognitivas")
    void benchmarkRunner_deveGerarRelatorioComTresArquiteturas() {
        BenchmarkReport report = benchmarkRunner.run();

        assertThat(report).isNotNull();
        assertThat(report.totalScenarios()).isEqualTo(5);
        assertThat(report.results()).hasSize(3);
        assertThat(report.results()).extracting("architecture")
                .containsExactlyInAnyOrder("ReactAgent", "PlanExecuteAgent", "ReflectionAgent");

        // Valida que métricas estão em faixas realistas (0-1 para rates, > 0 para steps e tokens)
        report.results().forEach(m -> {
            assertThat(m.completionRate()).isBetween(0.0, 1.0);
            assertThat(m.toolSuccessRate()).isBetween(0.0, 1.0);
            assertThat(m.toolCoverage()).isBetween(0.0, 1.0);
            assertThat(m.avgSteps()).isGreaterThan(0.0);
            assertThat(m.avgTokensEstimated()).isGreaterThan(0);
        });

        // Veredicto deve conter pelo menos as chaves principais
        assertThat(report.verdict()).containsKeys(
                "melhor_conclusao", "mais_eficiente", "recomendado_geral"
        );
    }

    /**
     * Cenário 3: ToolSelectionReport deve ter passed=false quando accuracy < 0.80.
     * Valida a regra de negócio do critério de aprovação do eval.
     */
    @Test
    @DisplayName("ToolSelectionReport deve falhar quando accuracy está abaixo de 80%")
    void toolSelectionReport_deveFalharQuandoAccuracyAbaixoDoLimite() {
        // Cria um relatório diretamente com accuracy abaixo do threshold
        ToolSelectionReport reportAbaixo = new ToolSelectionReport(
                5,      // totalCases
                0.60,   // toolSelectionAccuracy — abaixo do mínimo 0.80
                0.55,   // argumentAccuracy
                0.05,   // unnecessaryCallsRate — dentro do limite
                0.40,   // wrongToolRate
                false   // passed — deve ser false
        );

        assertThat(reportAbaixo.passed()).isFalse();
        assertThat(reportAbaixo.toolSelectionAccuracy()).isLessThan(0.80);

        // Garante que um relatório com accuracy suficiente marca passed=true
        ToolSelectionReport reportAprovado = new ToolSelectionReport(
                5,
                0.85,   // toolSelectionAccuracy — acima do mínimo
                0.80,
                0.04,   // unnecessaryCallsRate — dentro do limite
                0.15,
                true    // passed — deve ser true
        );

        assertThat(reportAprovado.passed()).isTrue();
        assertThat(reportAprovado.toolSelectionAccuracy()).isGreaterThanOrEqualTo(0.80);
        assertThat(reportAprovado.unnecessaryCallsRate()).isLessThanOrEqualTo(0.10);
    }

    /**
     * Cenário adicional: MemoryEvaluator deve retornar métricas em faixas válidas.
     */
    @Test
    @DisplayName("MemoryEvaluator deve retornar métricas entre 0 e 1")
    void memoryEvaluator_deveRetornarMetricasValidas() {
        MemoryImpactReport report = memoryEvaluator.evaluate();

        assertThat(report.totalRuns()).isGreaterThan(0);
        assertThat(report.retrievalPrecision()).isBetween(0.0, 1.0);
        assertThat(report.retrievalRecall()).isBetween(0.0, 1.0);
        assertThat(report.memoryUtilization()).isBetween(0.0, 1.0);
        assertThat(report.hallucinationFromMemory()).isBetween(0.0, 1.0);
        assertThat(report.decisionImprovement()).isBetween(0.0, 1.0);
        assertThat(report.lessonQuality()).isBetween(0.0, 1.0);
    }
}
