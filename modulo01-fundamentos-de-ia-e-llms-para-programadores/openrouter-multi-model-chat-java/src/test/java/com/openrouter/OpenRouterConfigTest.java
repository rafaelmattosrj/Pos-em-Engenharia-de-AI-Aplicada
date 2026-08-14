package com.openrouter;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre a mesma checagem de chave de API ausente do cliente Go irmao
 * (openrouter-multi-model-chat-go/openrouter/client.go — ErrEmptyAPIKey).
 */
class OpenRouterConfigTest {

    @Test
    void isValid_falseQuandoChaveNula() {
        OpenRouterConfig config = new OpenRouterConfig(null);
        assertThat(config.isValid()).isFalse();
    }

    @Test
    void isValid_falseQuandoChaveEmBranco() {
        OpenRouterConfig config = new OpenRouterConfig("   ");
        assertThat(config.isValid()).isFalse();
    }

    @Test
    void isValid_trueQuandoChavePresente() {
        OpenRouterConfig config = new OpenRouterConfig("sk-or-v1-teste");
        assertThat(config.isValid()).isTrue();
    }
}
