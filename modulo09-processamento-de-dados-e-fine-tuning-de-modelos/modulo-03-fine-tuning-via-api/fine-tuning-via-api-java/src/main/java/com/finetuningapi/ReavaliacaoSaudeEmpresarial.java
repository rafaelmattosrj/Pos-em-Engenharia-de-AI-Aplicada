package com.finetuningapi;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Callback real pros Modulos 1.2/1.3 (Modulo 3.2): reabre o mesmo gate de
 * decisao (DecisionFrameworkCore) com o caso Saude Empresarial atualizado
 * pros 9 meses que se passaram, usando a mesma taxa de crescimento de score
 * ja projetada no caso original -- nao inventa numero novo.
 */
public final class ReavaliacaoSaudeEmpresarial {

    public static final int MESES_DECORRIDOS = 9;

    private ReavaliacaoSaudeEmpresarial() {
    }

    public static DecisionFrameworkCore.Caso construirCasoNoveMesesDepois(DecisionFrameworkCore.Caso casoOriginal) {
        DecisionFrameworkCore.OpcaoReal opcaoReal = casoOriginal.financeiro().opcaoReal();
        double p3Atualizado = Math.round((casoOriginal.scores().get("p3")
                + opcaoReal.taxaCrescimentoScorePorMes() * MESES_DECORRIDOS) * 100) / 100.0;

        Map<String, Double> scoresAtualizados = new LinkedHashMap<>(casoOriginal.scores());
        scoresAtualizados.put("p3", p3Atualizado);

        double fatorCrescimentoVolume = Math.pow(1 + casoOriginal.financeiro().crescimentoMensalModa(), MESES_DECORRIDOS);
        int volumeAtualizado = (int) Math.round(casoOriginal.financeiro().volumeInicialMensal() * fatorCrescimentoVolume);

        DecisionFrameworkCore.Financeiro financeiroAtualizado = new DecisionFrameworkCore.Financeiro(
                volumeAtualizado, casoOriginal.financeiro().crescimentoMensalModa(), opcaoReal);

        return new DecisionFrameworkCore.Caso(casoOriginal.id(), scoresAtualizados, financeiroAtualizado);
    }
}
