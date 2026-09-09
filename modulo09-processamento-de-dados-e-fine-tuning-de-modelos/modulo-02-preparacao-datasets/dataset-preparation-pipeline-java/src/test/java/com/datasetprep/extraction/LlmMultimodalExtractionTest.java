package com.datasetprep.extraction;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class LlmMultimodalExtractionTest {

    @Test
    void normalizarRemoveAcentoCaixaEEspacosExtras() {
        assertThat(LlmMultimodalExtraction.normalizar("  José  DA Silva ")).isEqualTo("jose da silva");
    }

    @Test
    void compararComEsperadoBateParaValorNumericoEquivalente() {
        var extraido = Map.<String, Object>of("segurado", "Marcos Vinicius Andrade Pereira", "placa", "QJK-4F82", "valor", 3210.50);
        var esperado = Map.<String, Object>of("segurado", "Marcos Vinicius Andrade Pereira", "placa", "QJK-4F82", "valor", 3210.50);
        var resultado = LlmMultimodalExtraction.compararComEsperado(extraido, esperado);
        assertThat(resultado.get("acertos")).isEqualTo(3);
        assertThat(resultado.get("total")).isEqualTo(3);
    }

    @Test
    void compararComEsperadoToleraDiferencaDeAcentoECaixaEmString() {
        var extraido = Map.<String, Object>of("beneficiario", "carlos eduardo martins");
        var esperado = Map.<String, Object>of("beneficiario", "Carlos Eduardo Martins");
        var resultado = LlmMultimodalExtraction.compararComEsperado(extraido, esperado);
        assertThat(resultado.get("acertos")).isEqualTo(1);
    }

    @Test
    void compararComEsperadoDetectaDivergenciaDeValor() {
        var extraido = Map.<String, Object>of("valor", 100.0);
        var esperado = Map.<String, Object>of("valor", 200.0);
        var resultado = LlmMultimodalExtraction.compararComEsperado(extraido, esperado);
        assertThat(resultado.get("acertos")).isEqualTo(0);
    }
}
