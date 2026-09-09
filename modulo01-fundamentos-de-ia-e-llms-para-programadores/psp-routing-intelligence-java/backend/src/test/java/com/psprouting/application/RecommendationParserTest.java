package com.psprouting.application;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.psprouting.domain.PSP;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

/**
 * O prompt pede JSON puro ("sem markdown, sem texto adicional"), mas LLMs
 * gratuitos as vezes nao obedecem — estes testes cobrem tanto o caminho
 * feliz quanto as respostas mal-formadas que o parser precisa tolerar ou
 * rejeitar com um erro claro (equivalente, em espirito, ao tratamento de
 * erro de parsing no cliente OpenRouter do porte Go irmao).
 */
class RecommendationParserTest {

    private final RecommendationParser parser = new RecommendationParser(new ObjectMapper());

    @Test
    void parse_jsonValidoEDecodificadoCorretamente() {
        String raw = """
                {
                  "primary": "ADYEN",
                  "confidence": 0.87,
                  "reasoning": "Transacoes VISA em SP tem alta aprovacao no Adyen.",
                  "fallback": ["BRASPAG", "BRADESCO"]
                }
                """;

        ParsedRecommendation result = parser.parse(raw);

        assertThat(result.primary()).isEqualTo(PSP.ADYEN);
        assertThat(result.confidence()).isEqualTo(0.87);
        assertThat(result.reasoning()).contains("Adyen");
        assertThat(result.fallback()).containsExactly(PSP.BRASPAG, PSP.BRADESCO);
    }

    @Test
    void parse_removeCercaMarkdownAntesDeDecodificar() {
        String raw = """
                ```json
                {
                  "primary": "PICPAY",
                  "confidence": 0.95,
                  "reasoning": "Carteira PicPay.",
                  "fallback": []
                }
                ```
                """;

        ParsedRecommendation result = parser.parse(raw);

        assertThat(result.primary()).isEqualTo(PSP.PICPAY);
        assertThat(result.fallback()).isEmpty();
    }

    @Test
    void parse_semCampoFallbackRetornaListaVazia() {
        String raw = """
                { "primary": "BRASPAG", "confidence": 0.7, "reasoning": "PIX." }
                """;

        ParsedRecommendation result = parser.parse(raw);

        assertThat(result.fallback()).isEmpty();
    }

    @Test
    void parse_jsonInvalidoLancaInvalidLlmResponseException() {
        assertThatThrownBy(() -> parser.parse("isso nao e json"))
                .isInstanceOf(InvalidLlmResponseException.class)
                .hasMessageContaining("nao e um JSON valido");
    }

    @Test
    void parse_semCamposObrigatoriosLancaInvalidLlmResponseException() {
        assertThatThrownBy(() -> parser.parse("{\"reasoning\": \"faltam campos\"}"))
                .isInstanceOf(InvalidLlmResponseException.class)
                .hasMessageContaining("campos obrigatorios");
    }

    @Test
    void parse_pspDesconhecidoLancaInvalidLlmResponseException() {
        String raw = """
                { "primary": "PAYPAL", "confidence": 0.5, "reasoning": "PSP fora do dominio." }
                """;

        assertThatThrownBy(() -> parser.parse(raw))
                .isInstanceOf(InvalidLlmResponseException.class)
                .hasMessageContaining("PSP desconhecido");
    }
}
