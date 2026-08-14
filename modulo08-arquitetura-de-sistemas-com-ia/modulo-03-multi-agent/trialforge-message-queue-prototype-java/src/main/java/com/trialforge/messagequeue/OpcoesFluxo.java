package com.trialforge.messagequeue;

/**
 * Porte do objeto de opcoes de {@code rodarFluxoTrialForge}/{@code rodar_fluxo_trialforge}
 * (JS/Python) — os flags que disparam cada cenario de falha simulada, mais a
 * estrategia de reacao paralela. Imutavel, construida via {@link #padrao()}
 * e os metodos {@code com*}, que retornam uma nova instancia.
 */
public final class OpcoesFluxo {

    private final Estrategia estrategia;
    private final boolean forcarFalhaProtocolo;
    private final boolean forcarFalhaICF;
    private final boolean forcarTravamentoICF;
    private final boolean comEmendaEtica;
    private final boolean persistirFalhaNoRetry;

    private OpcoesFluxo(Estrategia estrategia, boolean forcarFalhaProtocolo, boolean forcarFalhaICF,
                         boolean forcarTravamentoICF, boolean comEmendaEtica, boolean persistirFalhaNoRetry) {
        this.estrategia = estrategia;
        this.forcarFalhaProtocolo = forcarFalhaProtocolo;
        this.forcarFalhaICF = forcarFalhaICF;
        this.forcarTravamentoICF = forcarTravamentoICF;
        this.comEmendaEtica = comEmendaEtica;
        this.persistirFalhaNoRetry = persistirFalhaNoRetry;
    }

    public static OpcoesFluxo padrao() {
        return new OpcoesFluxo(Estrategia.PROMISE_ALL_SETTLED, false, false, false, false, false);
    }

    public OpcoesFluxo comEstrategia(Estrategia novaEstrategia) {
        return new OpcoesFluxo(novaEstrategia, forcarFalhaProtocolo, forcarFalhaICF, forcarTravamentoICF, comEmendaEtica, persistirFalhaNoRetry);
    }

    public OpcoesFluxo forcarFalhaProtocolo() {
        return new OpcoesFluxo(estrategia, true, forcarFalhaICF, forcarTravamentoICF, comEmendaEtica, persistirFalhaNoRetry);
    }

    public OpcoesFluxo forcarFalhaICF() {
        return new OpcoesFluxo(estrategia, forcarFalhaProtocolo, true, forcarTravamentoICF, comEmendaEtica, persistirFalhaNoRetry);
    }

    public OpcoesFluxo forcarTravamentoICF() {
        return new OpcoesFluxo(estrategia, forcarFalhaProtocolo, forcarFalhaICF, true, comEmendaEtica, persistirFalhaNoRetry);
    }

    public OpcoesFluxo comEmendaEtica() {
        return new OpcoesFluxo(estrategia, forcarFalhaProtocolo, forcarFalhaICF, forcarTravamentoICF, true, persistirFalhaNoRetry);
    }

    public OpcoesFluxo persistirFalhaNoRetry() {
        return new OpcoesFluxo(estrategia, forcarFalhaProtocolo, forcarFalhaICF, forcarTravamentoICF, comEmendaEtica, true);
    }

    /** Usado pelo retry: mantem so os flags de falha do ICF, sem estrategia/emenda/protocolo. */
    static OpcoesFluxo apenasFalhasICF(boolean forcarFalhaICF, boolean forcarTravamentoICF) {
        return new OpcoesFluxo(Estrategia.PROMISE_ALL_SETTLED, false, forcarFalhaICF, forcarTravamentoICF, false, false);
    }

    public Estrategia getEstrategia() {
        return estrategia;
    }

    public boolean isForcarFalhaProtocolo() {
        return forcarFalhaProtocolo;
    }

    public boolean isForcarFalhaICF() {
        return forcarFalhaICF;
    }

    public boolean isForcarTravamentoICF() {
        return forcarTravamentoICF;
    }

    public boolean isComEmendaEtica() {
        return comEmendaEtica;
    }

    public boolean isPersistirFalhaNoRetry() {
        return persistirFalhaNoRetry;
    }
}
