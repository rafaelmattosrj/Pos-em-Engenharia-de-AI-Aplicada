package com.psprouting.domain;

import java.math.BigDecimal;

/**
 * Caso historico recuperado do Neo4j por similaridade de cosseno com a
 * transacao atual — item de {@code similarCases} na resposta de
 * {@code POST /api/routing/recommend} (ver exemplo em IDEIA.md).
 */
public record SimilarCase(
        PSP psp,
        TransactionStatus status,
        BigDecimal amount,
        PaymentMethod method,
        double similarity
) {
}
