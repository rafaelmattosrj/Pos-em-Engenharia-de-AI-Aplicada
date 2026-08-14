package com.trialforge.gateway;

import java.util.List;

/**
 * Resultado da indexacao (Modulo 4.1) de UM dominio: embeddings de cada
 * clausula (tema e texto) e as estatisticas BM25 do corpus — tudo calculado
 * uma vez, na inicializacao.
 */
public record IndicePreparado(
        List<Clausula> clausulas,
        Bm25.EstatisticasBM25 estatisticas,
        List<List<Double>> embeddingsTema,
        List<List<Double>> embeddingsTexto) {
}
