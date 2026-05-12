package com.iadeva.agent.core;

import org.springframework.stereotype.Component;

import java.util.concurrent.atomic.AtomicInteger;

/**
 * Circuit breaker para evitar loops infinitos com respostas inválidas da LLM.
 * Quebra após 3 respostas inválidas consecutivas, protegendo contra alucinações
 * ou falhas persistentes no modelo.
 */
@Component
public class CircuitBreaker {

    private static final int THRESHOLD = 3;

    private final AtomicInteger invalidCount = new AtomicInteger(0);

    /**
     * Registra uma resposta inválida do LLM.
     * Incrementa o contador de falhas consecutivas.
     */
    public void recordInvalid() {
        invalidCount.incrementAndGet();
    }

    /**
     * Registra uma resposta válida do LLM.
     * Reseta o contador de falhas consecutivas.
     */
    public void recordValid() {
        invalidCount.set(0);
    }

    /**
     * Verifica se o circuit breaker deve abrir (interromper o loop).
     *
     * @return true se o número de falhas consecutivas atingiu o threshold
     */
    public boolean shouldBreak() {
        return invalidCount.get() >= THRESHOLD;
    }

    /**
     * Reseta o circuit breaker para o estado inicial.
     * Deve ser chamado no início de cada nova execução do agente.
     */
    public void reset() {
        invalidCount.set(0);
    }

    /** Retorna o número atual de respostas inválidas consecutivas */
    public int getInvalidCount() {
        return invalidCount.get();
    }
}
