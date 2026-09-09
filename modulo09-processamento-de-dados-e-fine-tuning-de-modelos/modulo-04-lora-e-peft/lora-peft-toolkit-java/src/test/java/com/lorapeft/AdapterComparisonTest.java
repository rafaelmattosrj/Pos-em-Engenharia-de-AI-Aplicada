package com.lorapeft;

import com.google.gson.JsonObject;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.nio.file.Path;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class AdapterComparisonTest {

    // Saída real capturada nesta máquina em 2026-09-04 (mesmos fixtures do JS original)
    private static final String SAIDA_REAL_SEM_ADAPTADOR = """
            ==========
            <|channel>thought
            Here's a thinking process to extract the requested information:

            1.  **Analyze the Request:** The user wants to extract three specific pieces of information from the provided text (a medical receipt/summary):
                *   Beneficiário (Beneficiary)
                *   Procedimento (Procedure)
                *   Valor (Value/Amount)

            2.  **
            ==========
            Prompt: 119 tokens, 119.302 tokens-per-sec
            Generation: 80 tokens, 52.503 tokens-per-sec
            Peak memory: 9.406 GB""";

    private static final String SAIDA_REAL_COM_ADAPTADOR = """
            ==========
            {"beneficiario":"Felipe Alves Monteiro","procedimento":"consulta de clinica geral","valor":3450}
            ==========
            Prompt: 119 tokens, 161.590 tokens-per-sec
            Generation: 28 tokens, 43.129 tokens-per-sec
            Peak memory: 9.406 GB""";

    private static Path testJsonl() {
        return Path.of("..", "mlx-data", "test.jsonl");
    }

    @Test
    void carregaOExemploRealIndice8EEORecidoDoFelipeAlvesMonteiro() throws IOException {
        AdapterComparison.ExemploTeste exemplo = AdapterComparison.carregarExemploTeste(testJsonl(), AdapterComparison.INDICE_EXEMPLO);
        assertThat(exemplo.gabarito().get("beneficiario").getAsString()).isEqualTo("Felipe Alves Monteiro");
        assertThat(exemplo.gabarito().get("procedimento").getAsString()).isEqualTo("consulta de clinica geral");
        assertThat(exemplo.gabarito().get("valor").getAsInt()).isEqualTo(3450);
    }

    @Test
    void montaArgumentosSemAdapterPathQuandoNaoInformado() {
        List<String> args = AdapterComparison.montarArgumentosGenerate("x", null, AdapterComparison.MAX_TOKENS, AdapterComparison.MODELO_BASE);
        assertThat(args).doesNotContain("--adapter-path");
    }

    @Test
    void montaArgumentosComAdapterPathQuandoInformado() {
        List<String> args = AdapterComparison.montarArgumentosGenerate("x", "/caminho/adapters", AdapterComparison.MAX_TOKENS, AdapterComparison.MODELO_BASE);
        int idx = args.indexOf("--adapter-path");
        assertThat(idx).isNotEqualTo(-1);
        assertThat(args.get(idx + 1)).isEqualTo("/caminho/adapters");
    }

    @Test
    void parseiaSaidaRealSemAdaptador() {
        AdapterComparison.ResultadoParse r = AdapterComparison.parsearSaidaGenerate(SAIDA_REAL_SEM_ADAPTADOR);
        assertThat(r.tokensGerados()).isEqualTo(80);
        assertThat(r.bateuNoLimiteDeTokens()).isTrue();
        assertThat(r.json()).isNull();
    }

    @Test
    void parseiaSaidaRealComAdaptador() {
        AdapterComparison.ResultadoParse r = AdapterComparison.parsearSaidaGenerate(SAIDA_REAL_COM_ADAPTADOR);
        assertThat(r.tokensGerados()).isEqualTo(28);
        assertThat(r.bateuNoLimiteDeTokens()).isFalse();
        assertThat(r.json().get("beneficiario").getAsString()).isEqualTo("Felipe Alves Monteiro");
        assertThat(r.json().get("procedimento").getAsString()).isEqualTo("consulta de clinica geral");
        assertThat(r.json().get("valor").getAsInt()).isEqualTo(3450);
    }

    @Test
    void compararComGabaritoBateCampoACampoQuandoJsonIgualAoGabarito() {
        AdapterComparison.ResultadoParse r = AdapterComparison.parsearSaidaGenerate(SAIDA_REAL_COM_ADAPTADOR);
        JsonObject gabarito = new JsonObject();
        gabarito.addProperty("beneficiario", "Felipe Alves Monteiro");
        gabarito.addProperty("procedimento", "consulta de clinica geral");
        gabarito.addProperty("valor", 3450);
        assertThat(AdapterComparison.compararComGabarito(r.json(), gabarito)).isTrue();
    }

    @Test
    void compararComGabaritoFalhaSeFaltarJson() {
        AdapterComparison.ResultadoParse r = AdapterComparison.parsearSaidaGenerate(SAIDA_REAL_SEM_ADAPTADOR);
        JsonObject gabarito = new JsonObject();
        gabarito.addProperty("beneficiario", "Felipe Alves Monteiro");
        assertThat(AdapterComparison.compararComGabarito(r.json(), gabarito)).isFalse();
    }

    @Test
    void carregarExemploComIndiceForaDoIntervaloLancaExcecao() {
        assertThatThrownBy(() -> AdapterComparison.carregarExemploTeste(testJsonl(), 999))
                .isInstanceOf(IllegalArgumentException.class);
    }
}
