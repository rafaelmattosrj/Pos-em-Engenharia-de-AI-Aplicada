package com.trialforge.gateway.ollama;

/**
 * Replica ErroDeTimeoutOllama (JS) / ErroDeTimeoutOllama (Python): o Ollama
 * nao respondeu dentro do timeout configurado.
 */
public class OllamaTimeoutException extends Exception {

    public OllamaTimeoutException(String operacao, long timeoutMs) {
        super(String.format("%s: timeout de %dms excedido — Ollama não respondeu a tempo.", operacao, timeoutMs));
    }
}
