package com.psprouting.application;

import com.psprouting.domain.HistoricalTransaction;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.Transaction;

import java.math.BigDecimal;

/**
 * Serializa uma transacao em texto natural em pt-BR para gerar o embedding —
 * passo (1) do pipeline descrito em IDEIA.md:
 *
 * <pre>
 * "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA,
 *  regiao SP, hora 20h, categoria STREAMING."
 * </pre>
 *
 * Extraida como classe pura e testavel (sem dependencia do modelo de
 * embeddings) porque a mesma regra e usada tanto para a transacao recebida
 * em {@code /recommend} quanto para cada transacao historica do seed.
 */
public final class TransactionTextSerializer {

    private TransactionTextSerializer() {
    }

    public static String serialize(Transaction transaction) {
        return serialize(
                transaction.amount(),
                transaction.method(),
                transaction.brand(),
                transaction.userRegion(),
                transaction.hour(),
                transaction.merchantCategory()
        );
    }

    public static String serialize(HistoricalTransaction transaction) {
        return serialize(
                transaction.amount(),
                transaction.method(),
                transaction.brand(),
                transaction.region(),
                transaction.hour(),
                transaction.category()
        );
    }

    private static String serialize(
            BigDecimal amount,
            PaymentMethod method,
            String brand,
            String region,
            int hour,
            String category
    ) {
        StringBuilder text = new StringBuilder("Pagamento via ")
                .append(method)
                .append(", valor R$")
                .append(amount.setScale(2, java.math.RoundingMode.HALF_UP));

        if (brand != null && !brand.isBlank()) {
            text.append(", bandeira ").append(brand);
        }

        text.append(", regiao ").append(region)
                .append(", hora ").append(hour).append("h")
                .append(", categoria ").append(category)
                .append(".");

        return text.toString();
    }
}
