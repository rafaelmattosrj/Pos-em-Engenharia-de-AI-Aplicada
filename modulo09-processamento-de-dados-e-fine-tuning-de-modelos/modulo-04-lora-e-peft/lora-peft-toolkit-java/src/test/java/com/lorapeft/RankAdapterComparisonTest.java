package com.lorapeft;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class RankAdapterComparisonTest {

    private static final String SAIDA_SEM_ADAPTADOR = """
            ==========
            <|channel>thought
            Here's a thinking process to extract the requested information:

            1.  **Analyze the Request:** The user wants to extract three specific pieces of information from the provided text (an auto repair quote/budget):
            ==========
            Prompt: 155 tokens, 211.031 tokens-per-sec
            Generation: 80 tokens, 50.460 tokens-per-sec
            Peak memory: 9.447 GB""";

    private static final String SAIDA_RANK4 = """
            ==========
            {"segurado":"Ricardo Alves Monteiro","placa":"JBR-9021","valor":2820}
            ==========
            Prompt: 155 tokens, 218.733 tokens-per-sec
            Generation: 27 tokens, 41.540 tokens-per-sec
            Peak memory: 9.447 GB""";

    @Test
    void semAdaptadorBateNoLimiteDeTokensENaoChegaAJson() {
        RankAdapterComparison.ResultadoParse r = RankAdapterComparison.parsearSaida(SAIDA_SEM_ADAPTADOR);
        assertThat(r.tokens()).isEqualTo(80);
        assertThat(r.json()).isNull();
    }

    @Test
    void rank4JsonExatoBatendoComOGabarito() {
        RankAdapterComparison.ResultadoParse r = RankAdapterComparison.parsearSaida(SAIDA_RANK4);
        assertThat(r.json().get("segurado").getAsString()).isEqualTo("Ricardo Alves Monteiro");
        assertThat(r.json().get("placa").getAsString()).isEqualTo("JBR-9021");
        assertThat(r.json().get("valor").getAsInt()).isEqualTo(2820);
    }

    @Test
    void baterComGabaritoConfirmaERejeitaCorretamente() {
        RankAdapterComparison.ResultadoParse r = RankAdapterComparison.parsearSaida(SAIDA_RANK4);
        assertThat(RankAdapterComparison.baterComGabarito(r.json())).isTrue();
    }

    @Test
    void adaptersRetornaOsTresRanksComCaminhosCorretos() {
        var adapters = RankAdapterComparison.adapters("/base");
        assertThat(adapters).containsEntry("rank 4", "/base/mlx-adapters-rank4");
        assertThat(adapters).containsEntry("rank 8", "/base/mlx-adapters");
        assertThat(adapters).containsEntry("rank 16", "/base/mlx-adapters-rank16");
    }

    @Test
    void montaArgumentosSemAdapterPathQuandoNulo() {
        assertThat(RankAdapterComparison.montarArgumentos(null)).doesNotContain("--adapter-path");
    }
}
