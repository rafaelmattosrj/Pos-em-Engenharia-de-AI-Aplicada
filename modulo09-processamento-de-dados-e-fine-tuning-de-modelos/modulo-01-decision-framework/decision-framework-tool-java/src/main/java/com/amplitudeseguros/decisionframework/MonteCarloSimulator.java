package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Financeiro;

import java.util.Arrays;
import java.util.function.DoubleSupplier;

/**
 * Simulação de Monte Carlo -- incerteza nos parâmetros de negócio, amostrados de
 * distribuições triangulares. Equivalente a amostrarTriangular/simularMonteCarlo
 * em decision-framework-tool.js.
 */
public final class MonteCarloSimulator {

    private MonteCarloSimulator() {
    }

    public record Resultado(double media, double p5, double p50, double p95, double probabilidadePositivo, int n) {
    }

    /** Amostra de uma distribuição triangular via inversão de CDF. */
    public static double amostrarTriangular(double min, double moda, double max, DoubleSupplier rng) {
        double u = rng.getAsDouble();
        double f = (moda - min) / (max - min);
        if (u < f) {
            return min + Math.sqrt(u * (max - min) * (moda - min));
        }
        return max - Math.sqrt((1 - u) * (max - min) * (max - moda));
    }

    public static double percentil(double[] valoresOrdenados, double p) {
        int indice = (int) Math.min(valoresOrdenados.length - 1, Math.floor(p * valoresOrdenados.length));
        return valoresOrdenados[indice];
    }

    /** Roda n simulações de NPV amostrando crescimento e custo por chamada de distribuições triangulares. */
    public static Resultado simular(Financeiro financeiro, int n, DoubleSupplier rng) {
        double[] resultados = new double[n];
        for (int i = 0; i < n; i++) {
            double crescimento = amostrarTriangular(
                    financeiro.crescimentoMensal().min(), financeiro.crescimentoMensal().moda(),
                    financeiro.crescimentoMensal().max(), rng);
            double custoStatusQuo = amostrarTriangular(
                    financeiro.custoPorChamadaStatusQuo().min(), financeiro.custoPorChamadaStatusQuo().moda(),
                    financeiro.custoPorChamadaStatusQuo().max(), rng);
            double custoFineTuned = amostrarTriangular(
                    financeiro.custoPorChamadaFineTuned().min(), financeiro.custoPorChamadaFineTuned().moda(),
                    financeiro.custoPorChamadaFineTuned().max(), rng);

            NpvCalculator.Params params = NpvCalculator.Params.semAtraso(
                    financeiro.volumeInicialMensal(), crescimento, custoStatusQuo, custoFineTuned,
                    financeiro.custoTreinamento(), financeiro.horizonteMeses(), financeiro.taxaDescontoMensal());
            resultados[i] = NpvCalculator.calcularNPV(params).npv();
        }
        Arrays.sort(resultados);

        double media = Arrays.stream(resultados).average().orElse(0);
        long positivos = Arrays.stream(resultados).filter(v -> v > 0).count();

        return new Resultado(
                Math.round(media * 100.0) / 100.0,
                Math.round(percentil(resultados, 0.05) * 100.0) / 100.0,
                Math.round(percentil(resultados, 0.50) * 100.0) / 100.0,
                Math.round(percentil(resultados, 0.95) * 100.0) / 100.0,
                Math.round((double) positivos / n * 10000.0) / 10000.0,
                n);
    }
}
