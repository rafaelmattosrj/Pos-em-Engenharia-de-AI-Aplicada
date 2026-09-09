package com.amplitudeseguros.decisionframework.grpo;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Recompensa mecanicamente checável: fração de campos exatamente corretos --
 * o "grader" que o GRPO e o RFT usam, sem modelo de recompensa aprendido.
 * Equivalente a recompensaVerificavel()/GABARITO em grpo-verifiable-reward-demo.js.
 */
public final class VerifiableReward {

    public static final Map<String, Object> GABARITO = criarGabarito();

    private VerifiableReward() {
    }

    private static Map<String, Object> criarGabarito() {
        Map<String, Object> gabarito = new LinkedHashMap<>();
        gabarito.put("claimant_name", "Marcos Vinícius Almeida Teixeira");
        gabarito.put("claim_type", "Colisao veicular");
        gabarito.put("incident_date", "12/07/2026");
        gabarito.put("estimated_amount_brl", 8450.0);
        gabarito.put("status", "Em analise pericial");
        gabarito.put("prioridade", "alta");
        return gabarito;
    }

    public static double recompensaVerificavel(Map<String, Object> candidato) {
        if (candidato == null) {
            return 0.0;
        }
        int acertos = 0;
        for (Map.Entry<String, Object> entry : GABARITO.entrySet()) {
            Object obtido = candidato.get(entry.getKey());
            Object esperado = entry.getValue();
            if (esperado instanceof Number esperadoNum) {
                Double numObtido = paraDouble(obtido);
                if (numObtido != null && Math.abs(numObtido - esperadoNum.doubleValue()) < 0.01) {
                    acertos++;
                }
            } else if (String.valueOf(obtido == null ? "" : obtido).trim().equalsIgnoreCase(String.valueOf(esperado).trim())) {
                acertos++;
            }
        }
        return (double) acertos / GABARITO.size();
    }

    private static Double paraDouble(Object obtido) {
        if (obtido instanceof Number n) {
            return n.doubleValue();
        }
        if (obtido instanceof String s) {
            try {
                return Double.parseDouble(s);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }
}
