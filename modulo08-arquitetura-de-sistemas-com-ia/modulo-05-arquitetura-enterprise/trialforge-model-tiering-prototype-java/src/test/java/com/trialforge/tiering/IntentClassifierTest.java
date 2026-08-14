package com.trialforge.tiering;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class IntentClassifierTest {

    @Test
    void reconheceSinteseCsrPorPalavraCsr() {
        assertThat(IntentClassifier.classificarIntencao("Preciso da síntese do CSR final desse estudo."))
                .isEqualTo(IntentClassifier.SINTESE_CSR);
    }

    @Test
    void reconheceSinteseCsrPorRelatorioFinal() {
        assertThat(IntentClassifier.classificarIntencao("Quero o relatório final do estudo."))
                .isEqualTo(IntentClassifier.SINTESE_CSR);
    }

    @Test
    void reconheceConsultaClausulaComoPadrao() {
        assertThat(IntentClassifier.classificarIntencao("Quais são as regras de assentimento pra menores?"))
                .isEqualTo(IntentClassifier.CONSULTA_CLAUSULA);
    }
}
