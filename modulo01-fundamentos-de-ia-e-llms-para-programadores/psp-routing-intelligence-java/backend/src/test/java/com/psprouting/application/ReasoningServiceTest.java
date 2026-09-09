package com.psprouting.application;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.psprouting.domain.PSP;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.Transaction;
import dev.langchain4j.model.chat.ChatModel;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Confirma que o {@link ReasoningService} monta o prompt (via
 * {@link RoutingPromptBuilder}), chama o {@link ChatModel} e decodifica a
 * resposta (via {@link RecommendationParser}) — usando um {@link ChatModel}
 * fake em vez de uma chamada HTTP real ao OpenRouter.
 */
class ReasoningServiceTest {

    @Test
    void reason_montaPromptChamaLlmEDecodificaResposta() {
        ChatModel fakeChatModel = new ChatModel() {
            @Override
            public String chat(String prompt) {
                assertThat(prompt).contains("Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA");
                return """
                        {
                          "primary": "ADYEN",
                          "confidence": 0.87,
                          "reasoning": "Alta aprovacao no Adyen.",
                          "fallback": ["BRASPAG"]
                        }
                        """;
            }
        };

        ReasoningService service = new ReasoningService(
                fakeChatModel, new RoutingPromptBuilder(), new RecommendationParser(new ObjectMapper()));

        Transaction transaction = new Transaction(
                new BigDecimal("1500.00"), PaymentMethod.CREDIT_CARD, "VISA", "SP", 20, "STREAMING");

        ParsedRecommendation result = service.reason(transaction, List.of());

        assertThat(result.primary()).isEqualTo(PSP.ADYEN);
        assertThat(result.confidence()).isEqualTo(0.87);
        assertThat(result.fallback()).containsExactly(PSP.BRASPAG);
    }
}
