package com.iadeva.evals.model;

/**
 * Métricas consolidadas de benchmark para uma arquitetura cognitiva específica.
 * Calculadas a partir da execução dos cenários do dataset.
 */
public record BenchmarkMetrics(
        String architecture,
        double completionRate,
        double avgSteps,
        int avgTokensEstimated,
        double toolSuccessRate,
        double toolCoverage
) {}
