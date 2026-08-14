package com.trialforge.gateway;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * RAG (Modulo 4.1): Multi-Index + Hybrid Search. Tres mecanismos empilhados:
 *   1. Multi-Index: busca so dentro do indice do dominio certo, nao no banco inteiro.
 *   2. Hybrid Search: combina lexico (BM25) com denso (embedding), fundidos por RANK,
 *      nunca somando os dois scores brutos direto — escalas incompativeis.
 * As buscas abaixo nunca fazem rede: todo embedding do CORPUS ja aconteceu em
 * prepararIndices() — buscar so compara vetores e tokens ja prontos.
 */
public final class RagSearch {

    private RagSearch() {
    }

    public static Map<String, IndicePreparado> prepararIndices(Embedder embedder, DemoLogger logger) throws Exception {
        int totalClausulas = Indices.NOMES_INDICES.stream()
                .mapToInt(nome -> Indices.INDICES.get(nome).size())
                .sum();
        if (logger != null) {
            logger.log(String.format(
                    "[Indexação] Pré-computando embeddings e estatísticas BM25 de %d cláusulas em %d índices — uma vez, não a cada busca...",
                    totalClausulas, Indices.NOMES_INDICES.size()));
        }

        Map<String, IndicePreparado> preparados = new LinkedHashMap<>();
        for (String nomeIndice : Indices.NOMES_INDICES) {
            List<Clausula> clausulas = Indices.INDICES.get(nomeIndice);
            List<String> textos = clausulas.stream().map(Clausula::texto).toList();
            Bm25.EstatisticasBM25 estatisticas = Bm25.construirEstatisticasBM25(textos);

            List<List<Double>> embeddingsTema = new ArrayList<>();
            List<List<Double>> embeddingsTexto = new ArrayList<>();
            for (Clausula clausula : clausulas) {
                embeddingsTema.add(embedder.embedar(clausula.tema()));
                embeddingsTexto.add(embedder.embedar(clausula.texto()));
            }

            preparados.put(nomeIndice, new IndicePreparado(clausulas, estatisticas, embeddingsTema, embeddingsTexto));
        }

        if (logger != null) {
            logger.log("[Indexação] Concluída.\n");
        }
        return preparados;
    }

    /**
     * Hybrid Search dentro de UM indice (ja roteado pelo Multi-Index).
     * {@code campoEmbedding} escolhe, entre os embeddings JA pre-computados,
     * comparar contra o tema (mais preciso) ou o texto inteiro da clausula
     * (mais abrangente) — o que o Agentic RAG varia entre tentativas.
     */
    public static ResultadoBusca buscarClausulaHibrida(Map<String, IndicePreparado> preparados, String pergunta,
            List<Double> perguntaEmbedding, String nomeIndice, String campoEmbedding) {
        IndicePreparado indice = preparados.get(nomeIndice);
        List<List<Double>> embeddingsCorpus = "texto".equals(campoEmbedding) ? indice.embeddingsTexto() : indice.embeddingsTema();
        List<String> queryTokens = Tokenizer.tokenizar(pergunta);

        List<Double> similaridades = new ArrayList<>();
        for (List<Double> embeddingClausula : embeddingsCorpus) {
            similaridades.add(CosineSimilarity.similaridadeCosseno(perguntaEmbedding, embeddingClausula));
        }
        List<Integer> rankingDenso = RankUtils.ordenarPorScore(similaridades);

        List<Double> scoresBM25 = new ArrayList<>();
        for (int idx = 0; idx < indice.clausulas().size(); idx++) {
            scoresBM25.add(Bm25.scoreBM25(queryTokens, indice.estatisticas().tokensPorDoc().get(idx), indice.estatisticas()));
        }
        List<Integer> rankingEsparso = RankUtils.ordenarPorScore(scoresBM25);

        List<RankUtils.RankedScore> fusao = RankUtils.fusaoReciprocalRank(rankingDenso, rankingEsparso, 60);
        RankUtils.RankedScore melhor = fusao.get(0);

        return ResultadoBusca.bruto(
                nomeIndice,
                indice.clausulas().get(melhor.idx()),
                similaridades.get(melhor.idx()),
                scoresBM25.get(melhor.idx()),
                melhor.score());
    }

    /** Ultimo recurso do Agentic RAG: cruza TODOS os indices, nao so o roteado inicialmente. */
    public static ResultadoBusca buscarEmTodosIndices(Map<String, IndicePreparado> preparados, String pergunta,
            List<Double> perguntaEmbedding) {
        ResultadoBusca melhor = null;
        for (String nomeIndice : Indices.NOMES_INDICES) {
            ResultadoBusca resultado = buscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndice, "texto");
            if (melhor == null || resultado.similaridadeCosseno() > melhor.similaridadeCosseno()) {
                melhor = resultado;
            }
        }
        return melhor;
    }
}
