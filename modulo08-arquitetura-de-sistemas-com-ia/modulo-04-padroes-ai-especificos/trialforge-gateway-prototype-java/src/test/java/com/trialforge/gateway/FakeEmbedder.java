package com.trialforge.gateway;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Devolve um vetor fixo por texto exato — permite controlar precisamente
 * similaridade de cosseno em cada cenário de teste, sem depender de um
 * Ollama real. */
class FakeEmbedder implements Embedder {

    private final Map<String, List<Double>> vetores = new HashMap<>();
    private List<Double> padrao = List.of(-1.0, -1.0, -1.0);
    int chamadas = 0;

    FakeEmbedder com(String texto, List<Double> vetor) {
        vetores.put(texto, vetor);
        return this;
    }

    FakeEmbedder comPadrao(List<Double> vetor) {
        this.padrao = vetor;
        return this;
    }

    @Override
    public List<Double> embedar(String texto) {
        chamadas++;
        return vetores.getOrDefault(texto, padrao);
    }
}
