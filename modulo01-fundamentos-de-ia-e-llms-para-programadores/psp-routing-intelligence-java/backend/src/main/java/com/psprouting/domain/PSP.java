package com.psprouting.domain;

/**
 * Payment Service Providers suportados pelo roteamento.
 *
 * Conjunto fixo definido no IDEIA.md — mesmo vocabulario usado no prompt RAG
 * (ver {@code prompts/routing-prompt.txt}) e nos dados de seed.
 */
public enum PSP {
    ADYEN,
    BRASPAG,
    BRADESCO,
    SANTANDER,
    NUPAY,
    PICPAY,
    MERCADOPAGO
}
