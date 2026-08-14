package com.trialforge.agentcomponents;

import java.io.IOException;
import java.util.List;

/**
 * Abstracao do modelo de chat usado pela secao de Planejamento.
 *
 * Extraida como interface (o original em JS/Python chama o SDK do Ollama
 * diretamente) para permitir, se necessario, substituir por um dublê nos
 * testes — sem depender de um Ollama local rodando. {@link OllamaChatClient}
 * e a implementacao real, usada em producao/demo.
 */
public interface ChatClient {
    String chat(List<Memoria.Mensagem> mensagens) throws IOException, InterruptedException;
}
