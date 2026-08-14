package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class SemanticCacheTest {

    @Test
    void consultarCacheVazioDevolveSimilaridadeZero() {
        SemanticCache cache = new SemanticCache();
        SemanticCache.ResultadoConsulta resultado = cache.consultar(List.of(1.0, 0.0, 0.0));
        assertThat(resultado.entrada()).isNull();
        assertThat(resultado.similaridade()).isZero();
    }

    @Test
    void hitDevolveEntradaMaisSimilar() {
        SemanticCache cache = new SemanticCache();
        cache.adicionar("pergunta original", List.of(1.0, 0.0, 0.0), "resposta original");

        SemanticCache.ResultadoConsulta resultado = cache.consultar(List.of(1.0, 0.0, 0.0));

        assertThat(resultado.entrada()).isNotNull();
        assertThat(resultado.similaridade()).isEqualTo(1.0);
        assertThat(resultado.entrada().resposta()).isEqualTo("resposta original");
    }

    @Test
    void escolheEntradaDeMaiorSimilaridade() {
        SemanticCache cache = new SemanticCache();
        cache.adicionar("pergunta distante", List.of(0.0, 1.0, 0.0), "resposta distante");
        cache.adicionar("pergunta próxima", List.of(1.0, 0.0, 0.0), "resposta próxima");

        SemanticCache.ResultadoConsulta resultado = cache.consultar(List.of(1.0, 0.0, 0.0));

        assertThat(resultado.entrada().resposta()).isEqualTo("resposta próxima");
    }
}
