package com.lorapeft;

import java.util.ArrayList;
import java.util.List;

/**
 * Reimplementacao de {@code calcularNPV} (NPV/DCF) de
 * modulo-01-decision-framework/decision-framework-tool.js. Nao ha um pacote
 * Java compartilhado entre os submodulos deste repositorio (cada modulo
 * porta para uma pasta -java/-go independente, mesmo padrao usado em todo o
 * repo), entao esta classe reimplementa a mesma formula aqui, em vez de criar
 * uma dependencia Maven entre projetos de modulos diferentes. Qualquer
 * mudanca na formula original deve ser replicada aqui manualmente.
 */
public final class NpvCalculator {

    private NpvCalculator() {
    }

    public record Params(
            double volumeInicialMensal,
            double crescimentoMensal,
            double custoPorChamadaStatusQuo,
            double custoPorChamadaFineTuned,
            double custoTreinamento,
            int horizonteMeses,
            double taxaDescontoMensal,
            int atrasoMeses) {

        public static Params of(
                double volumeInicialMensal,
                double crescimentoMensal,
                double custoPorChamadaStatusQuo,
                double custoPorChamadaFineTuned,
                double custoTreinamento,
                int horizonteMeses,
                double taxaDescontoMensal) {
            return new Params(volumeInicialMensal, crescimentoMensal, custoPorChamadaStatusQuo,
                    custoPorChamadaFineTuned, custoTreinamento, horizonteMeses, taxaDescontoMensal, 0);
        }
    }

    public record Fluxo(int mes, double volume, double economiaDescontada, double npvAcumulado) {
    }

    public record Resultado(double npv, Integer mesBreakeven, List<Fluxo> fluxos) {
    }

    public static Resultado calcularNPV(Params params) {
        double npv = -params.custoTreinamento();
        double volume = params.volumeInicialMensal();
        List<Fluxo> fluxos = new ArrayList<>();
        Integer mesBreakeven = null;

        for (int mes = 1; mes <= params.horizonteMeses(); mes++) {
            volume *= 1 + params.crescimentoMensal();
            double economiaDescontada = 0;
            if (mes > params.atrasoMeses()) {
                double economiaMes = volume * (params.custoPorChamadaStatusQuo() - params.custoPorChamadaFineTuned());
                double fatorDesconto = Math.pow(1 + params.taxaDescontoMensal(), mes);
                economiaDescontada = economiaMes / fatorDesconto;
                npv += economiaDescontada;
            }
            if (mesBreakeven == null && npv > 0) {
                mesBreakeven = mes;
            }
            fluxos.add(new Fluxo(mes, volume, economiaDescontada, npv));
        }

        return new Resultado(round2(npv), mesBreakeven, fluxos);
    }

    private static double round2(double v) {
        return Math.round(v * 100.0) / 100.0;
    }
}
