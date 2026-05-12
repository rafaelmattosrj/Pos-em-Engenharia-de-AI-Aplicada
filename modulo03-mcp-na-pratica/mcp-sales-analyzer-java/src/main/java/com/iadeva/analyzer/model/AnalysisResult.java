package com.iadeva.analyzer.model;

import java.util.List;

// Resultado final da análise — retornado pelo endpoint POST /analyze
public record AnalysisResult(
        // Relatório textual com a resposta à pergunta do usuário
        String report,
        // Lista de tools efetivamente utilizadas durante o processamento
        List<String> toolsUsed,
        // Passos de processamento executados (para rastreabilidade/debug)
        List<String> processingSteps
) {}
