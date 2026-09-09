package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Financeiro;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

/**
 * Análise de sensibilidade -- varia cada parâmetro financeiro em +/- percentual
 * (padrão 20%) e mede o quanto o NPV oscila, ranqueado do maior impacto pro
 * menor. Equivalente a analisarSensibilidade em decision-framework-tool.js.
 */
public final class SensitivityAnalyzer {

    private SensitivityAnalyzer() {
    }

    public record Resultado(String parametro, double npvBaixo, double npvAlto, double amplitude) {
    }

    public static List<Resultado> analisar(Financeiro financeiro, double percentualVariacao) {
        NpvCalculator.Params base = NpvCalculator.paramsDeterministicos(financeiro);

        List<Resultado> resultado = new ArrayList<>();
        resultado.add(variar(base, "crescimentoMensal", base.crescimentoMensal(), percentualVariacao));
        resultado.add(variar(base, "custoPorChamadaStatusQuo", base.custoPorChamadaStatusQuo(), percentualVariacao));
        resultado.add(variar(base, "custoPorChamadaFineTuned", base.custoPorChamadaFineTuned(), percentualVariacao));
        resultado.add(variar(base, "custoTreinamento", base.custoTreinamento(), percentualVariacao));

        resultado.sort(Comparator.comparingDouble(Resultado::amplitude).reversed());
        return resultado;
    }

    private static Resultado variar(NpvCalculator.Params base, String parametro, double valorBase, double percentual) {
        double baixoValor = valorBase * (1 - percentual);
        double altoValor = valorBase * (1 + percentual);

        double npvBaixo = NpvCalculator.calcularNPV(comValor(base, parametro, baixoValor)).npv();
        double npvAlto = NpvCalculator.calcularNPV(comValor(base, parametro, altoValor)).npv();

        return new Resultado(parametro, npvBaixo, npvAlto, Math.round(Math.abs(npvAlto - npvBaixo) * 100.0) / 100.0);
    }

    private static NpvCalculator.Params comValor(NpvCalculator.Params base, String parametro, double valor) {
        return switch (parametro) {
            case "crescimentoMensal" -> new NpvCalculator.Params(
                    base.volumeInicialMensal(), valor, base.custoPorChamadaStatusQuo(),
                    base.custoPorChamadaFineTuned(), base.custoTreinamento(), base.horizonteMeses(),
                    base.taxaDescontoMensal(), base.atrasoMeses());
            case "custoPorChamadaStatusQuo" -> new NpvCalculator.Params(
                    base.volumeInicialMensal(), base.crescimentoMensal(), valor,
                    base.custoPorChamadaFineTuned(), base.custoTreinamento(), base.horizonteMeses(),
                    base.taxaDescontoMensal(), base.atrasoMeses());
            case "custoPorChamadaFineTuned" -> new NpvCalculator.Params(
                    base.volumeInicialMensal(), base.crescimentoMensal(), base.custoPorChamadaStatusQuo(),
                    valor, base.custoTreinamento(), base.horizonteMeses(),
                    base.taxaDescontoMensal(), base.atrasoMeses());
            case "custoTreinamento" -> new NpvCalculator.Params(
                    base.volumeInicialMensal(), base.crescimentoMensal(), base.custoPorChamadaStatusQuo(),
                    base.custoPorChamadaFineTuned(), valor, base.horizonteMeses(),
                    base.taxaDescontoMensal(), base.atrasoMeses());
            default -> throw new IllegalArgumentException("parametro desconhecido: " + parametro);
        };
    }
}
