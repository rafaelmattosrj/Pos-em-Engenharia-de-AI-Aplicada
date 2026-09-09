package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Financeiro;

import java.util.ArrayList;
import java.util.List;

/**
 * NPV / DCF -- fluxo de caixa descontado, fine-tuning vs. status quo.
 * Equivalente a paramsDeterministicos/calcularNPV em decision-framework-tool.js.
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

        public Params {
        }

        public static Params semAtraso(
                double volumeInicialMensal, double crescimentoMensal, double custoPorChamadaStatusQuo,
                double custoPorChamadaFineTuned, double custoTreinamento, int horizonteMeses, double taxaDescontoMensal) {
            return new Params(volumeInicialMensal, crescimentoMensal, custoPorChamadaStatusQuo,
                    custoPorChamadaFineTuned, custoTreinamento, horizonteMeses, taxaDescontoMensal, 0);
        }
    }

    public record FluxoMensal(int mes, double volume, double economiaDescontada, double npvAcumulado) {
    }

    public record Resultado(double npv, Integer mesBreakeven, List<FluxoMensal> fluxos) {
    }

    /** Converte o bloco `financeiro` num cenário determinístico usando o valor "moda" de cada distribuição. */
    public static Params paramsDeterministicos(Financeiro financeiro) {
        return Params.semAtraso(
                financeiro.volumeInicialMensal(),
                financeiro.crescimentoMensal().moda(),
                financeiro.custoPorChamadaStatusQuo().moda(),
                financeiro.custoPorChamadaFineTuned().moda(),
                financeiro.custoTreinamento(),
                financeiro.horizonteMeses(),
                financeiro.taxaDescontoMensal());
    }

    /**
     * NPV de migrar pra um modelo fine-tunado: economia mensal (volume x diferença
     * de custo por chamada), descontada mês a mês, menos o custo de treinar.
     * Volume cresce geometricamente pela taxa informada.
     */
    public static Resultado calcularNPV(Params p) {
        double npv = -p.custoTreinamento();
        double volume = p.volumeInicialMensal();
        List<FluxoMensal> fluxos = new ArrayList<>();
        Integer mesBreakeven = null;

        for (int mes = 1; mes <= p.horizonteMeses(); mes++) {
            volume *= 1 + p.crescimentoMensal();
            double economiaDescontada = 0;
            // durante o atraso (esperando dado), não há economia: ainda se paga o status quo
            if (mes > p.atrasoMeses()) {
                double economiaMes = volume * (p.custoPorChamadaStatusQuo() - p.custoPorChamadaFineTuned());
                double fatorDesconto = Math.pow(1 + p.taxaDescontoMensal(), mes);
                economiaDescontada = economiaMes / fatorDesconto;
                npv += economiaDescontada;
            }
            if (mesBreakeven == null && npv > 0) {
                mesBreakeven = mes;
            }
            fluxos.add(new FluxoMensal(mes, volume, economiaDescontada, npv));
        }

        return new Resultado(Math.round(npv * 100.0) / 100.0, mesBreakeven, fluxos);
    }
}
