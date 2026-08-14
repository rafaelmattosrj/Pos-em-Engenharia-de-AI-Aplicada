package com.ollama;

import dev.langchain4j.model.chat.ChatModel;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

/**
 * Testes de {@link ChatSession}, cobrindo os mesmos cenarios do cliente Go
 * irmao (ollama-local-llm-chat-go/ollama/client_test.go):
 * sucesso, recuperacao apos falha e falha persistente. A retentativa em si
 * (maxRetries) e responsabilidade do {@link ChatModel} do LangChain4j — aqui
 * validamos a orquestracao feita por ChatSession sobre esse colaborador.
 */
@ExtendWith(MockitoExtension.class)
class ChatSessionTest {

    @Mock
    private ChatModel chatModel;

    @Test
    void perguntar_retornaRespostaComSucesso() {
        when(chatModel.chat("pergunta")).thenReturn("resposta simulada");

        ChatSession session = new ChatSession(chatModel);
        List<ChatSession.Resultado> resultados = session.perguntar(List.of("pergunta"));

        assertThat(resultados).hasSize(1);
        ChatSession.Resultado resultado = resultados.get(0);
        assertThat(resultado.sucesso()).isTrue();
        assertThat(resultado.resposta()).isEqualTo("resposta simulada");
        assertThat(resultado.erro()).isNull();
        verify(chatModel, times(1)).chat("pergunta");
    }

    @Test
    void perguntar_capturaErroSemInterromperDemaisPerguntas() {
        when(chatModel.chat("pergunta com falha")).thenThrow(new RuntimeException("falha transitoria"));
        when(chatModel.chat("pergunta ok")).thenReturn("sucesso na proxima pergunta");

        ChatSession session = new ChatSession(chatModel);
        List<ChatSession.Resultado> resultados = session.perguntar(List.of("pergunta com falha", "pergunta ok"));

        assertThat(resultados).hasSize(2);

        ChatSession.Resultado falha = resultados.get(0);
        assertThat(falha.sucesso()).isFalse();
        assertThat(falha.resposta()).isNull();
        assertThat(falha.erro()).isEqualTo("falha transitoria");

        ChatSession.Resultado sucesso = resultados.get(1);
        assertThat(sucesso.sucesso()).isTrue();
        assertThat(sucesso.resposta()).isEqualTo("sucesso na proxima pergunta");

        verify(chatModel, times(1)).chat("pergunta com falha");
        verify(chatModel, times(1)).chat("pergunta ok");
    }

    @Test
    void perguntar_falhaPersistenteFicaRegistradaComoErro() {
        when(chatModel.chat(anyString())).thenThrow(new RuntimeException("indisponivel"));

        ChatSession session = new ChatSession(chatModel);
        List<ChatSession.Resultado> resultados = session.perguntar(List.of("pergunta"));

        assertThat(resultados).hasSize(1);
        ChatSession.Resultado resultado = resultados.get(0);
        assertThat(resultado.sucesso()).isFalse();
        assertThat(resultado.resposta()).isNull();
        assertThat(resultado.erro()).isEqualTo("indisponivel");
    }

    @Test
    void perguntar_listaVaziaNaoInvocaModelo() {
        ChatSession session = new ChatSession(chatModel);
        List<ChatSession.Resultado> resultados = session.perguntar(List.of());

        assertThat(resultados).isEmpty();
        verifyNoInteractions(chatModel);
    }

    @Test
    void perguntar_invocaModeloUmaVezPorPergunta() {
        List<String> perguntas = List.of("p1", "p2", "p3");
        when(chatModel.chat(anyString())).thenReturn("ok");

        ChatSession session = new ChatSession(chatModel);
        List<ChatSession.Resultado> resultados = session.perguntar(perguntas);

        assertThat(resultados).hasSize(3);
        assertThat(resultados).allSatisfy(r -> assertThat(r.sucesso()).isTrue());
        verify(chatModel, times(1)).chat(eq("p1"));
        verify(chatModel, times(1)).chat(eq("p2"));
        verify(chatModel, times(1)).chat(eq("p3"));
    }
}
