package com.example;

import dev.langchain4j.data.document.Document;
import dev.langchain4j.data.document.DocumentSplitter;
import dev.langchain4j.data.document.parser.apache.pdfbox.ApachePdfBoxDocumentParser;
import dev.langchain4j.data.document.splitter.DocumentSplitters;
import dev.langchain4j.data.embedding.Embedding;
import dev.langchain4j.data.segment.TextSegment;
import dev.langchain4j.community.store.embedding.neo4j.Neo4jEmbeddingStore;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.embedding.onnx.allminilml6v2.AllMiniLmL6V2EmbeddingModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import dev.langchain4j.store.embedding.EmbeddingMatch;
import dev.langchain4j.store.embedding.EmbeddingSearchRequest;
import io.github.cdimascio.dotenv.Dotenv;
import org.neo4j.driver.AuthTokens;
import org.neo4j.driver.Driver;
import org.neo4j.driver.GraphDatabase;
import org.neo4j.driver.Session;

import java.nio.file.Path;
import java.util.List;

import static dev.langchain4j.data.document.loader.FileSystemDocumentLoader.loadDocument;

public class Main {

    public static void main(String[] args) {
        System.out.println("🚀 Inicializando sistema de RAG com Neo4j (Java)...\n");

        // ─── Carrega variáveis de ambiente do .env ────────────────────────────
        Dotenv dotenv = Dotenv.load();
        String neo4jUri      = dotenv.get("NEO4J_URI");
        String neo4jUser     = dotenv.get("NEO4J_USER");
        String neo4jPassword = dotenv.get("NEO4J_PASSWORD");
        String apiKey        = dotenv.get("OPENROUTER_API_KEY");
        String nlpModel      = dotenv.get("NLP_MODEL");

        // ─── ETAPA 1: Carrega e divide o PDF ─────────────────────────────────
        System.out.println("📄 Carregando PDF (tensores.pdf)...");
        Document document = loadDocument(
                Path.of("tensores.pdf"),
                new ApachePdfBoxDocumentParser()
        );
        System.out.println("✅ PDF carregado com sucesso");

        DocumentSplitter splitter = DocumentSplitters.recursive(1000, 200);
        List<TextSegment> segments = splitter.split(document);
        System.out.printf("✂️  Dividido em %d chunks%n%n", segments.size());

        // ─── ETAPA 2: Carrega modelo de embeddings local ──────────────────────
        System.out.println("🧠 Carregando modelo de embeddings local (all-MiniLM-L6-v2)...");
        EmbeddingModel embeddingModel = new AllMiniLmL6V2EmbeddingModel();
        System.out.println("✅ Modelo de embeddings carregado\n");

        // ─── ETAPA 3: Conecta ao Neo4j e limpa dados anteriores ──────────────
        System.out.println("🗑️  Removendo documentos existentes no Neo4j...");
        try (Driver driver = GraphDatabase.driver(neo4jUri, AuthTokens.basic(neo4jUser, neo4jPassword));
             Session session = driver.session()) {
            session.run("MATCH (n:Chunk) DETACH DELETE n");
            try {
                session.run("DROP INDEX tensors_index IF EXISTS");
            } catch (Exception e) {
                // índice pode não existir na primeira execução
            }
            System.out.println("✅ Dados anteriores removidos\n");
        }

        // ─── ETAPA 4: Cria o store e ingere os chunks ─────────────────────────
        Neo4jEmbeddingStore embeddingStore = Neo4jEmbeddingStore.builder()
                .withBasicAuth(neo4jUri, neo4jUser, neo4jPassword)
                .dimension(384)
                .label("Chunk")
                .indexName("tensors_index")
                .build();

        System.out.println("📥 Gerando embeddings e armazenando no Neo4j...");
        for (int i = 0; i < segments.size(); i++) {
            TextSegment segment = segments.get(i);
            Embedding embedding = embeddingModel.embed(segment).content();
            embeddingStore.add(embedding, segment);
            System.out.printf("   ✅ Chunk %d/%d armazenado%n", i + 1, segments.size());
        }
        System.out.println("\n✅ Base de dados populada com sucesso!\n");

        // ─── ETAPA 5: Configura o LLM via OpenRouter ─────────────────────────
        ChatModel chatModel = OpenAiChatModel.builder()
                .baseUrl("https://openrouter.ai/api/v1")
                .apiKey(apiKey)
                .modelName(nlpModel)
                .temperature(0.3)
                .maxRetries(2)
                .build();

        RagAnswerService ragAnswerService = new RagAnswerService(chatModel::chat);

        // ─── ETAPA 6: Pipeline RAG — perguntas e respostas ───────────────────
        List<String> questions = List.of(
                "Como converter objetos JavaScript em tensores?",
                "O que é normalização de dados e por que é necessária?",
                "Como funciona uma rede neural no TensorFlow.js?",
                "O que significa treinar uma rede neural?",
                "o que é hot enconding e quando usar?"
        );

        System.out.println("🔍 ETAPA 2: Executando buscas por similaridade...\n");

        for (String question : questions) {
            System.out.println("=".repeat(80));
            System.out.println("📌 PERGUNTA: " + question);
            System.out.println("=".repeat(80));

            // Gera embedding da pergunta
            Embedding questionEmbedding = embeddingModel.embed(question).content();

            // Busca por similaridade no Neo4j
            System.out.println("🔍 Buscando no vector store do Neo4j...");
            List<EmbeddingMatch<TextSegment>> matches = embeddingStore.search(
                    EmbeddingSearchRequest.builder()
                            .queryEmbedding(questionEmbedding)
                            .maxResults(3)
                            .build()
            ).matches();

            if (matches.isEmpty()) {
                System.out.println("⚠️  Nenhum resultado encontrado na base de conhecimento.\n");
                continue;
            }

            double topScore = matches.get(0).score();
            System.out.printf("✅ Encontrados %d resultados relevantes (melhor score: %.3f)%n",
                    matches.size(), topScore);

            List<RagContextBuilder.Trecho> trechos = matches.stream()
                    .map(m -> new RagContextBuilder.Trecho(m.embedded().text(), m.score()))
                    .toList();

            RagAnswerService.Resultado resultado = ragAnswerService.responder(question, trechos);

            switch (resultado.tipo()) {
                case SEM_RESULTADOS -> System.out.println("⚠️  Nenhum resultado encontrado na base de conhecimento.\n");
                case SEM_CONTEXTO_RELEVANTE -> System.out.printf(
                        "⚠️  Nenhum resultado com score suficiente (> %.1f).%n%n", RagContextBuilder.MIN_SCORE);
                case SUCESSO -> {
                    System.out.println("🤖 Gerando resposta com IA...");
                    System.out.println("\n" + resultado.resposta() + "\n");
                }
                case FALHA -> {
                    System.out.println("🤖 Gerando resposta com IA...");
                    System.out.println("❌ Erro ao chamar LLM: " + resultado.erro() + "\n");
                }
            }
        }

        System.out.println("=".repeat(80));
        System.out.println("✅ Processamento concluído com sucesso!\n");
    }
}
