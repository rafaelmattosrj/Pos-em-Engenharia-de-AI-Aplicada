package com.iadeva.evals.model;

/**
 * Relatório da avaliação de seleção de ferramentas.
 * passed=true somente se toolSelectionAccuracy >= 0.80 e unnecessaryCallsRate <= 0.10.
 */
public record ToolSelectionReport(
        int totalCases,
        double toolSelectionAccuracy,
        double argumentAccuracy,
        double unnecessaryCallsRate,
        double wrongToolRate,
        boolean passed
) {}
