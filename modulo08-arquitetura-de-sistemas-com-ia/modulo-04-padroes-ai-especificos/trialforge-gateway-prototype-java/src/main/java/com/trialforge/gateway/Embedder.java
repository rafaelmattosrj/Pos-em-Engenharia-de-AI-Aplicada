package com.trialforge.gateway;

import java.util.List;

/**
 * Unico contrato de rede que o RAG precisa: transformar um texto num vetor
 * de embedding. Implementado por OllamaClient — a busca em si
 * (RagSearch.buscarClausulaHibrida) nunca faz rede, so compara vetores ja
 * prontos.
 */
@FunctionalInterface
public interface Embedder {
    List<Double> embedar(String texto) throws Exception;
}
