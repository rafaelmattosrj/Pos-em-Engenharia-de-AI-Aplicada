package com.trialforge.reactagent;

import java.io.IOException;

/** Resposta HTTP nao-200 do servidor do modelo (Ollama). Carrega o status code para
 *  que a politica de retry possa distinguir erro de configuracao (ex.: 404, modelo
 *  nao encontrado) de falha transitoria de rede. */
public class ModeloHttpException extends IOException {

    private final int statusCode;

    public ModeloHttpException(int statusCode, String corpo) {
        super("Ollama retornou status " + statusCode + ": " + corpo);
        this.statusCode = statusCode;
    }

    public int statusCode() {
        return statusCode;
    }
}
