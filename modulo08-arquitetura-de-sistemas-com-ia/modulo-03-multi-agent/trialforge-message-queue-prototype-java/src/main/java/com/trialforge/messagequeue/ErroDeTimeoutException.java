package com.trialforge.messagequeue;

/**
 * Porte de ErroDeTimeout (JS) / ErroDeTimeout (Python).
 *
 * <p>Sinaliza que um agente nao respondeu dentro do timeout — descoberto de
 * fora, pela corrida (Future#get com timeout) contra um temporizador real,
 * nunca por uma excecao lancada pelo proprio agente.</p>
 */
public class ErroDeTimeoutException extends RuntimeException {

    public ErroDeTimeoutException(String nomeAgente, long timeoutMs) {
        super(nomeAgente + ": timeout de " + timeoutMs + "ms excedido — agente não respondeu a tempo.");
    }
}
