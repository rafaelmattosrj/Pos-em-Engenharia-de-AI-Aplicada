package com.psprouting.application;

import com.psprouting.domain.PSP;

import java.util.List;

/**
 * Recomendacao decodificada da resposta JSON do LLM — ainda sem
 * {@code similarCases}, que e anexado pelo {@link RoutingService} apos a
 * busca de similaridade (o LLM nunca inventa esses casos, apenas os recebe
 * no prompt).
 */
public record ParsedRecommendation(
        PSP primary,
        double confidence,
        String reasoning,
        List<PSP> fallback
) {
}
