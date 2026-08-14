package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class Bm25Test {

    // Mesmo corpus/query dos testes puros dos dois originais: o doc que
    // compartilha "idade" e "minima" com a query deve vencer a busca lexical.
    @Test
    void docComTermosDaQueryVenceABuscaLexical() {
        List<String> corpus = List.of(
                "idade mínima de doze anos para participar do estudo",
                "consentimento do responsável legal é obrigatório",
                "retirada do participante a qualquer momento sem justificativa");
        Bm25.EstatisticasBM25 estatisticas = Bm25.construirEstatisticasBM25(corpus);
        List<String> query = Tokenizer.tokenizar("qual a idade mínima exigida");

        double melhorScore = -1;
        int vencedor = -1;
        for (int idx = 0; idx < corpus.size(); idx++) {
            double score = Bm25.scoreBM25(query, estatisticas.tokensPorDoc().get(idx), estatisticas);
            if (score > melhorScore) {
                melhorScore = score;
                vencedor = idx;
            }
        }

        assertThat(vencedor).isZero();
    }

    @Test
    void semTermoEmComumScoreEhZero() {
        List<String> corpus = List.of("gatos e cachorros", "carros e motos");
        Bm25.EstatisticasBM25 estatisticas = Bm25.construirEstatisticasBM25(corpus);
        List<String> query = Tokenizer.tokenizar("nenhumtermoemcomum");

        for (int idx = 0; idx < corpus.size(); idx++) {
            double score = Bm25.scoreBM25(query, estatisticas.tokensPorDoc().get(idx), estatisticas);
            assertThat(score).isZero();
        }
    }

    @Test
    void construirEstatisticasBM25CalculaAvgdlEDf() {
        Bm25.EstatisticasBM25 estatisticas = Bm25.construirEstatisticasBM25(List.of("a b c", "a b"));

        assertThat(estatisticas.n()).isEqualTo(2);
        assertThat(estatisticas.avgdl()).isEqualTo(2.5);
        assertThat(estatisticas.df().get("a")).isEqualTo(2);
        assertThat(estatisticas.df().get("c")).isEqualTo(1);
    }
}
