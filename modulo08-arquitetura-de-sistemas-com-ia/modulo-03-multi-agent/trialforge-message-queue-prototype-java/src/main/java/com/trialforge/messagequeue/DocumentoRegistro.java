package com.trialforge.messagequeue;

import java.util.concurrent.atomic.AtomicInteger;

/**
 * Registro idempotente: a mesma chave nunca gera um segundo documento — so
 * incrementa o contador de tentativas, que e o que permite um retry seguro
 * (paragrafos 72-73 do TP falam em "tentar de novo", mas retry so e seguro
 * se for idempotente).
 */
public final class DocumentoRegistro {

    private final ResultadoICFBase resultado;
    private final AtomicInteger tentativas = new AtomicInteger(1);

    public DocumentoRegistro(ResultadoICFBase resultado) {
        this.resultado = resultado;
    }

    public ResultadoICFBase resultado() {
        return resultado;
    }

    public int getTentativas() {
        return tentativas.get();
    }

    void incrementarTentativas() {
        tentativas.incrementAndGet();
    }
}
