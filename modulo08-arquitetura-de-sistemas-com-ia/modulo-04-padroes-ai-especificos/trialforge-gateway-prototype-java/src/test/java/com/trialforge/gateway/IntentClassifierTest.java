package com.trialforge.gateway;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.assertj.core.api.Assertions.assertThat;

/** Mesmos 7 casos usados pelos testes puros dos dois originais. */
class IntentClassifierTest {

    @ParameterizedTest
    @CsvSource(delimiterString = "|", value = {
            "Preciso da síntese do CSR final desse estudo.|sintese_csr",
            "Quero o relatório final do estudo.|sintese_csr",
            "Como os eventos adversos aparecem no relatório final?|sintese_csr",
            "Qual é o critério de idade mínima pra participar desse estudo?|consulta_protocolo",
            "Quais são os critérios de exclusão desse protocolo?|consulta_protocolo",
            "Quais são as regras de assentimento pra menores?|consulta_icf",
            "Qual o prazo de armazenamento das amostras biológicas?|consulta_icf",
    })
    void classificarIntencao(String pergunta, String esperado) {
        assertThat(IntentClassifier.classificarIntencao(pergunta)).isEqualTo(esperado);
    }

    @Test
    void classificarIntencaoIgnoraCaixa() {
        assertThat(IntentClassifier.classificarIntencao("SÍNTESE do CSR")).isEqualTo("sintese_csr");
    }
}
