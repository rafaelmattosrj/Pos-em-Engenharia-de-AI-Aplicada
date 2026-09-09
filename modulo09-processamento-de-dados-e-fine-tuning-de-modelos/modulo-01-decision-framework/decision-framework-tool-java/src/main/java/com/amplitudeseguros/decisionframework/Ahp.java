package com.amplitudeseguros.decisionframework;

import java.util.List;

/**
 * AHP -- Analytic Hierarchy Process (Saaty 1980). Deriva pesos de uma matriz de
 * comparação pareada pelo método da média geométrica das linhas, e calcula a
 * Razão de Consistência (CR) do julgamento. Equivalente a derivarPesosAHP,
 * calcularConsistenciaAHP e agregarMatrizesComite em decision-framework-tool.js.
 */
public final class Ahp {

    /** Random Index de Saaty (n=4) -- normaliza o Índice de Consistência em Razão de Consistência. */
    public static final double RANDOM_INDEX_N4 = 0.90;

    private Ahp() {
    }

    public record Consistencia(double lambdaMax, double ci, double cr, boolean consistente) {
    }

    /** Vetor de prioridades (pesos), soma 1.0, pela média geométrica das linhas (row geometric mean method). */
    public static double[] derivarPesos(double[][] matriz) {
        int n = matriz.length;
        double[] mediasGeometricas = new double[n];
        for (int i = 0; i < n; i++) {
            double produto = 1;
            for (double v : matriz[i]) {
                produto *= v;
            }
            mediasGeometricas[i] = Math.pow(produto, 1.0 / n);
        }
        double soma = 0;
        for (double v : mediasGeometricas) {
            soma += v;
        }
        double[] pesos = new double[n];
        for (int i = 0; i < n; i++) {
            pesos[i] = mediasGeometricas[i] / soma;
        }
        return pesos;
    }

    /** CR < 0.10 (limiar padrão de Saaty) considera o julgamento consistente o bastante para confiar nos pesos. */
    public static Consistencia calcularConsistencia(double[][] matriz, double[] pesos) {
        int n = matriz.length;
        double[] aw = new double[n];
        for (int i = 0; i < n; i++) {
            double soma = 0;
            for (int j = 0; j < n; j++) {
                soma += matriz[i][j] * pesos[j];
            }
            aw[i] = soma;
        }
        double somaRazoes = 0;
        for (int i = 0; i < n; i++) {
            somaRazoes += aw[i] / pesos[i];
        }
        double lambdaMax = somaRazoes / n;
        double ci = (lambdaMax - n) / (n - 1);
        double cr = ci / RANDOM_INDEX_N4;
        return new Consistencia(lambdaMax, ci, cr, cr < 0.10);
    }

    /**
     * Agrega várias matrizes de comparação pareada (uma por avaliador) pela média
     * geométrica célula a célula (método AIJ) -- preserva a propriedade recíproca
     * (a[j][i] = 1/a[i][j]), então o resultado ainda é uma matriz de comparação
     * pareada válida. Um comitê de 1 avaliador reduz ao caso original.
     */
    public static double[][] agregarMatrizesComite(List<double[][]> matrizes) {
        int nAvaliadores = matrizes.size();
        int n = matrizes.get(0).length;
        double[][] agregada = new double[n][n];
        for (int i = 0; i < n; i++) {
            for (int j = 0; j < n; j++) {
                double produto = 1;
                for (double[][] matriz : matrizes) {
                    produto *= matriz[i][j];
                }
                agregada[i][j] = Math.pow(produto, 1.0 / nAvaliadores);
            }
        }
        return agregada;
    }
}
