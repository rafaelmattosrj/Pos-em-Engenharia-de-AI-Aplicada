package com.finetuningapi;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class OcrConfidenceGateTest {

    @Test
    void os4DocumentosReaisPassamNoLimiarPadrao() {
        var exemplos = List.of(
                new OcrConfidenceGate.ExemploComOcr("doc-auto-1", 0.943),
                new OcrConfidenceGate.ExemploComOcr("doc-auto-2", 0.958),
                new OcrConfidenceGate.ExemploComOcr("doc-saude-1", 0.957),
                new OcrConfidenceGate.ExemploComOcr("doc-saude-2", 0.959));
        var resultado = OcrConfidenceGate.filtrar(exemplos);
        assertThat(resultado.aprovadosPorOcr()).hasSize(4);
        assertThat(resultado.sinalizadosParaRevisao()).isEmpty();
    }

    @Test
    void exemploAbaixoDoLimiarESinalizadoParaRevisao() {
        var resultado = OcrConfidenceGate.filtrar(List.of(new OcrConfidenceGate.ExemploComOcr("doc-degradado", 0.62)));
        assertThat(resultado.sinalizadosParaRevisao()).hasSize(1);
        assertThat(resultado.aprovados()).isEmpty();
    }

    @Test
    void exemploSemConfiancaOcrSeguirAprovadoSemGate() {
        var sintetico = new OcrConfidenceGate.ExemploComOcr("amplitude-auto-Oficina Estrela-5", null);
        var resultado = OcrConfidenceGate.filtrar(List.of(sintetico));
        assertThat(resultado.semConfianca()).hasSize(1);
        assertThat(resultado.aprovadosPorOcr()).isEmpty();
        assertThat(resultado.aprovados()).contains(sintetico);
    }

    @Test
    void confiancaExatamenteIgualAoLimiarEhAprovada() {
        var resultado = OcrConfidenceGate.filtrar(List.of(new OcrConfidenceGate.ExemploComOcr("x", OcrConfidenceGate.LIMIAR_PADRAO)));
        assertThat(resultado.aprovadosPorOcr()).hasSize(1);
    }

    @Test
    void limiarCustomizadoEhRespeitado() {
        var exemplo = new OcrConfidenceGate.ExemploComOcr("x", 0.7);
        var resultadoPadrao = OcrConfidenceGate.filtrar(List.of(exemplo));
        var resultadoFrouxo = OcrConfidenceGate.filtrar(List.of(exemplo), 0.6);
        assertThat(resultadoPadrao.sinalizadosParaRevisao()).hasSize(1);
        assertThat(resultadoFrouxo.aprovadosPorOcr()).hasSize(1);
    }
}
