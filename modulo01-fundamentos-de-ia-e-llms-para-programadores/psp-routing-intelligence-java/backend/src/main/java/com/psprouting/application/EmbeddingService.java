package com.psprouting.application;

import dev.langchain4j.data.embedding.Embedding;
import dev.langchain4j.model.embedding.EmbeddingModel;
import org.springframework.stereotype.Service;

/**
 * Passo (1) do pipeline: gera o vetor de 384 dimensoes de uma transacao a
 * partir da sua representacao textual, usando o modelo local
 * all-MiniLM-L6-v2 (ONNX, sem chamada de API) — mesmo modelo usado em
 * {@code embeddings-vector-search-java} e {@code pdf-rag-knowledge-base-java}.
 */
@Service
public class EmbeddingService {

    private final EmbeddingModel embeddingModel;

    public EmbeddingService(EmbeddingModel embeddingModel) {
        this.embeddingModel = embeddingModel;
    }

    public Embedding embed(String text) {
        return embeddingModel.embed(text).content();
    }
}
