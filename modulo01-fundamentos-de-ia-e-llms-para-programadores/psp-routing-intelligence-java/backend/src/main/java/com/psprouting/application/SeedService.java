package com.psprouting.application;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.psprouting.domain.HistoricalTransaction;
import com.psprouting.infrastructure.Neo4jTransactionRepository;
import dev.langchain4j.data.embedding.Embedding;
import org.springframework.core.io.ClassPathResource;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.util.List;

/**
 * Carrega {@code data/transactions-seed.json} (50 transacoes ficticias
 * representando padroes de aprovacao por PSP) e popula o Neo4j — passo
 * executado por {@code POST /api/routing/seed}, conforme "Como Rodar" no
 * IDEIA.md.
 *
 * Limpa os dados anteriores antes de re-popular, mesmo padrao adotado em
 * {@code embeddings-vector-search-java}/{@code pdf-rag-knowledge-base-java}
 * ao (re)ingerir os chunks do PDF.
 */
@Service
public class SeedService {

    private static final String SEED_PATH = "data/transactions-seed.json";

    private final Neo4jTransactionRepository repository;
    private final EmbeddingService embeddingService;
    private final ObjectMapper objectMapper;

    public SeedService(Neo4jTransactionRepository repository, EmbeddingService embeddingService, ObjectMapper objectMapper) {
        this.repository = repository;
        this.embeddingService = embeddingService;
        this.objectMapper = objectMapper;
    }

    public int seed() {
        List<HistoricalTransaction> transactions = loadSeedTransactions();

        repository.clear();

        for (HistoricalTransaction transaction : transactions) {
            String text = TransactionTextSerializer.serialize(transaction);
            Embedding embedding = embeddingService.embed(text);
            repository.store(transaction, text, embedding);
        }

        return transactions.size();
    }

    private List<HistoricalTransaction> loadSeedTransactions() {
        try {
            HistoricalTransaction[] transactions = objectMapper.readValue(
                    new ClassPathResource(SEED_PATH).getInputStream(),
                    HistoricalTransaction[].class
            );
            return List.of(transactions);
        } catch (IOException e) {
            throw new UncheckedIOException("Nao foi possivel carregar " + SEED_PATH, e);
        }
    }
}
