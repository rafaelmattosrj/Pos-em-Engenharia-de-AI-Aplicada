package com.openrouter;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Testes de {@link MultiModelChatRunner}, cobrindo os mesmos cenarios do
 * cliente Go irmao (openrouter-multi-model-chat-go/openrouter/client_test.go):
 * sucesso, erro da API propagado com a mensagem original e chave de API
 * ausente — sem depender de chamadas HTTP reais.
 */
class MultiModelChatRunnerTest {

    private static final String PERGUNTA = "Me conte uma curiosidade sobre LLMs.";

    @Test
    void executar_retornaRespostaComSucesso() {
        MultiModelChatRunner runner = new MultiModelChatRunner(
                (modelo, mensagem) -> "resposta simulada");

        List<MultiModelChatRunner.Resultado> resultados =
                runner.executar(List.of("google/gemma-3-27b-it:free"), PERGUNTA);

        assertThat(resultados).hasSize(1);
        MultiModelChatRunner.Resultado resultado = resultados.get(0);
        assertThat(resultado.sucesso()).isTrue();
        assertThat(resultado.resposta()).isEqualTo("resposta simulada");
        assertThat(resultado.erro()).isNull();
    }

    @Test
    void executar_propagaMensagemDeErroDaApi() {
        MultiModelChatRunner runner = new MultiModelChatRunner((modelo, mensagem) -> {
            throw new RuntimeException("rate limit excedido");
        });

        List<MultiModelChatRunner.Resultado> resultados =
                runner.executar(List.of("modelo-qualquer"), PERGUNTA);

        assertThat(resultados).hasSize(1);
        MultiModelChatRunner.Resultado resultado = resultados.get(0);
        assertThat(resultado.sucesso()).isFalse();
        assertThat(resultado.resposta()).isNull();
        assertThat(resultado.erro()).contains("rate limit excedido");
    }

    @Test
    void executar_falhaEmUmModeloNaoInterrompeOsDemais() {
        MultiModelChatRunner runner = new MultiModelChatRunner((modelo, mensagem) -> {
            if (modelo.equals("modelo-com-falha")) {
                throw new RuntimeException("indisponivel");
            }
            return "ok: " + modelo;
        });

        List<MultiModelChatRunner.Resultado> resultados = runner.executar(
                List.of("modelo-com-falha", "modelo-ok"), PERGUNTA);

        assertThat(resultados).hasSize(2);
        assertThat(resultados.get(0).sucesso()).isFalse();
        assertThat(resultados.get(0).erro()).isEqualTo("indisponivel");
        assertThat(resultados.get(1).sucesso()).isTrue();
        assertThat(resultados.get(1).resposta()).isEqualTo("ok: modelo-ok");
    }

    @Test
    void executar_chamaInvokerUmaVezPorModeloComAMensagemCorreta() {
        AtomicInteger chamadas = new AtomicInteger();
        List<String> modelos = List.of(
                "google/gemma-3-27b-it:free",
                "meta-llama/llama-3.2-3b-instruct:free",
                "mistralai/mistral-7b-instruct:free");

        MultiModelChatRunner runner = new MultiModelChatRunner((modelo, mensagem) -> {
            chamadas.incrementAndGet();
            assertThat(modelos).contains(modelo);
            assertThat(mensagem).isEqualTo(PERGUNTA);
            return "ok";
        });

        List<MultiModelChatRunner.Resultado> resultados = runner.executar(modelos, PERGUNTA);

        assertThat(chamadas.get()).isEqualTo(3);
        assertThat(resultados).hasSize(3);
        assertThat(resultados).allSatisfy(r -> assertThat(r.sucesso()).isTrue());
    }

    @Test
    void executar_listaVaziaNaoInvocaModelo() {
        AtomicInteger chamadas = new AtomicInteger();
        MultiModelChatRunner runner = new MultiModelChatRunner((modelo, mensagem) -> {
            chamadas.incrementAndGet();
            return "ok";
        });

        List<MultiModelChatRunner.Resultado> resultados = runner.executar(List.of(), PERGUNTA);

        assertThat(resultados).isEmpty();
        assertThat(chamadas.get()).isZero();
    }
}
