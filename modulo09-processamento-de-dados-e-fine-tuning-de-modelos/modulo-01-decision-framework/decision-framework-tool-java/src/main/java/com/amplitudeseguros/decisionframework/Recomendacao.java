package com.amplitudeseguros.decisionframework;

/** Equivalente ao objeto RECOMENDACAO em decision-framework-tool.js. */
public enum Recomendacao {
    FINE_TUNING("Fine-tuning vale a pena"),
    CONTINUAR_PROMPT_RAG("Continue com prompt + RAG. Fine-tuning ainda não."),
    ESPERAR("Espere acumular dado, depois treine."),
    BLOQUEADO_POR_GOVERNANCA("Bloqueado: resolva a governança do dado antes de reavaliar.");

    private final String mensagem;

    Recomendacao(String mensagem) {
        this.mensagem = mensagem;
    }

    public String mensagem() {
        return mensagem;
    }

    @Override
    public String toString() {
        return mensagem;
    }
}
