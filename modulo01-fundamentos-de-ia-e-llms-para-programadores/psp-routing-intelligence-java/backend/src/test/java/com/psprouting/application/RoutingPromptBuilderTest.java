package com.psprouting.application;

import com.psprouting.domain.PSP;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.Transaction;
import com.psprouting.domain.TransactionStatus;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * O template ({@code prompts/routing-prompt.txt}) e carregado do classpath —
 * este teste confirma que os placeholders {transaction}/{similarCases} sao
 * substituidos e que as instrucoes fixas do prompt (lista de PSPs, formato
 * de saida) chegam intactas ao LLM.
 */
class RoutingPromptBuilderTest {

    private final RoutingPromptBuilder builder = new RoutingPromptBuilder();

    @Test
    void build_incluiTransacaoECasosSimilaresNoPrompt() {
        Transaction transaction = new Transaction(
                new BigDecimal("1500.00"), PaymentMethod.CREDIT_CARD, "VISA", "SP", 20, "STREAMING");
        List<SimilarCase> similarCases = List.of(
                new SimilarCase(PSP.ADYEN, TransactionStatus.SUCCESS, new BigDecimal("1320.00"),
                        PaymentMethod.CREDIT_CARD, 0.94),
                new SimilarCase(PSP.BRADESCO, TransactionStatus.FAILED, new BigDecimal("1450.00"),
                        PaymentMethod.CREDIT_CARD, 0.89)
        );

        String prompt = builder.build(transaction, similarCases);

        assertThat(prompt).contains("Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA");
        assertThat(prompt).contains("PSP=ADYEN");
        assertThat(prompt).contains("status=SUCCESS");
        assertThat(prompt).contains("PSP=BRADESCO");
        assertThat(prompt).contains("status=FAILED");
        assertThat(prompt).doesNotContain("{transaction}");
        assertThat(prompt).doesNotContain("{similarCases}");
    }

    @Test
    void build_mantemInstrucoesFixasDoTemplate() {
        Transaction transaction = new Transaction(
                new BigDecimal("59.90"), PaymentMethod.PIX, null, "BA", 10, "STREAMING");

        String prompt = builder.build(transaction, List.of());

        assertThat(prompt).contains("ADYEN, BRASPAG, BRADESCO, SANTANDER, NUPAY, PICPAY, MERCADOPAGO");
        assertThat(prompt).contains("Responda EXCLUSIVAMENTE em JSON válido");
        assertThat(prompt).contains("(nenhum caso historico similar encontrado)");
    }
}
