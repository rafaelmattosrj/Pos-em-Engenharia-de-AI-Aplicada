package com.psprouting.infrastructure.config;

import dev.langchain4j.community.store.embedding.neo4j.Neo4jEmbeddingStore;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.embedding.onnx.allminilml6v2.AllMiniLmL6V2EmbeddingModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import com.psprouting.infrastructure.Neo4jTransactionRepository;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Beans de infraestrutura descritos em IDEIA.md: {@code EmbeddingModel},
 * {@code ChatModel} e o {@code Neo4jEmbeddingStore}.
 *
 * Mesmo padrao de configuracao usado em {@code embeddings-vector-search-java}
 * (embedding local ONNX) e {@code pdf-rag-knowledge-base-java}
 * (OpenAiChatModel apontando para o OpenRouter), agora expostos como beans
 * Spring em vez de instanciados diretamente em {@code main}.
 */
@Configuration
public class AppConfig {

    @Bean
    public EmbeddingModel embeddingModel() {
        return new AllMiniLmL6V2EmbeddingModel();
    }

    @Bean
    public ChatModel chatModel(
            @Value("${psp-routing.openrouter.api-key}") String apiKey,
            @Value("${psp-routing.openrouter.model}") String modelName
    ) {
        return OpenAiChatModel.builder()
                .baseUrl("https://openrouter.ai/api/v1")
                .apiKey(apiKey)
                .modelName(modelName)
                .temperature(0.3)
                .maxRetries(2)
                .build();
    }

    @Bean
    public Neo4jEmbeddingStore neo4jEmbeddingStore(
            @Value("${psp-routing.neo4j.uri}") String uri,
            @Value("${psp-routing.neo4j.user}") String user,
            @Value("${psp-routing.neo4j.password}") String password
    ) {
        return Neo4jEmbeddingStore.builder()
                .withBasicAuth(uri, user, password)
                .dimension(Neo4jTransactionRepository.DIMENSION)
                .label(Neo4jTransactionRepository.LABEL)
                .indexName(Neo4jTransactionRepository.INDEX_NAME)
                .build();
    }
}
