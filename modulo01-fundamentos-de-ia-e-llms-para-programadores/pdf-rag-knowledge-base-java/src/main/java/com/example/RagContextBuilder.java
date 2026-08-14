package com.example;

import java.util.List;
import java.util.stream.Collectors;

/**
 * Filtra os trechos recuperados do vector store por score minimo de
 * relevancia e monta o contexto textual usado no prompt do LLM.
 *
 * Extraido de Main.java para tornar a regra de negocio (filtro + junção)
 * testavel isoladamente, sem depender de Neo4j ou de um modelo de
 * embeddings real.
 */
public final class RagContextBuilder {

    /** Score minimo (exclusivo) para um trecho ser considerado relevante. */
    public static final double MIN_SCORE = 0.5;

    /** Separador usado para juntar múltiplos trechos relevantes. */
    public static final String SEPARATOR = "\n\n---\n\n";

    private RagContextBuilder() {
    }

    /** Um trecho recuperado do vector store, com seu texto e score de similaridade. */
    public record Trecho(String texto, double score) {
    }

    /**
     * Filtra os trechos com {@code score > MIN_SCORE} e junta seus textos
     * com {@link #SEPARATOR}. Retorna string vazia (nunca {@code null})
     * quando nenhum trecho atinge o score minimo.
     */
    public static String build(List<Trecho> matches) {
        return matches.stream()
                .filter(m -> m.score() > MIN_SCORE)
                .map(Trecho::texto)
                .collect(Collectors.joining(SEPARATOR));
    }
}
