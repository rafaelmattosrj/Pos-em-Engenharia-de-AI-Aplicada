package com.embeddings;

import dev.langchain4j.data.document.Document;
import dev.langchain4j.data.document.DocumentSplitter;
import dev.langchain4j.data.document.splitter.DocumentSplitters;
import dev.langchain4j.data.segment.TextSegment;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre o mesmo comportamento de chunking testado no porte Go
 * (splitter/splitter_test.go): texto curto vira um unico chunk, texto longo
 * respeita o chunkSize configurado, e chunks consecutivos tem overlap.
 *
 * Aqui usamos diretamente o {@link DocumentSplitters#recursive} do
 * LangChain4j com os mesmos parametros do Main (1000/200), ja que — ao
 * contrario do porte Go, que reimplementa o splitter na mao — a versao Java
 * delega essa logica para a biblioteca.
 */
class DocumentSplitterTest {

    private static final int CHUNK_SIZE = 1000;
    private static final int OVERLAP = 200;

    @Test
    void shortTextProducesSingleChunk() {
        DocumentSplitter splitter = DocumentSplitters.recursive(CHUNK_SIZE, OVERLAP);
        Document document = Document.from("hello world");

        List<TextSegment> segments = splitter.split(document);

        assertThat(segments).hasSize(1);
        assertThat(segments.get(0).text()).isEqualTo("hello world");
    }

    @Test
    void longTextProducesMultipleChunksRespectingChunkSize() {
        DocumentSplitter splitter = DocumentSplitters.recursive(CHUNK_SIZE, OVERLAP);
        String text = "palavra ".repeat(500); // 4000 chars
        Document document = Document.from(text);

        List<TextSegment> segments = splitter.split(document);

        assertThat(segments.size()).isGreaterThanOrEqualTo(2);
        assertThat(segments).allSatisfy(seg -> assertThat(seg.text().length()).isLessThanOrEqualTo(CHUNK_SIZE));
    }

    @Test
    void consecutiveChunksOverlap() {
        DocumentSplitter splitter = DocumentSplitters.recursive(CHUNK_SIZE, OVERLAP);
        String text = "palavra ".repeat(500); // 4000 chars, mesmo texto do teste de chunk size
        Document document = Document.from(text);

        List<TextSegment> segments = splitter.split(document);

        assertThat(segments.size()).isGreaterThanOrEqualTo(2);

        String firstChunk = segments.get(0).text();
        String secondChunk = segments.get(1).text();

        // O fim do primeiro chunk deve compartilhar algum sufixo/prefixo com o
        // inicio do segundo, caracterizando o overlap configurado (200 chars) —
        // mesma verificacao feita em splitter_test.go no porte Go.
        assertThat(longestSharedBoundary(firstChunk, secondChunk)).isGreaterThan(0);
    }

    /**
     * Retorna o tamanho do maior sufixo de {@code first} que tambem e
     * prefixo de {@code second}, procurando a partir de ate {@link #OVERLAP}
     * caracteres.
     */
    private int longestSharedBoundary(String first, String second) {
        int max = Math.min(OVERLAP, Math.min(first.length(), second.length()));
        for (int len = max; len > 0; len--) {
            String suffix = first.substring(first.length() - len);
            if (second.startsWith(suffix)) {
                return len;
            }
        }
        return 0;
    }

    @Test
    void textWithoutSeparatorsIsHardSplit() {
        DocumentSplitter splitter = DocumentSplitters.recursive(1000, 0);
        String text = "x".repeat(2500); // uma "palavra" gigante, sem espacos
        Document document = Document.from(text);

        List<TextSegment> segments = splitter.split(document);

        assertThat(segments.size()).isGreaterThanOrEqualTo(3);
        assertThat(segments).allSatisfy(seg -> assertThat(seg.text().length()).isLessThanOrEqualTo(1000));
    }
}
