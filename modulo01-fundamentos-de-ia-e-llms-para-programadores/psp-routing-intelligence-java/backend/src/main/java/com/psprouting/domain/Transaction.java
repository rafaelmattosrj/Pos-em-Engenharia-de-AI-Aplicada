package com.psprouting.domain;

import java.math.BigDecimal;

/**
 * Transacao recebida em {@code POST /api/routing/recommend} — as
 * caracteristicas usadas para gerar o embedding e buscar casos similares.
 *
 * {@code brand} e opcional: metodos {@link PaymentMethod#PIX} e
 * {@link PaymentMethod#WALLET} normalmente nao tem bandeira de cartao.
 */
public record Transaction(
        BigDecimal amount,
        PaymentMethod method,
        String brand,
        String userRegion,
        int hour,
        String merchantCategory
) {
}
