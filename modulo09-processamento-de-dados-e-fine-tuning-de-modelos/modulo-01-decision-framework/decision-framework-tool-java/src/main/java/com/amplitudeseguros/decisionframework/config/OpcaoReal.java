package com.amplitudeseguros.decisionframework.config;

/**
 * Parâmetros de Real Options, só presentes quando o caso é elegível (reprovação
 * exclusivamente por dado insuficiente, pergunta p3).
 */
public record OpcaoReal(double custoDeErroEsperadoPorChamada, double taxaCrescimentoScorePorMes, double scoreAlvo) {
}
