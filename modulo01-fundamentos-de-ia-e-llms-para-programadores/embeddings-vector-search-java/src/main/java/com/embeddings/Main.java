package com.embeddings;

import dev.langchain4j.data.document.Document;
import dev.langchain4j.data.document.DocumentSplitter;
import dev.langchain4j.data.document.parser.apache.pdfbox.ApachePdfBoxDocumentParser;
import dev.langchain4j.data.document.splitter.DocumentSplitters;
import dev.langchain4j.data.embedding.Embedding;
import dev.langchain4j.data.segment.TextSegment;
import dev.langchain4j.community.store.embedding.neo4j.Neo4jEmbeddingStore;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.embedding.onnx.allminilml6v2.AllMiniLmL6V2EmbeddingModel;
import dev.langchain4j.store.embedding.EmbeddingMatch;
import dev.langchain4j.store.embedding.EmbeddingSearchRequest;
import io.github.cdimascio.dotenv.Dotenv;
import org.neo4j.driver.AuthTokens;
import org.neo4j.driver.Driver;
import org.neo4j.driver.GraphDatabase;
import org.neo4j.driver.Session;

import com.embeddings.util.TextPreview;

import java.nio.file.Path;
import java.util.List;

import static dev.langchain4j.data.document.loader.FileSystemDocumentLoader.loadDocument;

/**
 * Demonstração de Embeddings + Busca por Similaridade com Neo4j (sem LLM).
 *
 * Passos:
 *  1. Carrega tensores.pdf e divide em chunks (1000 chars / 200 overlap)
 *  2. Gera embeddings localmente com all-MiniLM-L6-v2 (ONNX, sem API)
 *  3. Limpa dados anteriores no Neo4j e armazena os novos vetores
 *  4. Executa busca por similaridade para 5 perguntas
 *  5. Exibe os 3 chunks mais relevantes com score e preview de texto
 *     -- NENHUMA chamada a LLM --
 */
public class Main {

    // Quantos caracteres do chunk exibir no preview
    private static final int PREVIEW_LENGTH = 300;

    public static void main(String[] args) {

        System.out.println("=".repeat(80));
        System.out.println("  Embeddings + Vector Search com Neo4j  (Java / LangChain4j)");
        System.out.println("  Modelo local: all-MiniLM-L6-v2  |  Sem chamada a LLM");
        System.out.println("=".repeat(80));
        System.out.println();

        // ── Carrega variaveis de ambiente do .env ─────────────────────────────
        Dotenv dotenv = Dotenv.load();
        String neo4jUri      = dotenv.get("NEO4J_URI");
        String neo4jUser     = dotenv.get("NEO4J_USER");
        String neo4jPassword = dotenv.get("NEO4J_PASSWORD");

        // ── ETAPA 1: Carrega e divide o PDF ───────────────────────────────────
        System.out.println("ETAPA 1: Carregando PDF (tensores.pdf)...");
        Document document = loadDocument(
                Path.of("tensores.pdf"),
                new ApachePdfBoxDocumentParser()
        );
        System.out.println("  PDF carregado com sucesso.");

        DocumentSplitter splitter = DocumentSplitters.recursive(1000, 200);
        List<TextSegment> segments = splitter.split(document);
        System.out.printf("  Dividido em %d chunks (tamanho=1000, overlap=200)%n%n", segments.size());

        // ── ETAPA 2: Carrega modelo de embeddings local ───────────────────────
        System.out.println("ETAPA 2: Carregando modelo de embeddings local (all-MiniLM-L6-v2 ONNX)...");
        EmbeddingModel embeddingModel = new AllMiniLmL6V2EmbeddingModel();
        System.out.println("  Modelo carregado. Dimensao do vetor: 384\n");

        // ── ETAPA 3: Conecta ao Neo4j e limpa dados anteriores ───────────────
        System.out.println("ETAPA 3: Limpando dados anteriores no Neo4j...");
        try (Driver driver = GraphDatabase.driver(neo4jUri, AuthTokens.basic(neo4jUser, neo4jPassword));
             Session session = driver.session()) {
            session.run("MATCH (n:Chunk) DETACH DELETE n");
            try {
                session.run("DROP INDEX tensors_index IF EXISTS");
            } catch (Exception ignored) {
                // Indice pode nao existir na primeira execucao
            }
        }
        System.out.println("  Dados anteriores removidos.\n");

        // ── ETAPA 4: Cria o store e ingere os chunks ──────────────────────────
        System.out.println("ETAPA 4: Gerando embeddings e armazenando no Neo4j...");
        Neo4jEmbeddingStore embeddingStore = Neo4jEmbeddingStore.builder()
                .withBasicAuth(neo4jUri, neo4jUser, neo4jPassword)
                .dimension(384)
                .label("Chunk")
                .indexName("tensors_index")
                .build();

        for (int i = 0; i < segments.size(); i++) {
            TextSegment segment = segments.get(i);
            Embedding embedding = embeddingModel.embed(segment).content();
            embeddingStore.add(embedding, segment);
            System.out.printf("  [%d/%d] chunk armazenado%n", i + 1, segments.size());
        }
        System.out.println("\n  Base de dados populada com sucesso!\n");

        // ── ETAPA 5: Busca por similaridade (sem LLM) ─────────────────────────
        System.out.println("ETAPA 5: Executando buscas por similaridade...\n");

        List<String> questions = List.of(
                "O que sao tensores e como sao representados em JavaScript?",
                "Como converter objetos JavaScript em tensores?",
                "O que e normalizacao de dados e por que e necessaria?",
                "Como funciona uma rede neural no TensorFlow.js?",
                "O que e hot encoding e quando usar?"
        );

        for (String question : questions) {
            System.out.println("=".repeat(80));
            System.out.println("PERGUNTA: " + question);
            System.out.println("=".repeat(80));

            // Gera embedding da pergunta usando o mesmo modelo local
            Embedding questionEmbedding = embeddingModel.embed(question).content();

            // Busca os 3 chunks mais proximos no espaco vetorial do Neo4j
            List<EmbeddingMatch<TextSegment>> matches = embeddingStore.search(
                    EmbeddingSearchRequest.builder()
                            .queryEmbedding(questionEmbedding)
                            .maxResults(3)
                            .build()
            ).matches();

            if (matches.isEmpty()) {
                System.out.println("  Nenhum resultado encontrado.\n");
                continue;
            }

            System.out.printf("  Encontrados %d resultados:%n%n", matches.size());

            for (int i = 0; i < matches.size(); i++) {
                EmbeddingMatch<TextSegment> match = matches.get(i);
                double score = match.score();
                String fullText = match.embedded().text();
                String preview = TextPreview.truncate(fullText, PREVIEW_LENGTH);

                System.out.printf("  Resultado #%d  |  Score: %.4f%n", i + 1, score);
                System.out.println("  " + "-".repeat(60));
                // Indenta o preview para melhor legibilidade
                for (String line : preview.split("\n")) {
                    System.out.println("  " + line);
                }
                System.out.println();
            }
        }

        System.out.println("=".repeat(80));
        System.out.println("  Processamento concluido.");
        System.out.println("  Nenhum LLM foi chamado — apenas embeddings + busca vetorial.");
        System.out.println("=".repeat(80));
    }
}
