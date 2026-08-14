package com.ollama;

import io.github.cdimascio.dotenv.Dotenv;

/**
 * Configuracao de conexao com o Ollama, lida do arquivo .env (ou dos
 * valores padrao quando o arquivo nao existir / as variaveis nao
 * estiverem definidas).
 *
 * Extraido de Main.java para permitir reuso e, futuramente, testes
 * isolados da logica de configuracao.
 */
public record ChatConfig(String baseUrl, String modelName) {

    private static final String DEFAULT_BASE_URL = "http://localhost:11434/v1";
    private static final String DEFAULT_MODEL_NAME = "llama3.2";

    public static ChatConfig load() {
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing()   // nao falha se .env nao existir
                .load();

        String baseUrl = dotenv.get("OLLAMA_BASE_URL", DEFAULT_BASE_URL);
        String modelName = dotenv.get("OLLAMA_MODEL", DEFAULT_MODEL_NAME);

        return new ChatConfig(baseUrl, modelName);
    }
}
