package com.trialforge.guardrail;

import java.io.IOException;

/**
 * Abstracao do classificador de escopo usado pelo Guardrail.
 *
 * Extraida como interface (o original em JS/Python chama o SDK do Ollama
 * diretamente) para permitir testar {@link GuardrailGateway} com um dublê
 * determinístico, sem depender de um Ollama local rodando durante `mvn test`.
 * {@link OllamaClassifierClient} é a implementação real, usada em produção/demo.
 */
public interface ClassifierClient {
    String classify(String pergunta) throws IOException, InterruptedException;
}
