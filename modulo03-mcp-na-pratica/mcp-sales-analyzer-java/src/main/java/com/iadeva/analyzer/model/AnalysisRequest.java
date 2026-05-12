package com.iadeva.analyzer.model;

// Representa a requisição de análise recebida pelo endpoint POST /analyze
public record AnalysisRequest(
        // Pergunta do usuário sobre os dados (ex: "Qual produto vendeu mais?")
        String question,
        // Dados brutos em formato CSV ou JSON a serem analisados
        String data
) {}
