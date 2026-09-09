package com.psprouting.application;

/**
 * Lancada quando a resposta do LLM nao pode ser interpretada como o JSON de
 * recomendacao esperado (formato invalido, campos faltando, PSP fora do
 * enum) — o prompt pede JSON puro, mas LLMs eventualmente respondem com
 * markdown ou texto extra, entao a falha e tratada como um erro de negocio
 * (502 na API) em vez de propagar uma excecao de parsing generica.
 */
public class InvalidLlmResponseException extends RuntimeException {

    public InvalidLlmResponseException(String message, Throwable cause) {
        super(message, cause);
    }

    public InvalidLlmResponseException(String message) {
        super(message);
    }
}
