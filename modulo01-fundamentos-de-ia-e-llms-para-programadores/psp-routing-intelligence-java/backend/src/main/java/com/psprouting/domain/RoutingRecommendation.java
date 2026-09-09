package com.psprouting.domain;

import java.util.List;

/**
 * Resultado completo do pipeline de roteamento — corpo de resposta de
 * {@code POST /api/routing/recommend}, exatamente no shape descrito em
 * IDEIA.md: {@code primary}, {@code confidence}, {@code reasoning},
 * {@code fallback} (do LLM) mais {@code similarCases} (anexado pelo
 * {@code RoutingService} apos a busca de similaridade).
 */
public record RoutingRecommendation(
        PSP primary,
        double confidence,
        String reasoning,
        List<PSP> fallback,
        List<SimilarCase> similarCases
) {
}
