package com.trialforge.tiering;

import java.util.Locale;

/** Classificação de intenção — lógica pura, sem rede. */
public final class IntentClassifier {

    public static final String SINTESE_CSR = "sintese_csr";
    public static final String CONSULTA_CLAUSULA = "consulta_clausula";

    private IntentClassifier() {
    }

    public static String classificarIntencao(String pergunta) {
        String p = pergunta.toLowerCase(Locale.ROOT);
        if (p.contains("csr") || p.contains("relatório final") || p.contains("síntese")) {
            return SINTESE_CSR;
        }
        return CONSULTA_CLAUSULA;
    }
}
