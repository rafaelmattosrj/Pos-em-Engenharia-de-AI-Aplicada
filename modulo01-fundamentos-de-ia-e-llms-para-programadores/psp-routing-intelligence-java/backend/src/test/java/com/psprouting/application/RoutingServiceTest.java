package com.psprouting.application;

import com.psprouting.domain.PSP;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.RoutingRecommendation;
import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.Transaction;
import com.psprouting.domain.TransactionStatus;
import dev.langchain4j.data.embedding.Embedding;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Verifica a orquestracao dos 4 passos do pipeline descritos em IDEIA.md
 * (embedding -> busca de similaridade -> raciocinio do LLM -> combinacao da
 * resposta final), sem depender de Neo4j, do modelo de embeddings ou de uma
 * chamada HTTP real ao OpenRouter.
 */
@ExtendWith(MockitoExtension.class)
class RoutingServiceTest {

    @Mock
    private EmbeddingService embeddingService;
    @Mock
    private SimilaritySearchService similaritySearchService;
    @Mock
    private ReasoningService reasoningService;
    @Mock
    private Embedding embedding;

    @Test
    void recommend_combinaRecomendacaoDoLlmComOsCasosSimilaresEncontrados() {
        RoutingService routingService = new RoutingService(embeddingService, similaritySearchService, reasoningService);

        Transaction transaction = new Transaction(
                new BigDecimal("1500.00"), PaymentMethod.CREDIT_CARD, "VISA", "SP", 20, "STREAMING");

        List<SimilarCase> similarCases = List.of(
                new SimilarCase(PSP.ADYEN, TransactionStatus.SUCCESS, new BigDecimal("1320.00"),
                        PaymentMethod.CREDIT_CARD, 0.94)
        );
        ParsedRecommendation parsed = new ParsedRecommendation(
                PSP.ADYEN, 0.87, "Alta aprovacao no Adyen para esse perfil.", List.of(PSP.BRASPAG, PSP.BRADESCO));

        when(embeddingService.embed(
                "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA, regiao SP, hora 20h, categoria STREAMING."))
                .thenReturn(embedding);
        when(similaritySearchService.findSimilar(embedding)).thenReturn(similarCases);
        when(reasoningService.reason(eq(transaction), eq(similarCases))).thenReturn(parsed);

        RoutingRecommendation result = routingService.recommend(transaction);

        assertThat(result.primary()).isEqualTo(PSP.ADYEN);
        assertThat(result.confidence()).isEqualTo(0.87);
        assertThat(result.reasoning()).isEqualTo("Alta aprovacao no Adyen para esse perfil.");
        assertThat(result.fallback()).containsExactly(PSP.BRASPAG, PSP.BRADESCO);
        assertThat(result.similarCases()).isEqualTo(similarCases);

        verify(similaritySearchService).findSimilar(embedding);
        verify(reasoningService).reason(transaction, similarCases);
    }

    @Test
    void recommend_funcionaMesmoSemCasosSimilares() {
        RoutingService routingService = new RoutingService(embeddingService, similaritySearchService, reasoningService);

        Transaction transaction = new Transaction(
                new BigDecimal("9.90"), PaymentMethod.WALLET, "MERCADOPAGO", "RJ", 15, "NEWS");
        ParsedRecommendation parsed = new ParsedRecommendation(
                PSP.MERCADOPAGO, 0.6, "Unico PSP compativel com carteira MercadoPago.", List.of());

        when(embeddingService.embed(any())).thenReturn(embedding);
        when(similaritySearchService.findSimilar(embedding)).thenReturn(List.of());
        when(reasoningService.reason(eq(transaction), eq(List.of()))).thenReturn(parsed);

        RoutingRecommendation result = routingService.recommend(transaction);

        assertThat(result.similarCases()).isEmpty();
        assertThat(result.primary()).isEqualTo(PSP.MERCADOPAGO);
    }
}
