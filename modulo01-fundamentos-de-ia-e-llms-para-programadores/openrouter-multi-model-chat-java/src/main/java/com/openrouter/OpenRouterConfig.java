package com.openrouter;

import io.github.cdimascio.dotenv.Dotenv;

/**
 * Configuracao de acesso ao OpenRouter, lida do arquivo .env (ou das
 * variaveis de ambiente do processo, quando o arquivo nao existir).
 *
 * Extraido de Main.java para permitir testar a validacao da chave de API
 * isoladamente — equivalente ao ErrEmptyAPIKey do porte Go
 * (openrouter-multi-model-chat-go/openrouter/client.go).
 */
public record OpenRouterConfig(String apiKey) {

    public static OpenRouterConfig load() {
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing() // nao lanca excecao se .env nao existir
                .load();

        return new OpenRouterConfig(dotenv.get("OPENROUTER_API_KEY"));
    }

    /**
     * Uma chave de API e considerada valida quando presente e nao em branco —
     * equivalente a checagem {@code c.APIKey == ""} do cliente Go.
     */
    public boolean isValid() {
        return apiKey != null && !apiKey.isBlank();
    }
}
