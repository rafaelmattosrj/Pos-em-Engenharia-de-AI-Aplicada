package com.psprouting.application;

import com.psprouting.domain.SimilarCase;
import com.psprouting.infrastructure.Neo4jTransactionRepository;
import dev.langchain4j.data.embedding.Embedding;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Passo (2) do pipeline: busca no Neo4j as transacoes historicas mais
 * similares (cosine similarity) ao embedding da transacao atual.
 */
@Service
public class SimilaritySearchService {

    /** Numero de casos similares recuperados por consulta — "5 transacoes" no IDEIA.md. */
    public static final int TOP_K = 5;

    private final Neo4jTransactionRepository repository;

    public SimilaritySearchService(Neo4jTransactionRepository repository) {
        this.repository = repository;
    }

    public List<SimilarCase> findSimilar(Embedding embedding) {
        return repository.findSimilar(embedding, TOP_K);
    }
}
