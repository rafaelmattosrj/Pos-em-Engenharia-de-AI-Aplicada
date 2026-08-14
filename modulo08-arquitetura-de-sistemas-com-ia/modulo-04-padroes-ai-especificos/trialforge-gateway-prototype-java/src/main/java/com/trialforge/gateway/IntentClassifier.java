package com.trialforge.gateway;

/**
 * Intent-Based Routing (Modulo 4.2): classificador de regra deterministica —
 * a ponta mais simples da escada descrita no 4.2, suficiente pra separar os
 * unicos 2 fluxos que pedem tratamento diferente aqui.
 */
public final class IntentClassifier {

    private IntentClassifier() {
    }

    public static String classificarIntencao(String pergunta) {
        String p = pergunta.toLowerCase();
        if (p.contains("csr") || p.contains("relatório final") || p.contains("síntese")
                || p.contains("evento adverso") || p.contains("desfecho")) {
            return "sintese_csr";
        }
        if (p.contains("critério") || p.contains("inclusão") || p.contains("exclusão")
                || p.contains("idade mínima") || p.contains("protocolo")) {
            return "consulta_protocolo";
        }
        return "consulta_icf";
    }
}
