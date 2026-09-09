package com.amplitudeseguros.decisionframework.config;

/** Bloco financeiro de um caso: parâmetros de NPV/DCF, Monte Carlo e (opcional) Real Options. */
public record Financeiro(
        double volumeInicialMensal,
        Triangular crescimentoMensal,
        Triangular custoPorChamadaStatusQuo,
        Triangular custoPorChamadaFineTuned,
        double custoTreinamento,
        int horizonteMeses,
        double taxaDescontoMensal,
        OpcaoReal opcaoReal) {
}
