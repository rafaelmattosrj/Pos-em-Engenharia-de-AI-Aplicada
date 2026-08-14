package com.trialforge.gateway;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * BM25 classico (Hybrid Search, Modulo 4.1) — a metade lexical da fusao com
 * o embedding denso.
 */
public final class Bm25 {

    private Bm25() {
    }

    /** Estatisticas do corpus, pre-computadas uma vez na indexacao. */
    public record EstatisticasBM25(List<List<String>> tokensPorDoc, int n, double avgdl, Map<String, Integer> df) {
    }

    public static EstatisticasBM25 construirEstatisticasBM25(List<String> documentos) {
        List<List<String>> tokensPorDoc = new ArrayList<>();
        int somaComprimentos = 0;
        for (String doc : documentos) {
            List<String> tokens = Tokenizer.tokenizar(doc);
            tokensPorDoc.add(tokens);
            somaComprimentos += tokens.size();
        }
        int n = documentos.size();
        double avgdl = (double) somaComprimentos / n;

        Map<String, Integer> df = new HashMap<>();
        for (List<String> tokens : tokensPorDoc) {
            Set<String> vistos = new HashSet<>(tokens);
            for (String termo : vistos) {
                df.merge(termo, 1, Integer::sum);
            }
        }

        return new EstatisticasBM25(tokensPorDoc, n, avgdl, df);
    }

    public static double scoreBM25(List<String> queryTokens, List<String> docTokens, EstatisticasBM25 estatisticas) {
        return scoreBM25(queryTokens, docTokens, estatisticas, 1.5, 0.75);
    }

    public static double scoreBM25(List<String> queryTokens, List<String> docTokens, EstatisticasBM25 estatisticas,
            double k1, double b) {
        double score = 0;
        for (String termo : queryTokens) {
            long freqNoDoc = docTokens.stream().filter(t -> t.equals(termo)).count();
            if (freqNoDoc == 0) {
                continue;
            }
            int docFreq = estatisticas.df().getOrDefault(termo, 0);
            // +1 evita idf negativo com poucos docs
            double idf = Math.log((estatisticas.n() - docFreq + 0.5) / (docFreq + 0.5) + 1);
            double numerador = freqNoDoc * (k1 + 1);
            double denominador = freqNoDoc + k1 * (1 - b + (b * docTokens.size()) / estatisticas.avgdl());
            score += idf * (numerador / denominador);
        }
        return score;
    }
}
