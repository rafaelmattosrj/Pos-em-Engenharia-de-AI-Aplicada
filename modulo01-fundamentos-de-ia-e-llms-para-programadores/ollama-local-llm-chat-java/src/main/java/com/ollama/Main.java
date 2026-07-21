package com.ollama;

import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import io.github.cdimascio.dotenv.Dotenv;

import java.util.List;

/**
 * Demonstracao de chat com LLM local via Ollama.
 *
 * O Ollama expoe uma API OpenAI-compatible em http://localhost:11434/v1,
 * portanto usamos o cliente OpenAI do LangChain4j apontando para esse endpoint.
 *
 * Equivalente Java do exemplo em exemplo-10-ollama/request.sh.
 */
public class Main {

    public static void main(String[] args) {
        System.out.println("=".repeat(70));
        System.out.println("  Ollama Local LLM Chat - Java + LangChain4j");
        System.out.println("=".repeat(70));
        System.out.println();

        // ── Carrega variaveis de ambiente do .env ─────────────────────────────
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing()   // nao falha se .env nao existir
                .load();

        String baseUrl    = dotenv.get("OLLAMA_BASE_URL", "http://localhost:11434/v1");
        String modelName  = dotenv.get("OLLAMA_MODEL",    "llama3.2");

        System.out.println("Configuracao:");
        System.out.println("  Base URL : " + baseUrl);
        System.out.println("  Modelo   : " + modelName);
        System.out.println();

        // ── Configura o cliente LangChain4j apontando para o Ollama ──────────
        // O Ollama aceita qualquer valor no campo apiKey quando a API OpenAI-
        // compatible esta habilitada, por isso usamos "ollama" como placeholder.
        ChatModel chatModel = OpenAiChatModel.builder()
                .baseUrl(baseUrl)
                .apiKey("ollama")       // Ollama nao exige chave real
                .modelName(modelName)
                .temperature(0.7)
                .maxRetries(1)
                .build();

        // ── Perguntas de demonstracao ─────────────────────────────────────────
        List<String> perguntas = List.of(
                "Explique o que e inteligencia artificial em 3 frases simples.",
                "Qual e a diferenca entre machine learning e deep learning?",
                "O que e um LLM (Large Language Model) e como ele funciona?",
                "Cite 3 casos de uso praticos de IA na engenharia de software.",
                "O que e RAG (Retrieval-Augmented Generation) e para que serve?"
        );

        // ── Loop de perguntas e respostas ─────────────────────────────────────
        for (int i = 0; i < perguntas.size(); i++) {
            String pergunta = perguntas.get(i);

            System.out.println("-".repeat(70));
            System.out.printf("[%d/%d] PERGUNTA:%n%s%n%n", i + 1, perguntas.size(), pergunta);

            try {
                long inicio = System.currentTimeMillis();

                String resposta = chatModel.chat(pergunta);

                long duracao = System.currentTimeMillis() - inicio;

                System.out.println("RESPOSTA:");
                System.out.println(resposta);
                System.out.printf("%n(tempo: %d ms)%n", duracao);

            } catch (Exception e) {
                System.out.println("ERRO ao chamar o modelo: " + e.getMessage());
                System.out.println();
                System.out.println("Verifique se o Ollama esta rodando:");
                System.out.println("  ollama serve");
                System.out.println("  ollama pull " + modelName);
            }

            System.out.println();
        }

        System.out.println("=".repeat(70));
        System.out.println("  Sessao de chat encerrada.");
        System.out.println("=".repeat(70));
    }
}
