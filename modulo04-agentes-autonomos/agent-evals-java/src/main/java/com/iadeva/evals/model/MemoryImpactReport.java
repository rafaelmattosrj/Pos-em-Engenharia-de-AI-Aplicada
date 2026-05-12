package com.iadeva.evals.model;

/**
 * Relatório de impacto da memória comparando execuções com e sem memória.
 * Métricas calculadas sobre múltiplas execuções simuladas.
 */
public record MemoryImpactReport(
        int totalRuns,
        double retrievalPrecision,
        double retrievalRecall,
        double memoryUtilization,
        double hallucinationFromMemory,
        double decisionImprovement,
        double lessonQuality
) {}
