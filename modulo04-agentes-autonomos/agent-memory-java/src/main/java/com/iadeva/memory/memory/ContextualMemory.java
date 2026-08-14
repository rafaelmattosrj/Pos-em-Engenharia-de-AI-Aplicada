package com.iadeva.memory.memory;

// Memória contextual — equivalente ao contextual_memory.py — busca semântica via embeddings

import org.springframework.ai.embedding.EmbeddingModel;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Memória contextual: armazena conteúdo com embeddings e permite busca semântica por similaridade.
 * Utiliza cosseno como métrica de distância — threshold configurável via app.memory.similarity-threshold.
 */
@Component
public class ContextualMemory {

    private final EmbeddingModel embeddingModel;
    private final double similarityThreshold;

    // Armazena entradas em memória (sem persistência em disco — reconstruído a cada inicialização)
    private final List<EmbeddingEntry> entries = new ArrayList<>();

    public ContextualMemory(
            EmbeddingModel embeddingModel,
            @Value("${app.memory.similarity-threshold:0.7}") double similarityThreshold) {
        this.embeddingModel = embeddingModel;
        this.similarityThreshold = similarityThreshold;
    }

    /**
     * Armazena conteúdo gerando seu embedding automaticamente.
     *
     * @param content  texto a ser indexado
     * @param metadata metadados associados (fonte, tipo, etc.)
     */
    public void store(String content, Map<String, Object> metadata) {
        float[] embedding = embeddingModel.embed(content);
        entries.add(new EmbeddingEntry(content, embedding, metadata));
    }

    /**
     * Busca os topK conteúdos mais semanticamente próximos da query.
     * Filtra resultados abaixo do threshold de similaridade configurado.
     *
     * @param query texto de busca
     * @param topK  número máximo de resultados
     * @return lista de entradas ordenadas por similaridade decrescente
     */
    public List<EmbeddingEntry> search(String query, int topK) {
        float[] queryEmbedding = embeddingModel.embed(query);

        return entries.stream()
                .map(entry -> new ScoredEntry(entry, cosineSimilarity(queryEmbedding, entry.embedding())))
                .filter(scored -> scored.score() >= similarityThreshold)
                .sorted((a, b) -> Double.compare(b.score(), a.score()))
                .limit(topK)
                .map(ScoredEntry::entry)
                .collect(Collectors.toList());
    }

    /**
     * Retorna o threshold de similaridade configurado — útil para testes e diagnóstico.
     */
    public double getSimilarityThreshold() {
        return similarityThreshold;
    }

    // Calcula similaridade de cosseno entre dois vetores
    private double cosineSimilarity(float[] a, float[] b) {
        double dotProduct = 0.0;
        double normA = 0.0;
        double normB = 0.0;

        int len = Math.min(a.length, b.length);
        for (int i = 0; i < len; i++) {
            dotProduct += a[i] * b[i];
            normA += a[i] * a[i];
            normB += b[i] * b[i];
        }

        double denominator = Math.sqrt(normA) * Math.sqrt(normB);
        return denominator == 0.0 ? 0.0 : dotProduct / denominator;
    }

    /**
     * Entrada da memória contextual: conteúdo + embedding + metadados.
     */
    public record EmbeddingEntry(
            String content,
            float[] embedding,
            Map<String, Object> metadata
    ) {}

    // Auxiliar interno para ordenação por score
    private record ScoredEntry(EmbeddingEntry entry, double score) {}
}
