package com.ollama;

import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.openai.OpenAiChatModel;

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

    private static final List<String> PERGUNTAS = List.of(
            "Explique o que e inteligencia artificial em 3 frases simples.",
            "Qual e a diferenca entre machine learning e deep learning?",
            "O que e um LLM (Large Language Model) e como ele funciona?",
            "Cite 3 casos de uso praticos de IA na engenharia de software.",
            "O que e RAG (Retrieval-Augmented Generation) e para que serve?"
    );

    public static void main(String[] args) {
        System.out.println("=".repeat(70));
        System.out.println("  Ollama Local LLM Chat - Java + LangChain4j");
        System.out.println("=".repeat(70));
        System.out.println();

        // ── Carrega variaveis de ambiente do .env ─────────────────────────────
        ChatConfig config = ChatConfig.load();

        System.out.println("Configuracao:");
        System.out.println("  Base URL : " + config.baseUrl());
        System.out.println("  Modelo   : " + config.modelName());
        System.out.println();

        // ── Configura o cliente LangChain4j apontando para o Ollama ──────────
        // O Ollama aceita qualquer valor no campo apiKey quando a API OpenAI-
        // compatible esta habilitada, por isso usamos "ollama" como placeholder.
        ChatModel chatModel = OpenAiChatModel.builder()
                .baseUrl(config.baseUrl())
                .apiKey("ollama")       // Ollama nao exige chave real
                .modelName(config.modelName())
                .temperature(0.7)
                .maxRetries(1)
                .build();

        ChatSession session = new ChatSession(chatModel);

        // ── Loop de perguntas e respostas ─────────────────────────────────────
        for (int i = 0; i < PERGUNTAS.size(); i++) {
            String pergunta = PERGUNTAS.get(i);

            System.out.println("-".repeat(70));
            System.out.printf("[%d/%d] PERGUNTA:%n%s%n%n", i + 1, PERGUNTAS.size(), pergunta);

            ChatSession.Resultado resultado = session.perguntar(List.of(pergunta)).get(0);

            if (resultado.sucesso()) {
                System.out.println("RESPOSTA:");
                System.out.println(resultado.resposta());
                System.out.printf("%n(tempo: %d ms)%n", resultado.duracaoMs());
            } else {
                System.out.println("ERRO ao chamar o modelo: " + resultado.erro());
                System.out.println();
                System.out.println("Verifique se o Ollama esta rodando:");
                System.out.println("  ollama serve");
                System.out.println("  ollama pull " + config.modelName());
            }

            System.out.println();
        }

        System.out.println("=".repeat(70));
        System.out.println("  Sessao de chat encerrada.");
        System.out.println("=".repeat(70));
    }
}
