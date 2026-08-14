package com.openrouter;

/**
 * Abstrai a chamada a um modelo especifico do OpenRouter, permitindo
 * substituir a implementacao real (LangChain4j + HTTP) por um dublê nos
 * testes de {@link MultiModelChatRunner}.
 */
@FunctionalInterface
public interface ModeloInvoker {
    String complete(String modelo, String mensagem) throws Exception;
}
