package com.trialforge.tiering;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.offset;

class VectorMathTest {

    @Test
    void vetoresIdenticos_dá1() {
        assertThat(VectorMath.similaridadeCosseno(new double[]{1, 0, 0}, new double[]{1, 0, 0}))
                .isCloseTo(1.0, offset(1e-9));
    }

    @Test
    void vetoresOrtogonais_dá0() {
        assertThat(VectorMath.similaridadeCosseno(new double[]{1, 0, 0}, new double[]{0, 1, 0}))
                .isCloseTo(0.0, offset(1e-9));
    }
}
