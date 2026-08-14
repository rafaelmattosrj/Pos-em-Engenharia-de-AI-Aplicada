package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class RankUtilsTest {

    @Test
    void ordenarPorScoreDevolveIndicesDecrescentes() {
        List<Integer> resultado = RankUtils.ordenarPorScore(List.of(0.2, 0.9, 0.5));
        assertThat(resultado).containsExactly(1, 2, 0);
    }

    // Mesmo caso dos testes puros: ranking1 e ranking2 concordam que o doc 0
    // é o melhor — sem empate, resultado não depende de estabilidade de sort.
    @Test
    void fusaoReciprocalRankDocNoTopoDosDoisVence() {
        List<RankUtils.RankedScore> fusao = RankUtils.fusaoReciprocalRank(List.of(0, 1, 2), List.of(0, 2, 1), 60);
        assertThat(fusao.get(0).idx()).isZero();
    }

    @Test
    void fusaoReciprocalRankIncluiTodosOsDocumentos() {
        List<RankUtils.RankedScore> fusao = RankUtils.fusaoReciprocalRank(List.of(0, 1), List.of(2, 3), 60);
        assertThat(fusao).hasSize(4);
    }
}
