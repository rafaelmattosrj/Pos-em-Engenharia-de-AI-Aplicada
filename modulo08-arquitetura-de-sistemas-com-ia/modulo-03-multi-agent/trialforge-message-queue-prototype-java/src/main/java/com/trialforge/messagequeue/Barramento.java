package com.trialforge.messagequeue;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.function.Consumer;

/**
 * Porte minimo do EventEmitter nativo do Node.js (JS) / da classe Barramento
 * sobre asyncio (Python) — a fila de mensagens do prototipo (Slide 3). So o
 * necessario para este demo: "once" (ouvinte de disparo unico por evento) e
 * "emit" (publica e segue em frente, sem bloquear quem publicou).
 *
 * <p>Adaptacao: em vez de um loop de eventos single-thread (Node/asyncio),
 * "emit" despacha o ouvinte para o {@link ExecutorService} do fluxo — o
 * equivalente idiomatico em Java de "nao-bloqueante" quando o runtime usa
 * threads reais em vez de um event loop.</p>
 */
public final class Barramento {

    private final Map<String, Consumer<DadoProtocolo>> ouvintesUnicos = new ConcurrentHashMap<>();
    private final ExecutorService executor;

    public Barramento(ExecutorService executor) {
        this.executor = executor;
    }

    public void once(String evento, Consumer<DadoProtocolo> ouvinte) {
        ouvintesUnicos.put(evento, ouvinte);
    }

    public void emit(String evento, DadoProtocolo dado) {
        Consumer<DadoProtocolo> ouvinte = ouvintesUnicos.remove(evento);
        if (ouvinte != null) {
            executor.submit(() -> ouvinte.accept(dado)); // nao-bloqueante: publica e segue em frente
        }
    }
}
