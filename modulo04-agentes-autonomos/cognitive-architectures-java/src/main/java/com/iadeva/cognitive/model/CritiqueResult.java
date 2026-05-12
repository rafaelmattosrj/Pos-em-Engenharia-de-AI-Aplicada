package com.iadeva.cognitive.model;

/**
 * Resultado da avaliação crítica produzida pelo CritiqueEvaluator.
 * Usa LLM como juiz para pontuar o output em três dimensões.
 *
 * @param score        pontuação global (média ponderada das três dimensões), escala 0.0 – 1.0
 * @param passed       true se score >= threshold configurado (padrão 0.7)
 * @param feedback     texto explicativo com sugestões de melhoria
 * @param correctness  dimensão de correção factual, escala 0.0 – 1.0
 * @param completeness dimensão de completude da resposta, escala 0.0 – 1.0
 * @param quality      dimensão de qualidade de escrita e clareza, escala 0.0 – 1.0
 */
public record CritiqueResult(
        double score,
        boolean passed,
        String feedback,
        double correctness,
        double completeness,
        double quality
) {}
