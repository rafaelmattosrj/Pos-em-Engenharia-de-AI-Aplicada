package com.example;

import org.junit.jupiter.api.Test;

import java.util.List;

import static com.example.RagContextBuilder.Trecho;
import static org.assertj.core.api.Assertions.assertThat;

class RagContextBuilderTest {

    @Test
    void build_filtraTrechosAbaixoDoScoreMinimo() {
        List<Trecho> matches = List.of(
                new Trecho("trecho relevante", 0.8),
                new Trecho("trecho irrelevante", 0.3)
        );

        String contexto = RagContextBuilder.build(matches);

        assertThat(contexto).isEqualTo("trecho relevante");
    }

    @Test
    void build_scoreExatamenteNoLimiteNaoEIncluido() {
        List<Trecho> matches = List.of(new Trecho("no limite", RagContextBuilder.MIN_SCORE));

        String contexto = RagContextBuilder.build(matches);

        assertThat(contexto).isEmpty();
    }

    @Test
    void build_juntaMultiplosTrechosComSeparador() {
        List<Trecho> matches = List.of(
                new Trecho("primeiro", 0.9),
                new Trecho("segundo", 0.7)
        );

        String contexto = RagContextBuilder.build(matches);

        assertThat(contexto).isEqualTo("primeiro" + RagContextBuilder.SEPARATOR + "segundo");
    }

    @Test
    void build_semMatchesRetornaStringVazia() {
        String contexto = RagContextBuilder.build(List.of());

        assertThat(contexto).isEmpty();
    }

    @Test
    void build_todosOsMatchesAbaixoDoScoreRetornaVazio() {
        List<Trecho> matches = List.of(
                new Trecho("a", 0.1),
                new Trecho("b", 0.49)
        );

        String contexto = RagContextBuilder.build(matches);

        assertThat(contexto).isEmpty();
    }
}
