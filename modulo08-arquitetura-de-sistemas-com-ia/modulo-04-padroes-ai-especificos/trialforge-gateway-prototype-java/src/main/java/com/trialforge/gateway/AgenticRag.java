package com.trialforge.gateway;

import java.util.List;
import java.util.Map;

/**
 * Agentic RAG (Modulo 4.1): busca -> avalia confianca -> busca de novo com
 * estrategia mais ampla, ate MAX_ITERACOES_AGENTIC. Se nunca atingir o
 * limiar, devolve o melhor achado e deixa o Confidence Threshold (Modulo
 * 4.4) escalar pro Approval Gate — o RAG nao decide sozinho quando desistir,
 * ele so para de insistir e passa a decisao adiante.
 */
public final class AgenticRag {

    // mesmo limite do retry com limite do Modulo 3.5: insistir alem disso nao ajuda mais
    public static final int MAX_ITERACOES_AGENTIC = 3;

    private AgenticRag() {
    }

    public static ResultadoBusca buscarClausulaAgentica(Map<String, IndicePreparado> preparados, String pergunta,
            List<Double> perguntaEmbedding, String nomeIndiceInicial, double limiarConfianca, DemoLogger logger) {
        ResultadoBusca melhorResultado = null;

        for (int iteracao = 1; iteracao <= MAX_ITERACOES_AGENTIC; iteracao++) {
            ResultadoBusca resultado;
            String estrategia;
            if (iteracao == 1) {
                estrategia = String.format("índice \"%s\", comparando com o tema da cláusula", nomeIndiceInicial);
                resultado = RagSearch.buscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndiceInicial, "tema");
            } else if (iteracao == 2) {
                estrategia = String.format("índice \"%s\", ampliando pro texto completo da cláusula", nomeIndiceInicial);
                resultado = RagSearch.buscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndiceInicial, "texto");
            } else {
                estrategia = "todos os índices, cruzando domínios (último recurso)";
                resultado = RagSearch.buscarEmTodosIndices(preparados, pergunta, perguntaEmbedding);
            }

            if (logger != null) {
                logger.log(String.format("  [Agentic RAG] Iteração %d/%d — %s — confiança %.3f",
                        iteracao, MAX_ITERACOES_AGENTIC, estrategia, resultado.similaridadeCosseno()));
            }
            if (melhorResultado == null || resultado.similaridadeCosseno() > melhorResultado.similaridadeCosseno()) {
                melhorResultado = resultado;
            }

            if (resultado.similaridadeCosseno() >= limiarConfianca) {
                return melhorResultado.comIteracao(iteracao, false);
            }
        }

        if (logger != null) {
            logger.log(String.format(
                    "  [Agentic RAG] %d iterações esgotadas sem atingir o limiar — segue com o melhor resultado (confiança %.3f) e deixa o Confidence Threshold decidir.",
                    MAX_ITERACOES_AGENTIC, melhorResultado.similaridadeCosseno()));
        }
        return melhorResultado.comIteracao(MAX_ITERACOES_AGENTIC, true);
    }
}
