package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.within;

class CosineSimilarityTest {

    private static final List<Double> VETOR_A = List.of(1.0, 0.0, 0.0);
    private static final List<Double> VETOR_B = List.of(1.0, 0.0, 0.0);
    private static final List<Double> VETOR_ORTOGONAL = List.of(0.0, 1.0, 0.0);
    private static final List<Double> VETOR_OPOSTO = List.of(-1.0, 0.0, 0.0);

    @Test
    void vetoresIdenticosDaoSimilaridadeUm() {
        assertThat(CosineSimilarity.similaridadeCosseno(VETOR_A, VETOR_B)).isCloseTo(1.0, within(1e-9));
    }

    @Test
    void vetoresOrtogonaisDaoSimilaridadeZero() {
        assertThat(CosineSimilarity.similaridadeCosseno(VETOR_A, VETOR_ORTOGONAL)).isCloseTo(0.0, within(1e-9));
    }

    @Test
    void vetoresOpostosDaoSimilaridadeMenosUm() {
        assertThat(CosineSimilarity.similaridadeCosseno(VETOR_A, VETOR_OPOSTO)).isCloseTo(-1.0, within(1e-9));
    }
}
