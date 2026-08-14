package com.ollama;

import dev.langchain4j.model.chat.ChatModel;

import java.util.ArrayList;
import java.util.List;

/**
 * Executa uma sequencia de perguntas contra um {@link ChatModel}, capturando
 * sucesso/erro e duracao de cada chamada individualmente — uma falha em uma
 * pergunta nao interrompe as demais.
 *
 * Extraido de Main.java para tornar a logica de orquestracao testavel sem
 * depender de uma instancia real do Ollama (equivalente ao pacote `ollama`
 * do porte Go, testado via servidor HTTP em memoria).
 */
public class ChatSession {

    /**
     * Resultado de uma pergunta: {@code erro == null} indica sucesso, com a
     * resposta preenchida em {@code resposta}; caso contrario, {@code erro}
     * contem a mensagem de falha e {@code resposta} e {@code null}.
     */
    public record Resultado(String pergunta, String resposta, long duracaoMs, String erro) {
        public boolean sucesso() {
            return erro == null;
        }
    }

    private final ChatModel chatModel;

    public ChatSession(ChatModel chatModel) {
        this.chatModel = chatModel;
    }

    public List<Resultado> perguntar(List<String> perguntas) {
        List<Resultado> resultados = new ArrayList<>();

        for (String pergunta : perguntas) {
            long inicio = System.currentTimeMillis();
            try {
                String resposta = chatModel.chat(pergunta);
                long duracao = System.currentTimeMillis() - inicio;
                resultados.add(new Resultado(pergunta, resposta, duracao, null));
            } catch (Exception e) {
                long duracao = System.currentTimeMillis() - inicio;
                resultados.add(new Resultado(pergunta, null, duracao, e.getMessage()));
            }
        }

        return resultados;
    }
}
