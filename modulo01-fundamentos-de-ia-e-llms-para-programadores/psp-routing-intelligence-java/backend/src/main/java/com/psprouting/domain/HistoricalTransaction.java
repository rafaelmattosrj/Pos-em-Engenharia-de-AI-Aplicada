package com.psprouting.domain;

import java.math.BigDecimal;

/**
 * Transacao historica de {@code data/transactions-seed.json}, usada para
 * popular o Neo4j com casos reais de aprovacao/reprovacao por PSP.
 *
 * Mesmo shape do {@link Transaction}, acrescido do {@code psp} que de fato
 * processou a transacao e do {@code status} do desfecho — e por isso um
 * registro (e nao uma reutilizacao de {@code Transaction}) para deixar
 * explicito que carrega informacao que só existe no histórico, nunca na
 * requisicao de entrada.
 */
public record HistoricalTransaction(
        BigDecimal amount,
        PaymentMethod method,
        String brand,
        String region,
        int hour,
        String category,
        PSP psp,
        TransactionStatus status
) {
}
