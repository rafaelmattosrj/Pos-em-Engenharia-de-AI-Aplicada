package com.trialforge.gateway;

import java.util.List;
import java.util.function.Consumer;

/**
 * Segundo (e unico outro) contrato de rede do gateway: gerar uma resposta em
 * streaming, token a token — onChunk e chamado uma vez por pedaco de texto
 * recebido, na ordem em que chegam.
 */
@FunctionalInterface
public interface ChatStreamer {
    void chatStream(String model, List<ChatMessage> messages, Consumer<String> onChunk) throws Exception;
}
