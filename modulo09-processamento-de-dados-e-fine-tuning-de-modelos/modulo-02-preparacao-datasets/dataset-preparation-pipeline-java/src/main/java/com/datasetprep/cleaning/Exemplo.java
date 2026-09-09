package com.datasetprep.cleaning;

import java.util.Map;

/** Um exemplo de treino simulado: instrucao + entrada + saida estruturada + metadata. */
public record Exemplo(String instrucao, String entrada, Map<String, Object> saida, Metadata metadata) {

    public record Metadata(String caso, String fonte, String id) {
    }

    public Exemplo comMetadata(Metadata novaMetadata) {
        return new Exemplo(instrucao, entrada, saida, novaMetadata);
    }
}
