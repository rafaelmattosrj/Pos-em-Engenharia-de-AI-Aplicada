package com.trialforge.tiering;

import java.io.IOException;
import java.util.function.Consumer;

/**
 * Abstração das duas operações do Ollama usadas neste protótipo: chat (com
 * streaming, já que o original usa {@code stream: true} em gerarComTier) e
 * embeddings. Extraída como interface para permitir testar
 * {@link CascadeGateway} com um dublê determinístico, sem depender de um
 * Ollama local rodando durante `mvn test`.
 */
public interface OllamaGateway {

    /**
     * Gera uma resposta em streaming: cada pedaço de texto recebido é
     * repassado a {@code onChunk} (equivalente a imprimir cada parte no
     * console, como no original), e o texto completo é devolvido no final.
     */
    String chatStream(String modelo, String systemPrompt, String userPrompt, Consumer<String> onChunk)
            throws IOException, InterruptedException;

    double[] embed(String modelo, String texto) throws IOException, InterruptedException;
}
