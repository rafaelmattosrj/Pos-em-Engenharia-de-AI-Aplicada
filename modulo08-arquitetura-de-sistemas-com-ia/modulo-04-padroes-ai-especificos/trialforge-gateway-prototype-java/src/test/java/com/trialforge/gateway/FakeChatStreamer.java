package com.trialforge.gateway;

import java.util.List;
import java.util.function.Consumer;

/** Devolve uma resposta fixa (emitida num único chunk) e registra o modelo
 * com que foi chamado. */
class FakeChatStreamer implements ChatStreamer {

    private final String resposta;
    int chamadas = 0;
    String ultimoModelo;

    FakeChatStreamer(String resposta) {
        this.resposta = resposta;
    }

    @Override
    public void chatStream(String model, List<ChatMessage> messages, Consumer<String> onChunk) {
        chamadas++;
        ultimoModelo = model;
        onChunk.accept(resposta);
    }
}
