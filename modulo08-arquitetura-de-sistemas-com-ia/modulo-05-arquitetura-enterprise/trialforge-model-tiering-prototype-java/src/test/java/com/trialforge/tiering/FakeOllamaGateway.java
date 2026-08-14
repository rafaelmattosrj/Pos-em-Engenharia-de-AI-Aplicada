package com.trialforge.tiering;

import java.util.HashMap;
import java.util.Map;
import java.util.function.Consumer;

/**
 * Dublê determinístico de {@link OllamaGateway}: embeddings e respostas de
 * chat pré-configurados por texto — sem chamar nenhum Ollama de verdade, e
 * sem latência de rede (essencial para os testes de concorrência, que disparam
 * dezenas de chamadas "simultâneas").
 */
class FakeOllamaGateway implements OllamaGateway {

    private final Map<String, double[]> embeddings = new HashMap<>();
    private final Map<String, String> chats = new HashMap<>();
    private String chatPadrao = "resposta padrão";
    private double[] embeddingPadrao = {0, 0};

    FakeOllamaGateway comEmbedding(String texto, double[] vetor) {
        embeddings.put(texto, vetor);
        return this;
    }

    FakeOllamaGateway comEmbeddingPadrao(double[] vetor) {
        this.embeddingPadrao = vetor;
        return this;
    }

    /** Resposta de chat, roteada por modelo + trecho contido no prompt do usuário. */
    FakeOllamaGateway comChat(String modelo, String trechoDoPrompt, String resposta) {
        chats.put(modelo + "|" + trechoDoPrompt, resposta);
        return this;
    }

    FakeOllamaGateway comChatPadrao(String resposta) {
        this.chatPadrao = resposta;
        return this;
    }

    @Override
    public String chatStream(String modelo, String systemPrompt, String userPrompt, Consumer<String> onChunk) {
        String resposta = chatPadrao;
        for (Map.Entry<String, String> entry : chats.entrySet()) {
            String[] chaveModeloTrecho = entry.getKey().split("\\|", 2);
            if (chaveModeloTrecho[0].equals(modelo) && userPrompt.contains(chaveModeloTrecho[1])) {
                resposta = entry.getValue();
                break;
            }
        }
        onChunk.accept(resposta);
        return resposta;
    }

    @Override
    public double[] embed(String modelo, String texto) {
        return embeddings.getOrDefault(texto, embeddingPadrao);
    }
}
