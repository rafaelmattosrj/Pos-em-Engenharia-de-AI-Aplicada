package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.net.ConnectException;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

/**
 * Cobre a mesma disciplina de ehErroTransitorio / chamarModeloComRetry em
 * react-agent-prototype.js / react_agent_prototype.py: falha transitória tenta de
 * novo (com backoff), falha terminal (ex.: 404 / modelo não encontrado) sobe direto.
 */
class RetryingChatClientTest {

    private final ObjectMapper mapper = new ObjectMapper();

    private ModeloResposta respostaFinalFake(String texto) {
        ObjectNode mensagem = mapper.createObjectNode();
        mensagem.put("role", "assistant");
        mensagem.put("content", texto);
        return new ModeloResposta(mensagem, texto, List.of());
    }

    // ---------- ehErroTransitorio: classificação pura ----------

    @Test
    void ehErroTransitorio_conexaoRecusada_eTransitorio() {
        assertThat(RetryingChatClient.ehErroTransitorio(new ConnectException("Connection refused"))).isTrue();
    }

    @Test
    void ehErroTransitorio_mensagemDeTimeout_eTransitorio() {
        assertThat(RetryingChatClient.ehErroTransitorio(new IOException("request timeout"))).isTrue();
    }

    @Test
    void ehErroTransitorio_erroHttp404_naoETransitorio() {
        assertThat(RetryingChatClient.ehErroTransitorio(new ModeloHttpException(404, "model not found"))).isFalse();
    }

    @Test
    void ehErroTransitorio_erroHttpGenerico_naoETransitorio() {
        assertThat(RetryingChatClient.ehErroTransitorio(new ModeloHttpException(500, "internal error"))).isFalse();
    }

    @Test
    void ehErroTransitorio_mensagemSemPadraoConhecido_naoETransitorio() {
        assertThat(RetryingChatClient.ehErroTransitorio(new IOException("algo inesperado aconteceu"))).isFalse();
    }

    // ---------- retry com backoff ----------

    @Test
    void chat_falhaTransitoriaSeguidaDeSucesso_retenta() throws Exception {
        FakeChatClient fake = new FakeChatClient()
                .comFalha(new ConnectException("Connection refused"))
                .comResposta(respostaFinalFake("ok na segunda tentativa"));

        RetryingChatClient client = new RetryingChatClient(fake, 3);
        ModeloResposta resposta = client.chat(List.of(), List.of());

        assertThat(resposta.content()).isEqualTo("ok na segunda tentativa");
        assertThat(fake.chamadas).isEqualTo(2);
    }

    @Test
    void chat_erroTerminal_naoRetenta() {
        FakeChatClient fake = new FakeChatClient().comFalha(new ModeloHttpException(404, "model not found"));

        RetryingChatClient client = new RetryingChatClient(fake, 3);

        assertThatThrownBy(() -> client.chat(List.of(), List.of())).isInstanceOf(ModeloHttpException.class);
        assertThat(fake.chamadas).isEqualTo(1);
    }

    @Test
    void chat_esgotaTentativas_propagaUltimoErro() {
        FakeChatClient fake = new FakeChatClient()
                .comFalha(new ConnectException("Connection refused"))
                .comFalha(new ConnectException("Connection refused"))
                .comFalha(new ConnectException("Connection refused"));

        RetryingChatClient client = new RetryingChatClient(fake, 3);

        assertThatThrownBy(() -> client.chat(List.of(), List.of())).isInstanceOf(ConnectException.class);
        assertThat(fake.chamadas).isEqualTo(3);
    }
}
