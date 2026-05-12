package com.iadeva.evals.controller;

import com.iadeva.evals.evaluator.BenchmarkRunner;
import com.iadeva.evals.evaluator.MemoryEvaluator;
import com.iadeva.evals.evaluator.ToolSelectionEvaluator;
import com.iadeva.evals.model.BenchmarkReport;
import com.iadeva.evals.model.MemoryImpactReport;
import com.iadeva.evals.model.ToolSelectionReport;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.attribute.BasicFileAttributes;
import java.time.Instant;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;

/**
 * Controller REST que expõe os endpoints de avaliação de agentes.
 * Todos os endpoints são POST para evitar execuções acidentais via GET.
 */
@RestController
@RequestMapping("/evals")
public class EvalsController {

    private final BenchmarkRunner benchmarkRunner;
    private final ToolSelectionEvaluator toolSelectionEvaluator;
    private final MemoryEvaluator memoryEvaluator;
    private final String reportsPath;

    public EvalsController(
            BenchmarkRunner benchmarkRunner,
            ToolSelectionEvaluator toolSelectionEvaluator,
            MemoryEvaluator memoryEvaluator,
            @Value("${evals.reports.path:./reports}") String reportsPath
    ) {
        this.benchmarkRunner = benchmarkRunner;
        this.toolSelectionEvaluator = toolSelectionEvaluator;
        this.memoryEvaluator = memoryEvaluator;
        this.reportsPath = reportsPath;
    }

    /**
     * POST /evals/benchmark
     * Executa benchmark comparativo das 3 arquiteturas cognitivas.
     * Salva relatório em ./reports/ e retorna o resultado completo.
     */
    @PostMapping("/benchmark")
    public ResponseEntity<BenchmarkReport> runBenchmark() {
        BenchmarkReport report = benchmarkRunner.run();
        return ResponseEntity.ok(report);
    }

    /**
     * POST /evals/tool-selection
     * Avalia a qualidade das decisões de seleção de ferramentas usando o modelo LLM.
     * Retorna 422 se o agente não atingir o threshold mínimo de acurácia.
     */
    @PostMapping("/tool-selection")
    public ResponseEntity<ToolSelectionReport> runToolSelectionEval() {
        ToolSelectionReport report = toolSelectionEvaluator.evaluate();
        // Retorna 422 Unprocessable Entity quando o agente não passa no critério mínimo
        if (!report.passed()) {
            return ResponseEntity.unprocessableEntity().body(report);
        }
        return ResponseEntity.ok(report);
    }

    /**
     * POST /evals/memory-impact
     * Compara execuções com e sem memória e retorna relatório de impacto.
     */
    @PostMapping("/memory-impact")
    public ResponseEntity<MemoryImpactReport> runMemoryEval() {
        MemoryImpactReport report = memoryEvaluator.evaluate();
        return ResponseEntity.ok(report);
    }

    /**
     * GET /evals/reports
     * Lista todos os relatórios salvos em ./reports/ com nome e timestamp de criação.
     */
    @GetMapping("/reports")
    public ResponseEntity<List<Map<String, Object>>> listReports() {
        File dir = new File(reportsPath);
        if (!dir.exists() || !dir.isDirectory()) {
            return ResponseEntity.ok(Collections.emptyList());
        }

        File[] files = dir.listFiles((d, name) -> name.endsWith(".json"));
        if (files == null || files.length == 0) {
            return ResponseEntity.ok(Collections.emptyList());
        }

        List<Map<String, Object>> reportList = Arrays.stream(files)
                .map(file -> {
                    Instant createdAt;
                    try {
                        BasicFileAttributes attrs = Files.readAttributes(
                                file.toPath(), BasicFileAttributes.class);
                        createdAt = attrs.creationTime().toInstant();
                    } catch (Exception e) {
                        createdAt = Instant.EPOCH;
                    }
                    return Map.<String, Object>of(
                            "nome", file.getName(),
                            "tamanhoBytes", file.length(),
                            "criadoEm", createdAt.toString()
                    );
                })
                .sorted((a, b) -> b.get("criadoEm").toString().compareTo(a.get("criadoEm").toString()))
                .toList();

        return ResponseEntity.ok(reportList);
    }
}
