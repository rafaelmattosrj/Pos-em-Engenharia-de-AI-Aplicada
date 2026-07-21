package com.openrouter;

import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import io.github.cdimascio.dotenv.Dotenv;

import java.time.Duration;
import java.util.List;

/**
 * Demonstração de chat com múltiplos modelos gratuitos via OpenRouter.
 *
 * O OpenRouter (https://openrouter.ai) é um proxy unificado que expõe dezenas de
 * modelos (OpenAI, Google, Meta, Mistral, etc.) através de uma única API compatível
 * com o formato OpenAI. Basta apontar o baseUrl para https://openrouter.ai/api/v1
 * e usar o LangChain4j OpenAiChatModel normalmente.
 *
 * Nota sobre headers específicos do OpenRouter:
 *   O OpenRouter aceita (e recomenda) os headers opcionais:
 *     - HTTP-Referer: URL do seu site/projeto
 *     - X-Title:      Nome do seu aplicativo
 *   Esses headers são usados para aparecer nas estatísticas do dashboard do OpenRouter.
 *   O LangChain4j OpenAiChatModel não oferece suporte nativo a headers HTTP customizados,
 *   por isso eles não são enviados aqui. Se necessário, use o cliente HTTP diretamente
 *   (OkHttp/HttpClient) ou um OpenAI SDK que permita customizar headers por requisição.
 */
public class Main {

    // URL base da API do OpenRouter (compatível com OpenAI)
    private static final String OPENROUTER_BASE_URL = "https://openrouter.ai/api/v1";

    // Pergunta enviada para todos os modelos
    private static final String PERGUNTA = "Me conte uma curiosidade sobre LLMs (Large Language Models).";

    // Lista de modelos gratuitos disponíveis no OpenRouter
    private static final List<String> MODELOS = List.of(
            "google/gemma-3-27b-it:free",
            "meta-llama/llama-3.2-3b-instruct:free",
            "mistralai/mistral-7b-instruct:free"
    );

    public static void main(String[] args) {
        // Carrega variáveis do arquivo .env (na raiz do projeto)
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing() // não lança exceção se .env não existir
                .load();

        String apiKey = dotenv.get("OPENROUTER_API_KEY");

        if (apiKey == null || apiKey.isBlank()) {
            System.err.println("ERRO: A variável OPENROUTER_API_KEY não foi encontrada.");
            System.err.println("Crie um arquivo .env na raiz do projeto com o conteúdo:");
            System.err.println("  OPENROUTER_API_KEY=sk-or-...");
            System.exit(1);
        }

        System.out.println("=".repeat(60));
        System.out.println("  OpenRouter Multi-Model Chat com LangChain4j");
        System.out.println("=".repeat(60));
        System.out.println("Pergunta enviada para todos os modelos:");
        System.out.println("  \"" + PERGUNTA + "\"");
        System.out.println("=".repeat(60));

        for (String modelo : MODELOS) {
            System.out.println();
            System.out.println("Modelo: " + modelo);
            System.out.println("-".repeat(60));

            try {
                String resposta = chamarModelo(apiKey, modelo, PERGUNTA);
                System.out.println(resposta);
            } catch (Exception e) {
                System.err.println("Erro ao chamar o modelo [" + modelo + "]: " + e.getMessage());
            }

            System.out.println("-".repeat(60));
        }

        System.out.println();
        System.out.println("=".repeat(60));
        System.out.println("  Fim da demonstração.");
        System.out.println("=".repeat(60));
    }

    /**
     * Cria um OpenAiChatModel apontando para o OpenRouter e envia a mensagem.
     *
     * @param apiKey  Chave da API do OpenRouter (começa com sk-or-...)
     * @param modelo  Identificador do modelo no formato "provedor/nome:variante"
     * @param mensagem Texto da pergunta
     * @return Texto da resposta do modelo
     */
    private static String chamarModelo(String apiKey, String modelo, String mensagem) {
        ChatLanguageModel chatModel = OpenAiChatModel.builder()
                .baseUrl(OPENROUTER_BASE_URL)
                .apiKey(apiKey)
                .modelName(modelo)
                .timeout(Duration.ofSeconds(60))
                // logRequests e logResponses são úteis durante o desenvolvimento
                // para inspecionar o payload JSON enviado e recebido.
                // Desative em produção para não expor dados sensíveis nos logs.
                .logRequests(false)
                .logResponses(false)
                .build();

        return chatModel.generate(mensagem);
    }
}
