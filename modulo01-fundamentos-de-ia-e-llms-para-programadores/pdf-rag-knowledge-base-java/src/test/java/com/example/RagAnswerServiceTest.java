package com.example;

import org.junit.jupiter.api.Test;

import java.util.List;

import static com.example.RagContextBuilder.Trecho;
import static org.assertj.core.api.Assertions.assertThat;

/**
 * Testes de {@link RagAnswerService}, cobrindo os mesmos cenarios do cliente
 * OpenRouter do porte Go irmao (pdf-rag-knowledge-base-go/openrouter/client_test.go
 * — sucesso e erro da API) mais os estados especificos do pipeline RAG (sem
 * resultados de busca / sem contexto relevante), sem depender de Neo4j, do
 * modelo de embeddings ou de uma chamada HTTP real ao OpenRouter.
 */
class RagAnswerServiceTest {

    @Test
    void responder_semMatchesRetornaSemResultados() {
        RagAnswerService service = new RagAnswerService(prompt -> "não deveria ser chamado");

        RagAnswerService.Resultado resultado = service.responder("pergunta", List.of());

        assertThat(resultado.tipo()).isEqualTo(RagAnswerService.Tipo.SEM_RESULTADOS);
        assertThat(resultado.resposta()).isNull();
        assertThat(resultado.erro()).isNull();
    }

    @Test
    void responder_matchesAbaixoDoScoreMinimoRetornaSemContextoRelevante() {
        RagAnswerService service = new RagAnswerService(prompt -> "não deveria ser chamado");
        List<Trecho> matches = List.of(new Trecho("trecho pouco relevante", 0.2));

        RagAnswerService.Resultado resultado = service.responder("pergunta", matches);

        assertThat(resultado.tipo()).isEqualTo(RagAnswerService.Tipo.SEM_CONTEXTO_RELEVANTE);
    }

    @Test
    void responder_sucessoRetornaRespostaDoLlm() {
        RagAnswerService service = new RagAnswerService(prompt -> {
            assertThat(prompt).contains("Como converter objetos JavaScript em tensores?");
            assertThat(prompt).contains("trecho relevante");
            return "resposta gerada pelo LLM";
        });
        List<Trecho> matches = List.of(new Trecho("trecho relevante", 0.9));

        RagAnswerService.Resultado resultado = service.responder(
                "Como converter objetos JavaScript em tensores?", matches);

        assertThat(resultado.tipo()).isEqualTo(RagAnswerService.Tipo.SUCESSO);
        assertThat(resultado.resposta()).isEqualTo("resposta gerada pelo LLM");
        assertThat(resultado.erro()).isNull();
    }

    @Test
    void responder_erroDoLlmEPropagadoETruncadoEm200Caracteres() {
        String mensagemLonga = "rate limit excedido: " + "x".repeat(250);
        RagAnswerService service = new RagAnswerService(prompt -> {
            throw new RuntimeException(mensagemLonga);
        });
        List<Trecho> matches = List.of(new Trecho("trecho relevante", 0.9));

        RagAnswerService.Resultado resultado = service.responder("pergunta", matches);

        assertThat(resultado.tipo()).isEqualTo(RagAnswerService.Tipo.FALHA);
        assertThat(resultado.resposta()).isNull();
        assertThat(resultado.erro()).hasSize(200);
        assertThat(mensagemLonga).startsWith(resultado.erro());
    }

    @Test
    void responder_erroSemMensagemNaoLancaExcecao() {
        RagAnswerService service = new RagAnswerService(prompt -> {
            throw new RuntimeException((String) null);
        });
        List<Trecho> matches = List.of(new Trecho("trecho relevante", 0.9));

        RagAnswerService.Resultado resultado = service.responder("pergunta", matches);

        assertThat(resultado.tipo()).isEqualTo(RagAnswerService.Tipo.FALHA);
        assertThat(resultado.erro()).isEmpty();
    }
}
