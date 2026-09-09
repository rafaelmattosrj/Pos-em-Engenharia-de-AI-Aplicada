package com.finetuningapi;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class ReavaliacaoSaudeEmpresarialTest {

    @Test
    void reavaliacaoNoveMesesDepoisAprovaOCasoQueAntesReprovavaSoPorDado() throws Exception {
        var config = DecisionFrameworkCore.carregarConfiguracao();
        double[] pesosAHP = DecisionFrameworkCore.derivarPesosAHP(config.matrizAhp());
        var casoOriginal = config.caso("amplitude-saude-empresarial");
        var casoAtualizado = ReavaliacaoSaudeEmpresarial.construirCasoNoveMesesDepois(casoOriginal);

        var resultadoOriginal = DecisionFrameworkCore.avaliarFramework(casoOriginal.scores(), pesosAHP, config.limiarVerde());
        var resultadoAtualizado = DecisionFrameworkCore.avaliarFramework(casoAtualizado.scores(), pesosAHP, config.limiarVerde());

        assertThat(resultadoOriginal.aprovado()).isFalse();
        assertThat(resultadoOriginal.falhaSoDado()).isTrue();

        assertThat(casoAtualizado.scores().get("p3")).isGreaterThanOrEqualTo(config.limiarVerde());
        assertThat(resultadoAtualizado.aprovado()).isTrue();
        assertThat(resultadoAtualizado.perguntasFalhas()).isEmpty();

        assertThat(casoAtualizado.scores().get("p1")).isEqualTo(casoOriginal.scores().get("p1"));
        assertThat(casoAtualizado.scores().get("p2")).isEqualTo(casoOriginal.scores().get("p2"));
        assertThat(casoAtualizado.scores().get("p4")).isEqualTo(casoOriginal.scores().get("p4"));

        assertThat(casoAtualizado.financeiro().volumeInicialMensal())
                .isGreaterThan(casoOriginal.financeiro().volumeInicialMensal());
    }

    @Test
    void ahpDerivaPesosQueSomamUmEDaMaiorPesoAP3() throws Exception {
        var config = DecisionFrameworkCore.carregarConfiguracao();
        double[] pesos = DecisionFrameworkCore.derivarPesosAHP(config.matrizAhp());
        double soma = 0;
        for (double p : pesos) soma += p;
        assertThat(soma).isCloseTo(1.0, org.assertj.core.data.Offset.offset(1e-9));
        int indiceP3 = DecisionFrameworkCore.CHAVES_PERGUNTAS.indexOf("p3");
        double maiorPeso = java.util.Arrays.stream(pesos).max().orElseThrow();
        assertThat(pesos[indiceP3]).isEqualTo(maiorPeso);
    }
}
