package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Governanca;

import java.util.ArrayList;
import java.util.List;

/**
 * Gate de governança/compliance (LGPD) — roda ANTES de tudo. Diferente das 4
 * perguntas ponderadas por AHP, governança é binária: reprova o caso ali, sem
 * gastar o resto do pipeline (AHP, NPV, Monte Carlo, Real Options).
 * Equivalente a validarGovernancaDado() em decision-framework-tool.js.
 */
public final class GovernanceGate {

    private GovernanceGate() {
    }

    public record Resultado(boolean aprovado, List<String> motivos) {
    }

    public static Resultado validar(Governanca g) {
        List<String> motivos = new ArrayList<>();

        if (!g.baseLegalDefinida()) {
            motivos.add("sem base legal definida pro tratamento do dado (LGPD Art. 7º/11)");
        }
        if (g.dadoSensivelLGPD() && !g.dpaAssinado()) {
            motivos.add("dado de categoria sensível (LGPD Art. 5º, II) sem DPA assinado com o provedor de fine-tuning");
        }

        return new Resultado(motivos.isEmpty(), motivos);
    }
}
