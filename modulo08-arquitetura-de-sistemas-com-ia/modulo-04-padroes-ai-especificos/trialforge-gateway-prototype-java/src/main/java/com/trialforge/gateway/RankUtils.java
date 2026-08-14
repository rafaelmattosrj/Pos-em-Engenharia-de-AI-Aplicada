package com.trialforge.gateway;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Ordenacao por score e Reciprocal Rank Fusion (Cormack, Clarke & Buttcher,
 * SIGIR 2009) — funde dois rankings pela POSICAO de cada documento em cada
 * um, nao pelo valor do score, pra combinar BM25 (escala aberta) com cosseno
 * (-1 a 1) sem normalizar nada a mao.
 */
public final class RankUtils {

    private RankUtils() {
    }

    public record RankedScore(int idx, double score) {
    }

    /** Devolve os indices dos documentos ordenados por score decrescente
     * (ordenacao estavel — em empate, mantem a ordem original). */
    public static List<Integer> ordenarPorScore(List<Double> scores) {
        List<Integer> indices = new ArrayList<>();
        for (int i = 0; i < scores.size(); i++) {
            indices.add(i);
        }
        indices.sort(Comparator.comparingDouble((Integer idx) -> scores.get(idx)).reversed());
        return indices;
    }

    public static List<RankedScore> fusaoReciprocalRank(List<Integer> ranking1, List<Integer> ranking2, int k) {
        Map<Integer, Double> scores = new LinkedHashMap<>();
        for (List<Integer> ranking : List.of(ranking1, ranking2)) {
            for (int posicao = 0; posicao < ranking.size(); posicao++) {
                int idx = ranking.get(posicao);
                scores.merge(idx, 1.0 / (k + posicao + 1), Double::sum);
            }
        }

        List<RankedScore> resultado = new ArrayList<>();
        for (Map.Entry<Integer, Double> entry : scores.entrySet()) {
            resultado.add(new RankedScore(entry.getKey(), entry.getValue()));
        }
        resultado.sort(Comparator.comparingDouble(RankedScore::score).reversed());
        return resultado;
    }
}
