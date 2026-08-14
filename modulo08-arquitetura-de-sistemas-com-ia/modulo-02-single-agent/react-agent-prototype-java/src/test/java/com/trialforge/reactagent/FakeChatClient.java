package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;

/**
 * Dublê determinístico de {@link ChatClient}, usado nos testes de {@link AgenteICF} e
 * {@link RetryingChatClient} — evita depender de um Ollama local rodando durante
 * {@code mvn test}, igual ao {@code FakeClassifierClient} do projeto irmão
 * manipulation-guardrail-prototype-java.
 */
class FakeChatClient implements ChatClient {

    /** Uma resposta enfileirada: ou uma {@link ModeloResposta} de sucesso, ou uma exceção a lançar. */
    sealed interface Evento permits Evento.Sucesso, Evento.Falha {
        record Sucesso(ModeloResposta resposta) implements Evento {
        }

        record Falha(IOException erro) implements Evento {
        }
    }

    private final Deque<Evento> eventos = new ArrayDeque<>();
    int chamadas = 0;

    FakeChatClient comResposta(ModeloResposta resposta) {
        eventos.add(new Evento.Sucesso(resposta));
        return this;
    }

    FakeChatClient comFalha(IOException erro) {
        eventos.add(new Evento.Falha(erro));
        return this;
    }

    @Override
    public ModeloResposta chat(List<ObjectNode> historico, List<ObjectNode> tools) throws IOException {
        chamadas++;
        if (eventos.isEmpty()) {
            throw new IllegalStateException("FakeChatClient: nenhum evento enfileirado para a chamada " + chamadas);
        }
        Evento evento = eventos.poll();
        if (evento instanceof Evento.Falha falha) {
            throw falha.erro();
        }
        return ((Evento.Sucesso) evento).resposta();
    }
}
