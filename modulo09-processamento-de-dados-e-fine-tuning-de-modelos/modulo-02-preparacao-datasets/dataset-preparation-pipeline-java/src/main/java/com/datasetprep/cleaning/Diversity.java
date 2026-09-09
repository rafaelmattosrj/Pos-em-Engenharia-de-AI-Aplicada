package com.datasetprep.cleaning;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Porte da secao 4 de dataset-cleaning-balancing-tool.js: entropia de
 * Shannon da distribuicao por fonte, e numero efetivo de fontes
 * (Hill number de ordem 1 = exp(H); Hill 1973, Jost 2006).
 */
public final class Diversity {

    private Diversity() {
    }

    public static Map<String, Double> distribuicaoDe(Map<String, Integer> contagens) {
        int total = contagens.values().stream().mapToInt(Integer::intValue).sum();
        Map<String, Double> dist = new LinkedHashMap<>();
        for (Map.Entry<String, Integer> e : contagens.entrySet()) {
            dist.put(e.getKey(), total == 0 ? 0.0 : (double) e.getValue() / total);
        }
        return dist;
    }

    /** Entropia de Shannon em nats (log natural). H=0 significa uma unica fonte dominando tudo. */
    public static double entropiaShannon(Map<String, Double> distribuicao) {
        double h = 0;
        for (double p : distribuicao.values()) {
            if (p > 0) h += p * Math.log(p);
        }
        return -h;
    }

    /** Numero efetivo de fontes (Hill number de ordem 1) = exp(H). */
    public static double numeroEfetivoFontes(Map<String, Double> distribuicao) {
        return Math.exp(entropiaShannon(distribuicao));
    }
}
