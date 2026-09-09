package com.amplitudeseguros.decisionframework.config;

/**
 * Gate binário de governança de dado (LGPD): dadoSensivelLGPD exige dpaAssinado;
 * em qualquer caso, baseLegalDefinida é obrigatória. Avaliado ANTES do AHP.
 */
public record Governanca(
        boolean dadoSensivelLGPD,
        boolean baseLegalDefinida,
        String baseLegalDescricao,
        boolean dpaAssinado) {
}
