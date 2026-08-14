package com.ollama;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Verifica que, na ausencia de um arquivo .env (cenario padrao no ambiente
 * de testes), ChatConfig.load() cai nos valores default documentados no
 * README — equivalente aos defaults usados em main.go (getEnv com fallback).
 */
class ChatConfigTest {

    @Test
    void load_usaValoresDefaultQuandoVariaveisNaoDefinidas() {
        ChatConfig config = ChatConfig.load();

        assertThat(config.baseUrl()).isNotBlank();
        assertThat(config.modelName()).isNotBlank();

        // Sem .env no diretorio de execucao dos testes e sem as variaveis de
        // ambiente OLLAMA_BASE_URL/OLLAMA_MODEL definidas, os defaults do
        // README devem prevalecer.
        if (System.getenv("OLLAMA_BASE_URL") == null) {
            assertThat(config.baseUrl()).isEqualTo("http://localhost:11434/v1");
        }
        if (System.getenv("OLLAMA_MODEL") == null) {
            assertThat(config.modelName()).isEqualTo("llama3.2");
        }
    }
}
