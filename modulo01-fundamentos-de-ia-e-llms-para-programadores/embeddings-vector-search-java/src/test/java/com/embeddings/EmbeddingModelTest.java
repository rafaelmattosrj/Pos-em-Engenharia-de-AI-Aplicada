package com.embeddings;

import dev.langchain4j.data.embedding.Embedding;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.embedding.onnx.allminilml6v2.AllMiniLmL6V2EmbeddingModel;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre o mesmo cenario de "sucesso" testado no cliente de embeddings do
 * porte Go (embeddings/ollama_test.go via httptest.Server): gerar um vetor
 * de embedding com a dimensao esperada (384) para um texto.
 *
 * Aqui nao ha chamada HTTP para mockar — a versao Java roda o modelo
 * all-MiniLM-L6-v2 in-process via ONNX Runtime (sem rede, sem servidor
 * externo), entao os cenarios de erro de servidor/timeout do cliente HTTP Go
 * nao tem equivalente nesta arquitetura.
 */
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class EmbeddingModelTest {

    private static final int EXPECTED_DIMENSION = 384;

    private EmbeddingModel model;

    @BeforeAll
    void loadModel() {
        model = new AllMiniLmL6V2EmbeddingModel();
    }

    @Test
    void modelReportsExpectedDimension() {
        assertThat(model.dimension()).isEqualTo(EXPECTED_DIMENSION);
    }

    @Test
    void embedReturnsVectorWithExpectedDimension() {
        Embedding embedding = model.embed("O que sao tensores?").content();

        assertThat(embedding.vector()).hasSize(EXPECTED_DIMENSION);
    }

    @Test
    void embedIsDeterministicForTheSameText() {
        String text = "tensores em JavaScript";

        Embedding first = model.embed(text).content();
        Embedding second = model.embed(text).content();

        assertThat(first.vector()).isEqualTo(second.vector());
    }

    @Test
    void embedProducesDifferentVectorsForDifferentTexts() {
        Embedding a = model.embed("tensores e redes neurais").content();
        Embedding b = model.embed("receita de bolo de chocolate").content();

        assertThat(a.vector()).isNotEqualTo(b.vector());
    }

    @Test
    void semanticallyCloserTextsHaveHigherCosineSimilarity() {
        Embedding tensorQuestion = model.embed("O que sao tensores?").content();
        Embedding tensorRelated = model.embed("Tensores sao estruturas de dados para redes neurais").content();
        Embedding unrelated = model.embed("Receita de bolo de chocolate").content();

        double similarityToRelated = cosineSimilarity(tensorQuestion.vector(), tensorRelated.vector());
        double similarityToUnrelated = cosineSimilarity(tensorQuestion.vector(), unrelated.vector());

        assertThat(similarityToRelated).isGreaterThan(similarityToUnrelated);
    }

    private double cosineSimilarity(float[] a, float[] b) {
        double dot = 0, normA = 0, normB = 0;
        for (int i = 0; i < a.length; i++) {
            dot += a[i] * b[i];
            normA += a[i] * a[i];
            normB += b[i] * b[i];
        }
        return dot / (Math.sqrt(normA) * Math.sqrt(normB));
    }
}
