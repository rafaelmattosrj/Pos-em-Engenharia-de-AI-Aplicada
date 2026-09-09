package com.finetuningapi;

import java.util.ArrayList;
import java.util.List;

/**
 * Gate de confianca de OCR (Modulo 3.2 / dataset-upload-and-tracking-tool.js):
 * exemplo cujo metadata.confiancaOcr esta abaixo do limiar e sinalizado pra
 * revisao humana; exemplo sem confiancaOcr (texto sintetico, nunca escaneado)
 * segue aprovado sem passar pelo gate.
 */
public final class OcrConfidenceGate {

    public static final double LIMIAR_PADRAO = 0.85;

    private OcrConfidenceGate() {
    }

    public record ExemploComOcr(String id, Double confiancaOcr) {
    }

    public record ResultadoFiltro(
            List<ExemploComOcr> aprovados,
            List<ExemploComOcr> aprovadosPorOcr,
            List<ExemploComOcr> semConfianca,
            List<ExemploComOcr> sinalizadosParaRevisao,
            double limiar) {
    }

    public static ResultadoFiltro filtrar(List<ExemploComOcr> exemplos) {
        return filtrar(exemplos, LIMIAR_PADRAO);
    }

    public static ResultadoFiltro filtrar(List<ExemploComOcr> exemplos, double limiar) {
        List<ExemploComOcr> semConfianca = new ArrayList<>();
        List<ExemploComOcr> aprovadosPorOcr = new ArrayList<>();
        List<ExemploComOcr> sinalizados = new ArrayList<>();

        for (ExemploComOcr exemplo : exemplos) {
            Double confianca = exemplo.confiancaOcr();
            if (confianca == null) {
                semConfianca.add(exemplo);
            } else if (confianca >= limiar) {
                aprovadosPorOcr.add(exemplo);
            } else {
                sinalizados.add(exemplo);
            }
        }

        List<ExemploComOcr> aprovados = new ArrayList<>(aprovadosPorOcr);
        aprovados.addAll(semConfianca);
        return new ResultadoFiltro(aprovados, aprovadosPorOcr, semConfianca, sinalizados, limiar);
    }
}
