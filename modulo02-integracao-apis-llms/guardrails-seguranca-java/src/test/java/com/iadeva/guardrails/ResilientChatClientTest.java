package com.iadeva.guardrails;

import com.iadeva.guardrails.support.MockOpenAiServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;
import org.springframework.ai.openai.api.OpenAiApi;
import org.springframework.ai.retry.RetryUtils;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

// Equivalente a llm/resilient_test.go — cadeia de fallback entre modelos.
class ResilientChatClientTest {

    private MockOpenAiServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.close();
        }
    }

    @Test
    void call_fallsBackToNextModelOnFailure() {
        List<String> seenModels = new CopyOnWriteArrayList<>();
        server = MockOpenAiServer.start(body -> {
            String model = String.valueOf(body.get("model"));
            seenModels.add(model);
            if ("model-a".equals(model)) {
                return new MockOpenAiServer.Response(500, "erro");
            }
            return new MockOpenAiServer.Response(200, "ok");
        });

        ResilientChatClient client = new ResilientChatClient(newChatModel(), "model-a,model-b");

        String content = client.call("sys", "user");

        assertThat(content).isEqualTo("ok");
        // O OpenAiChatModel já retenta internamente (via RetryTemplate) o mesmo modelo antes de
        // propagar a exceção — por isso "model-a" aparece múltiplas vezes antes do fallback para
        // "model-b". Não fixamos o número exato de retentativas (detalhe interno do Spring AI),
        // só que o fallback ocorreu corretamente: só model-a e model-b foram chamados, nessa ordem,
        // terminando em model-b.
        assertThat(seenModels).isNotEmpty();
        assertThat(seenModels).allMatch(m -> m.equals("model-a") || m.equals("model-b"));
        assertThat(seenModels).contains("model-a");
        assertThat(seenModels.get(seenModels.size() - 1)).isEqualTo("model-b");
    }

    @Test
    void call_throwsWhenAllModelsFail() {
        server = MockOpenAiServer.fixedResponse(500, "erro");
        ResilientChatClient client = new ResilientChatClient(newChatModel(), "model-a");

        assertThatThrownBy(() -> client.call("sys", "user"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("All fallback models failed");
    }

    private OpenAiChatModel newChatModel() {
        OpenAiApi api = new OpenAiApi(server.baseUrl(), "test-key");
        // SHORT_RETRY_TEMPLATE (backoff curto) em vez do default (backoff exponencial de
        // minutos) — os testes de falha simulam erro 500 e não devem levar minutos para rodar.
        return new OpenAiChatModel(api, OpenAiChatOptions.builder().model("placeholder").build(),
                null, RetryUtils.SHORT_RETRY_TEMPLATE);
    }
}
