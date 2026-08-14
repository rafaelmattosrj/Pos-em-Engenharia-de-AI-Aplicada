package com.trialforge.tiering;

/** Dublê determinístico de {@link ApprovalPrompt}: devolve uma resposta fixa, sem ler stdin. */
class FakeApprovalPrompt implements ApprovalPrompt {

    private final boolean resposta;
    String ultimoRascunho;
    int chamadas = 0;

    FakeApprovalPrompt(boolean resposta) {
        this.resposta = resposta;
    }

    @Override
    public boolean approve(String rascunho) {
        chamadas++;
        ultimoRascunho = rascunho;
        return resposta;
    }
}
