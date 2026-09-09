package com.finetuningapi;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Conversao do schema canonico (instrucao/entrada/saida/metadata) pro formato
 * de fine-tuning supervisionado da Vertex AI (contents/role/parts).
 * Porte de dataset-upload-and-tracking-tool.js (Modulo 3.2) e
 * finetuning-automation-tool.js (Modulo 3.4): a saida e um objeto estruturado,
 * serializado como JSON no turno do modelo.
 */
public final class GeminiConverter {

    private GeminiConverter() {
    }

    public record Exemplo(String instrucao, String entrada, Map<String, Object> saida) {
    }

    public record Turno(String role, String text) {
    }

    public record ConversaoGemini(List<Turno> contents) {
    }

    public static ConversaoGemini converter(Exemplo exemplo) {
        if (exemplo.instrucao() == null || exemplo.instrucao().isBlank()) {
            throw new IllegalArgumentException("exemplo sem instrucao valida");
        }
        if (exemplo.entrada() == null || exemplo.entrada().isBlank()) {
            throw new IllegalArgumentException("exemplo sem entrada valida");
        }
        if (exemplo.saida() == null) {
            throw new IllegalArgumentException("exemplo sem saida valida");
        }
        String textoUsuario = exemplo.instrucao() + "\n\n" + exemplo.entrada();
        String textoModelo = JsonUtil.toJson(exemplo.saida());
        return new ConversaoGemini(List.of(new Turno("user", textoUsuario), new Turno("model", textoModelo)));
    }

    /**
     * Variante do extra Dolly-15k (dolly-vertex-pipeline.js): a saida do Dolly
     * e texto solto, nao objeto estruturado -- por isso nao passa por
     * JsonUtil.toJson, senao a resposta esperada sairia com aspas extras.
     */
    public static ConversaoGemini converterTexto(String instrucao, String entrada, String saidaTexto) {
        if (instrucao == null || instrucao.isBlank()) throw new IllegalArgumentException("instrucao obrigatoria");
        if (entrada == null || entrada.isBlank()) throw new IllegalArgumentException("entrada obrigatoria");
        if (saidaTexto == null || saidaTexto.isBlank()) throw new IllegalArgumentException("saida (texto) obrigatoria");
        String textoUsuario = instrucao + "\n\n" + entrada;
        return new ConversaoGemini(List.of(new Turno("user", textoUsuario), new Turno("model", saidaTexto)));
    }

    public static Map<String, Object> exemploMap(String segurado, String placa, double valor) {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("segurado", segurado);
        m.put("placa", placa);
        m.put("valor", valor);
        return m;
    }
}
