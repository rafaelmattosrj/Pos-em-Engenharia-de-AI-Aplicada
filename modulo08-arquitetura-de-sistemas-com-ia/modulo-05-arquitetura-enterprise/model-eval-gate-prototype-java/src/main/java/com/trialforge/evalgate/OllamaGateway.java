package com.trialforge.evalgate;

import java.io.IOException;

/**
 * Abstracao das duas operacoes do Ollama usadas neste protótipo: chat (geracao)
 * e embeddings. Extraída como interface (os originais chamam o SDK do Ollama
 * diretamente) para permitir testar {@link EvalGate} com um dublê
 * determinístico, sem depender de um Ollama local rodando durante `mvn test`.
 */
public interface OllamaGateway {
    String chat(String modelo, String systemPrompt, String userPrompt) throws IOException, InterruptedException;

    double[] embed(String modelo, String texto) throws IOException, InterruptedException;
}
