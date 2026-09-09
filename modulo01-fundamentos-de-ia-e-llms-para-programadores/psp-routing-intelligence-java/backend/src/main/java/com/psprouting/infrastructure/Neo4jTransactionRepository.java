package com.psprouting.infrastructure;

import com.psprouting.domain.HistoricalTransaction;
import com.psprouting.domain.PSP;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.TransactionStatus;
import dev.langchain4j.community.store.embedding.neo4j.Neo4jEmbeddingStore;
import dev.langchain4j.data.document.Metadata;
import dev.langchain4j.data.embedding.Embedding;
import dev.langchain4j.data.segment.TextSegment;
import dev.langchain4j.store.embedding.EmbeddingMatch;
import dev.langchain4j.store.embedding.EmbeddingSearchRequest;
import org.neo4j.driver.AuthTokens;
import org.neo4j.driver.Driver;
import org.neo4j.driver.GraphDatabase;
import org.neo4j.driver.Session;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.util.List;

/**
 * Armazena e busca transacoes historicas no Neo4j, usando o mesmo padrao de
 * {@code Neo4jEmbeddingStore} + label + indice ja usado em
 * {@code embeddings-vector-search-java} e {@code pdf-rag-knowledge-base-java}
 * — aqui o label e {@code Transaction} e o indice {@code psp_routing_index},
 * em vez de {@code Chunk}/{@code tensors_index}, porque o dominio e distinto.
 *
 * Cada transacao e guardada como um {@link TextSegment} cujo texto e a
 * descricao natural (ver {@link com.psprouting.application.TransactionTextSerializer})
 * e cujos metadados carregam {@code psp}, {@code status}, {@code amount} e
 * {@code method} — usados para reconstruir o {@link SimilarCase} na busca,
 * sem precisar reprocessar o texto.
 */
@Repository
public class Neo4jTransactionRepository {

    public static final String LABEL = "Transaction";
    public static final String INDEX_NAME = "psp_routing_index";
    public static final int DIMENSION = 384;

    private final Neo4jEmbeddingStore embeddingStore;
    private final String uri;
    private final String user;
    private final String password;

    public Neo4jTransactionRepository(
            Neo4jEmbeddingStore embeddingStore,
            @Value("${psp-routing.neo4j.uri}") String uri,
            @Value("${psp-routing.neo4j.user}") String user,
            @Value("${psp-routing.neo4j.password}") String password
    ) {
        this.embeddingStore = embeddingStore;
        this.uri = uri;
        this.user = user;
        this.password = password;
    }

    /** Remove todas as transacoes historicas armazenadas — usado antes de re-popular o seed. */
    public void clear() {
        try (Driver driver = GraphDatabase.driver(uri, AuthTokens.basic(user, password));
             Session session = driver.session()) {
            session.run("MATCH (n:" + LABEL + ") DETACH DELETE n");
            try {
                session.run("DROP INDEX " + INDEX_NAME + " IF EXISTS");
            } catch (Exception ignored) {
                // indice pode nao existir na primeira execucao
            }
        }
    }

    public void store(HistoricalTransaction transaction, String naturalText, Embedding embedding) {
        Metadata metadata = new Metadata()
                .put("psp", transaction.psp().name())
                .put("status", transaction.status().name())
                .put("amount", transaction.amount().doubleValue())
                .put("method", transaction.method().name());

        embeddingStore.add(embedding, TextSegment.from(naturalText, metadata));
    }

    /** Busca os {@code maxResults} casos historicos mais similares ao embedding informado. */
    public List<SimilarCase> findSimilar(Embedding queryEmbedding, int maxResults) {
        List<EmbeddingMatch<TextSegment>> matches = embeddingStore.search(
                EmbeddingSearchRequest.builder()
                        .queryEmbedding(queryEmbedding)
                        .maxResults(maxResults)
                        .build()
        ).matches();

        return matches.stream().map(this::toSimilarCase).toList();
    }

    private SimilarCase toSimilarCase(EmbeddingMatch<TextSegment> match) {
        Metadata metadata = match.embedded().metadata();
        return new SimilarCase(
                PSP.valueOf(metadata.getString("psp")),
                TransactionStatus.valueOf(metadata.getString("status")),
                BigDecimal.valueOf(metadata.getDouble("amount")),
                PaymentMethod.valueOf(metadata.getString("method")),
                match.score()
        );
    }
}
