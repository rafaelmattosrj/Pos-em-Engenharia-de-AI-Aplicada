package com.example;

import java.util.List;

/**
 * Orquestra a etapa de resposta do pipeline RAG para uma unica pergunta:
 * recebe os trechos ja recuperados do vector store, filtra os relevantes,
 * monta o prompt e chama o LLM — capturando cada desfecho possivel em um
 * {@link Resultado} em vez de imprimir diretamente, o que torna o fluxo
 * testavel sem depender de Neo4j, do modelo de embeddings ou de uma chamada
 * HTTP real ao OpenRouter.
 *
 * Extraido de Main.java; equivalente, em espirito, ao cliente
 * {@code openrouter} testado via httptest no porte Go irmao
 * (pdf-rag-knowledge-base-go/openrouter/client_test.go).
 */
public class RagAnswerService {

    public enum Tipo {
        /** Nenhum resultado retornado pela busca por similaridade. */
        SEM_RESULTADOS,
        /** Havia resultados, mas nenhum atingiu o score minimo de relevancia. */
        SEM_CONTEXTO_RELEVANTE,
        /** O LLM respondeu com sucesso. */
        SUCESSO,
        /** A chamada ao LLM falhou. */
        FALHA
    }

    public record Resultado(String pergunta, Tipo tipo, String resposta, String erro) {
        public static Resultado semResultados(String pergunta) {
            return new Resultado(pergunta, Tipo.SEM_RESULTADOS, null, null);
        }

        public static Resultado semContextoRelevante(String pergunta) {
            return new Resultado(pergunta, Tipo.SEM_CONTEXTO_RELEVANTE, null, null);
        }

        public static Resultado sucesso(String pergunta, String resposta) {
            return new Resultado(pergunta, Tipo.SUCESSO, resposta, null);
        }

        public static Resultado falha(String pergunta, String erro) {
            return new Resultado(pergunta, Tipo.FALHA, null, erro);
        }
    }

    @FunctionalInterface
    public interface ChatFn {
        String chat(String prompt) throws Exception;
    }

    private static final int MAX_ERROR_LENGTH = 200;

    private final ChatFn chat;

    public RagAnswerService(ChatFn chat) {
        this.chat = chat;
    }

    public Resultado responder(String pergunta, List<RagContextBuilder.Trecho> matches) {
        if (matches.isEmpty()) {
            return Resultado.semResultados(pergunta);
        }

        String contexto = RagContextBuilder.build(matches);
        if (contexto.isBlank()) {
            return Resultado.semContextoRelevante(pergunta);
        }

        String prompt = PromptBuilder.build(pergunta, contexto);
        try {
            String resposta = chat.chat(prompt);
            return Resultado.sucesso(pergunta, resposta);
        } catch (Exception e) {
            return Resultado.falha(pergunta, ErrorMessages.truncate(e.getMessage(), MAX_ERROR_LENGTH));
        }
    }
}
