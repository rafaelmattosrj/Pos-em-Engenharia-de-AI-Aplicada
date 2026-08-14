package com.trialforge.guardrail;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Dublê determinístico de {@link ClassifierClient} para testes: devolve uma
 * classificação fixa por pergunta, sem chamar nenhum Ollama de verdade.
 */
class FakeClassifierClient implements ClassifierClient {

    private final Map<String, String> respostasPorPergunta = new LinkedHashMap<>();
    private String respostaPadrao = "legitima";
    int chamadas = 0;

    FakeClassifierClient comResposta(String pergunta, String classificacaoBruta) {
        respostasPorPergunta.put(pergunta, classificacaoBruta);
        return this;
    }

    FakeClassifierClient comRespostaPadrao(String classificacaoBruta) {
        this.respostaPadrao = classificacaoBruta;
        return this;
    }

    @Override
    public String classify(String pergunta) {
        chamadas++;
        return respostasPorPergunta.getOrDefault(pergunta, respostaPadrao);
    }
}
