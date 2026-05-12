package com.iadeva.analyzer.model;

import java.util.List;

// Resultado intermediário da extração de intenção — produzido pelo IntentNode
public record IntentResult(
        // Tipo de dado identificado: "csv" ou "json"
        String dataType,
        // Dados já convertidos/normalizados para processamento (geralmente JSON)
        String parsedData,
        // Pergunta original do usuário
        String question,
        // Lista de tools sugeridas pelo LLM para responder à pergunta
        List<String> suggestedTools
) {}
