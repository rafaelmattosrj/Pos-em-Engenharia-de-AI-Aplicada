package com.psprouting.application;

import com.psprouting.domain.PSP;
import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.Transaction;
import com.psprouting.domain.TransactionStatus;
import com.psprouting.domain.HistoricalTransaction;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Verifica que o texto natural gerado bate com o exemplo do IDEIA.md:
 * "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA, regiao SP,
 *  hora 20h, categoria STREAMING."
 */
class TransactionTextSerializerTest {

    @Test
    void serialize_comBandeiraIncluiClausulaDeBandeira() {
        Transaction transaction = new Transaction(
                new BigDecimal("1500.00"), PaymentMethod.CREDIT_CARD, "VISA", "SP", 20, "STREAMING");

        String text = TransactionTextSerializer.serialize(transaction);

        assertThat(text).isEqualTo(
                "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA, regiao SP, hora 20h, categoria STREAMING.");
    }

    @Test
    void serialize_semBandeiraOmiteClausulaDeBandeira() {
        Transaction transaction = new Transaction(
                new BigDecimal("59.90"), PaymentMethod.PIX, null, "BA", 10, "STREAMING");

        String text = TransactionTextSerializer.serialize(transaction);

        assertThat(text).isEqualTo("Pagamento via PIX, valor R$59.90, regiao BA, hora 10h, categoria STREAMING.");
        assertThat(text).doesNotContain("bandeira");
    }

    @Test
    void serialize_arredondaValorParaDuasCasasDecimais() {
        Transaction transaction = new Transaction(
                new BigDecimal("19.9"), PaymentMethod.WALLET, "PICPAY", "MG", 8, "GAMING");

        String text = TransactionTextSerializer.serialize(transaction);

        assertThat(text).contains("valor R$19.90");
    }

    @Test
    void serialize_historicalTransactionUsaOsMesmosCamposDeTransaction() {
        HistoricalTransaction historical = new HistoricalTransaction(
                new BigDecimal("1320.00"), PaymentMethod.CREDIT_CARD, "VISA", "SP", 22, "STREAMING",
                PSP.ADYEN, TransactionStatus.SUCCESS);

        String text = TransactionTextSerializer.serialize(historical);

        assertThat(text).isEqualTo(
                "Pagamento via CREDIT_CARD, valor R$1320.00, bandeira VISA, regiao SP, hora 22h, categoria STREAMING.");
    }
}
