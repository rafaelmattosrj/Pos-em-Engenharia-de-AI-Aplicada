package com.iadeva.evals.model;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Relatório completo de benchmark comparando as 3 arquiteturas cognitivas.
 * Inclui veredicto com a arquitetura recomendada para cada critério.
 */
public record BenchmarkReport(
        Instant generatedAt,
        int totalScenarios,
        List<BenchmarkMetrics> results,
        Map<String, String> verdict
) {}
