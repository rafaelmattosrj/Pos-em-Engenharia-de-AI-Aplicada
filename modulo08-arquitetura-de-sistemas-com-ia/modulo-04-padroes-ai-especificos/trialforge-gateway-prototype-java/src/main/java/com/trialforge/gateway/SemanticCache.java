package com.trialforge.gateway;

import java.util.ArrayList;
import java.util.List;

/**
 * Semantic Cache (Modulo 4.3) — em memoria, populado so com respostas de
 * perguntas de rotina (sintese de CSR nunca usa cache).
 */
public class SemanticCache {

    /** Item do cache: pergunta original, seu embedding e a resposta ja aprovada. */
    public record Entrada(String pergunta, List<Double> embedding, String resposta) {
    }

    public record ResultadoConsulta(double similaridade, Entrada entrada) {
    }

    private final List<Entrada> entradas = new ArrayList<>();

    /** Percorre o cache inteiro e devolve a entrada de maior similaridade de
     * cosseno em relacao ao embedding da pergunta atual. */
    public ResultadoConsulta consultar(List<Double> perguntaEmbedding) {
        double melhorSimilaridade = 0;
        Entrada melhorEntrada = null;
        for (Entrada entrada : entradas) {
            double sim = CosineSimilarity.similaridadeCosseno(perguntaEmbedding, entrada.embedding());
            if (sim > melhorSimilaridade) {
                melhorSimilaridade = sim;
                melhorEntrada = entrada;
            }
        }
        return new ResultadoConsulta(melhorSimilaridade, melhorEntrada);
    }

    /** Alimenta o cache com essa pergunta+resposta pra proxima vez. */
    public void adicionar(String pergunta, List<Double> embedding, String resposta) {
        entradas.add(new Entrada(pergunta, embedding, resposta));
    }

    public int tamanho() {
        return entradas.size();
    }
}
