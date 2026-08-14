package com.trialforge.evalgate;

import java.util.HashMap;
import java.util.Map;

/**
 * Dublê determinístico de {@link OllamaGateway}: devolve respostas de chat e
 * embeddings pré-configurados por texto, sem chamar nenhum Ollama de verdade.
 * Embeddings são vetores simples (uma dimensão por "conceito"), suficientes
 * para exercitar a matemática de similaridade de cosseno de ponta a ponta.
 */
class FakeOllamaGateway implements OllamaGateway {

    private final Map<String, String> chatPorPergunta = new HashMap<>();
    private final Map<String, double[]> embeddingsPorTexto = new HashMap<>();
    private String chatPadrao = "resposta padrão";

    FakeOllamaGateway comChat(String pergunta, String resposta) {
        chatPorPergunta.put(pergunta, resposta);
        return this;
    }

    FakeOllamaGateway comChatPadrao(String resposta) {
        this.chatPadrao = resposta;
        return this;
    }

    FakeOllamaGateway comEmbedding(String texto, double[] vetor) {
        embeddingsPorTexto.put(texto, vetor);
        return this;
    }

    @Override
    public String chat(String modelo, String systemPrompt, String userPrompt) {
        for (Map.Entry<String, String> entry : chatPorPergunta.entrySet()) {
            if (userPrompt.contains(entry.getKey())) {
                return entry.getValue();
            }
        }
        return chatPadrao;
    }

    @Override
    public double[] embed(String modelo, String texto) {
        double[] vetor = embeddingsPorTexto.get(texto);
        if (vetor != null) {
            return vetor;
        }
        // fallback determinístico: hash simples do texto vira vetor 2D
        return new double[]{texto.length(), texto.hashCode() % 97};
    }
}
