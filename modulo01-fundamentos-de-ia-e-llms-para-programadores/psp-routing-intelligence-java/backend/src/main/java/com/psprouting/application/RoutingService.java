package com.psprouting.application;

import com.psprouting.domain.RoutingRecommendation;
import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.Transaction;
import dev.langchain4j.data.embedding.Embedding;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Orquestra o pipeline completo descrito em IDEIA.md para
 * {@code POST /api/routing/recommend}:
 *
 * <pre>
 * (1) EmbeddingService        — serializa a transacao e gera o vetor
 * (2) SimilaritySearchService — busca as 5 transacoes historicas mais similares no Neo4j
 * (3) ReasoningService        — RAG + LLM via LangChain4j/OpenRouter
 * (4) combina a recomendacao do LLM com os casos similares na resposta final
 * </pre>
 */
@Service
public class RoutingService {

    private final EmbeddingService embeddingService;
    private final SimilaritySearchService similaritySearchService;
    private final ReasoningService reasoningService;

    public RoutingService(
            EmbeddingService embeddingService,
            SimilaritySearchService similaritySearchService,
            ReasoningService reasoningService
    ) {
        this.embeddingService = embeddingService;
        this.similaritySearchService = similaritySearchService;
        this.reasoningService = reasoningService;
    }

    public RoutingRecommendation recommend(Transaction transaction) {
        String text = TransactionTextSerializer.serialize(transaction);
        Embedding embedding = embeddingService.embed(text);

        List<SimilarCase> similarCases = similaritySearchService.findSimilar(embedding);

        ParsedRecommendation parsed = reasoningService.reason(transaction, similarCases);

        return new RoutingRecommendation(
                parsed.primary(),
                parsed.confidence(),
                parsed.reasoning(),
                parsed.fallback(),
                similarCases
        );
    }
}
