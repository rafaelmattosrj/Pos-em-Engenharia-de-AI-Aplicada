package com.trialforge.messagequeue;

/**
 * Estrategia de reacao paralela ao evento "protocolo:pronto".
 *
 * <p>{@link #PROMISE_ALL} replica o bug documentado nos paragrafos 68-71 do
 * TP: equivalente a {@code Promise.all} (JS) / {@code asyncio.gather} sem
 * {@code return_exceptions=True} (Python) — a primeira falha derruba o lote
 * inteiro, mesmo que o outro agente tenha terminado bem.</p>
 *
 * <p>{@link #PROMISE_ALL_SETTLED} e a correcao: equivalente a
 * {@code Promise.allSettled} (JS) / {@code gather(return_exceptions=True)}
 * (Python) — cada resultado e preservado independente do outro ter falhado.
 * E a estrategia padrao, igual ao original.</p>
 */
public enum Estrategia {
    PROMISE_ALL,
    PROMISE_ALL_SETTLED
}
