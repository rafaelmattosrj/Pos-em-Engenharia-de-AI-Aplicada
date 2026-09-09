package com.finetuningapi;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class GeminiConverterTest {

    private static final GeminiConverter.Exemplo EXEMPLO_AUTO = new GeminiConverter.Exemplo(
            "Extraia segurado, placa e valor do orçamento de oficina abaixo.",
            "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA Segurado: Camila Costa Ribeiro Placa do veiculo: AZS-6617 Valor total do reparo: R$ 1.780,50",
            GeminiConverter.exemploMap("Camila Costa Ribeiro", "AZS-6617", 1780.5));

    @Test
    void conversaoGeraExatamenteDoisTurnosUserEModel() {
        var convertido = GeminiConverter.converter(EXEMPLO_AUTO);
        assertThat(convertido.contents()).hasSize(2);
        assertThat(convertido.contents().get(0).role()).isEqualTo("user");
        assertThat(convertido.contents().get(1).role()).isEqualTo("model");
    }

    @Test
    void turnoDoUsuarioConcatenaInstrucaoEEntrada() {
        var convertido = GeminiConverter.converter(EXEMPLO_AUTO);
        String texto = convertido.contents().get(0).text();
        assertThat(texto).contains(EXEMPLO_AUTO.instrucao()).contains(EXEMPLO_AUTO.entrada());
    }

    @Test
    void turnoDoModeloEhOJsonExatoDaSaidaEsperada() {
        var convertido = GeminiConverter.converter(EXEMPLO_AUTO);
        Map<String, Object> saidaDecodificada = JsonUtil.parseObjeto(convertido.contents().get(1).text());
        assertThat(saidaDecodificada).containsEntry("segurado", "Camila Costa Ribeiro");
        assertThat(saidaDecodificada).containsEntry("placa", "AZS-6617");
        assertThat(((Number) saidaDecodificada.get("valor")).doubleValue()).isEqualTo(1780.5);
    }

    @Test
    void rejeitaExemploSemInstrucao() {
        assertThatThrownBy(() -> GeminiConverter.converter(new GeminiConverter.Exemplo(null, "x", Map.of())))
                .isInstanceOf(IllegalArgumentException.class).hasMessageContaining("instrucao");
    }

    @Test
    void rejeitaExemploSemEntrada() {
        assertThatThrownBy(() -> GeminiConverter.converter(new GeminiConverter.Exemplo("x", null, Map.of())))
                .hasMessageContaining("entrada");
    }

    @Test
    void rejeitaExemploSemSaida() {
        assertThatThrownBy(() -> GeminiConverter.converter(new GeminiConverter.Exemplo("x", "y", null)))
                .hasMessageContaining("saida");
    }

    @Test
    void converterTextoNaoFazJsonStringifyNaSaida() {
        var r = GeminiConverter.converterTexto("Pergunta", "Contexto", "Resposta em texto puro");
        assertThat(r.contents().get(1).text()).isEqualTo("Resposta em texto puro").doesNotStartWith("\"");
    }
}
