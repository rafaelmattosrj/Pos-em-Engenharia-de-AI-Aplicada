package com.trialforge.gateway;

import java.util.List;

/**
 * Base do Semantic Cache e do score de confianca da busca de clausula do
 * RAG — usa embeddings reais, nao heuristica de palavra-chave.
 */
public final class CosineSimilarity {

    private CosineSimilarity() {
    }

    public static double similaridadeCosseno(List<Double> a, List<Double> b) {
        double produto = 0;
        double normaA = 0;
        double normaB = 0;
        for (int i = 0; i < a.size(); i++) {
            produto += a.get(i) * b.get(i);
            normaA += a.get(i) * a.get(i);
            normaB += b.get(i) * b.get(i);
        }
        return produto / (Math.sqrt(normaA) * Math.sqrt(normaB));
    }
}
