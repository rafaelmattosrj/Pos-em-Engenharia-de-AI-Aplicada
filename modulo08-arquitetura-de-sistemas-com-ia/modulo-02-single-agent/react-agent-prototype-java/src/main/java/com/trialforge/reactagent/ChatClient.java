package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.util.List;

/**
 * Abstracao da chamada "Pensamento" do loop ReAct: envia o historico + o schema
 * de ferramentas disponiveis, recebe de volta ou uma resposta final, ou uma
 * chamada de ferramenta a executar.
 *
 * Extraida como interface (o original em JS/Python chama o SDK do Ollama
 * diretamente) para permitir testar {@link AgenteICF} com um dublê determinístico,
 * sem depender de um Ollama local rodando durante {@code mvn test}.
 * {@link OllamaReactClient} e a implementacao real; {@link RetryingChatClient}
 * decora qualquer implementacao com retry+backoff.
 */
public interface ChatClient {
    ModeloResposta chat(List<ObjectNode> historico, List<ObjectNode> tools) throws IOException, InterruptedException;
}
