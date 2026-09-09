package com.amplitudeseguros.decisionframework.config;

import java.util.Map;

/** Um caso de negócio da Amplitude Seguros: scores das 4 perguntas, governança e (opcional) financeiro. */
public record Caso(
        String id,
        String nome,
        String tarefa,
        Map<String, Double> scores,
        Governanca governanca,
        Financeiro financeiro) {
}
