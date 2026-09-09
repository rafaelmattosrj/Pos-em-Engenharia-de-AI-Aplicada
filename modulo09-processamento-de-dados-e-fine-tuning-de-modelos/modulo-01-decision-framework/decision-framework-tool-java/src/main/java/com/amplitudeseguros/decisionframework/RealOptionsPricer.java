package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Financeiro;
import com.amplitudeseguros.decisionframework.config.OpcaoReal;

import java.util.Arrays;
import java.util.function.DoubleSupplier;

/**
 * Real Options -- precificação binomial (Cox-Ross-Rubinstein) do valor de
 * esperar, com volatilidade derivada da mesma simulação de Monte Carlo usada
 * em MonteCarloSimulator. Só se aplica quando a única reprovação é dado
 * insuficiente (falhaSoDado). Equivalente a calcularValorPresenteFluxos,
 * derivarVolatilidade e precificarOpcaoDeEsperar em decision-framework-tool.js.
 */
public final class RealOptionsPricer {

    private RealOptionsPricer() {
    }

    public record Resultado(
            int mesesParaEsperar,
            double sigma,
            double u,
            double d,
            double probabilidadeRiscoNeutra,
            double valorPresenteFluxosBrutos,
            double valorExercerAgora,
            double valorOpcaoEsperar,
            double valorDeEsperar,
            Recomendacao recomendacao,
            String reavaliacaoAgendadaEm) {
    }

    /** Valor presente BRUTO dos fluxos de economia, sem subtrair o custo de treino. */
    public static double calcularValorPresenteFluxos(NpvCalculator.Params params) {
        return NpvCalculator.calcularNPV(params).npv() + params.custoTreinamento();
    }

    /**
     * Volatilidade mensal do valor presente bruto, a partir do desvio padrão do
     * log-retorno de cada simulação em relação à mediana.
     */
    public static double derivarVolatilidade(Financeiro financeiro, int n, DoubleSupplier rng) {
        double[] valores = new double[n];
        for (int i = 0; i < n; i++) {
            double crescimento = MonteCarloSimulator.amostrarTriangular(
                    financeiro.crescimentoMensal().min(), financeiro.crescimentoMensal().moda(),
                    financeiro.crescimentoMensal().max(), rng);
            double custoStatusQuo = MonteCarloSimulator.amostrarTriangular(
                    financeiro.custoPorChamadaStatusQuo().min(), financeiro.custoPorChamadaStatusQuo().moda(),
                    financeiro.custoPorChamadaStatusQuo().max(), rng);
            double custoFineTuned = MonteCarloSimulator.amostrarTriangular(
                    financeiro.custoPorChamadaFineTuned().min(), financeiro.custoPorChamadaFineTuned().moda(),
                    financeiro.custoPorChamadaFineTuned().max(), rng);
            NpvCalculator.Params params = NpvCalculator.Params.semAtraso(
                    financeiro.volumeInicialMensal(), crescimento, custoStatusQuo, custoFineTuned,
                    financeiro.custoTreinamento(), financeiro.horizonteMeses(), financeiro.taxaDescontoMensal());
            valores[i] = calcularValorPresenteFluxos(params);
        }
        Arrays.sort(valores);
        double mediana = MonteCarloSimulator.percentil(valores, 0.5);

        double somaLog = 0;
        int qtd = 0;
        for (double v : valores) {
            if (v > 0 && mediana > 0) {
                somaLog += Math.log(v / mediana);
                qtd++;
            }
        }
        double media = somaLog / qtd;
        double variancia = 0;
        for (double v : valores) {
            if (v > 0 && mediana > 0) {
                double logRetorno = Math.log(v / mediana);
                variancia += Math.pow(logRetorno - media, 2);
            }
        }
        variancia /= (qtd - 1);
        return Math.sqrt(variancia);
    }

    /**
     * Precifica o valor de esperar como opção real (árvore binomial CRR, opção
     * americana com única janela de exercício, indução retroativa).
     */
    public static Resultado precificarOpcaoDeEsperar(
            Financeiro financeiro, double scoreAtualP3, double limiarVerde, int n, DoubleSupplier rng) {
        OpcaoReal opcaoReal = financeiro.opcaoReal();
        int mesesParaEsperar = (int) Math.ceil(
                (opcaoReal.scoreAlvo() - scoreAtualP3) / opcaoReal.taxaCrescimentoScorePorMes());
        double custoTreinamento = financeiro.custoTreinamento();

        NpvCalculator.Params paramsModa = NpvCalculator.Params.semAtraso(
                financeiro.volumeInicialMensal(), financeiro.crescimentoMensal().moda(),
                financeiro.custoPorChamadaStatusQuo().moda(), financeiro.custoPorChamadaFineTuned().moda(),
                custoTreinamento, financeiro.horizonteMeses(), financeiro.taxaDescontoMensal());
        double s0 = calcularValorPresenteFluxos(paramsModa);

        double sigma = derivarVolatilidade(financeiro, n, rng);
        double deltaT = 1;
        double r = financeiro.taxaDescontoMensal();
        double u = Math.exp(sigma * Math.sqrt(deltaT));
        double d = 1 / u;
        double p = (Math.exp(r * deltaT) - d) / (u - d);

        double[] valoresOpcao = new double[mesesParaEsperar + 1];
        for (int j = 0; j <= mesesParaEsperar; j++) {
            double sFinal = s0 * Math.pow(u, mesesParaEsperar - j) * Math.pow(d, j);
            valoresOpcao[j] = Math.max(sFinal - custoTreinamento, 0);
        }
        for (int passo = mesesParaEsperar; passo > 0; passo--) {
            double[] proximoPasso = new double[passo];
            for (int j = 0; j < passo; j++) {
                double valorEsperado = p * valoresOpcao[j] + (1 - p) * valoresOpcao[j + 1];
                proximoPasso[j] = valorEsperado / Math.exp(r * deltaT);
            }
            valoresOpcao = proximoPasso;
        }
        double valorOpcaoEsperar = Math.round(valoresOpcao[0] * 100.0) / 100.0;

        NpvCalculator.Params paramsAgora = new NpvCalculator.Params(
                financeiro.volumeInicialMensal(), financeiro.crescimentoMensal().moda(),
                financeiro.custoPorChamadaStatusQuo().moda(),
                financeiro.custoPorChamadaFineTuned().moda() + opcaoReal.custoDeErroEsperadoPorChamada(),
                custoTreinamento, financeiro.horizonteMeses(), financeiro.taxaDescontoMensal(), 0);
        double sAgora = calcularValorPresenteFluxos(paramsAgora);
        double valorExercerAgora = Math.round(Math.max(sAgora - custoTreinamento, 0) * 100.0) / 100.0;

        double valorDeEsperar = Math.round((valorOpcaoEsperar - valorExercerAgora) * 100.0) / 100.0;

        return new Resultado(
                mesesParaEsperar,
                Math.round(sigma * 10000.0) / 10000.0,
                Math.round(u * 10000.0) / 10000.0,
                Math.round(d * 10000.0) / 10000.0,
                Math.round(p * 10000.0) / 10000.0,
                Math.round(s0 * 100.0) / 100.0,
                valorExercerAgora,
                valorOpcaoEsperar,
                valorDeEsperar,
                valorDeEsperar > 0 ? Recomendacao.ESPERAR : Recomendacao.FINE_TUNING,
                "Módulo 3.2");
    }
}
