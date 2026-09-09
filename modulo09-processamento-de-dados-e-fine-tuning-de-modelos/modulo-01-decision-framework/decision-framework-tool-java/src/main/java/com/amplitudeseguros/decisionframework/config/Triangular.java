package com.amplitudeseguros.decisionframework.config;

/** Distribuição triangular {min, moda, max} usada para os parâmetros de negócio incertos. */
public record Triangular(double min, double moda, double max) {
}
