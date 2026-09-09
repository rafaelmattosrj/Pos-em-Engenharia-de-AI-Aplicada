package com.amplitudeseguros.decisionframework.config;

import java.util.List;

/** Matriz de comparação pareada (escala de Saaty) e a ordem das perguntas que ela representa. */
public record AhpConfig(List<String> ordemPerguntas, double[][] matriz) {
}
