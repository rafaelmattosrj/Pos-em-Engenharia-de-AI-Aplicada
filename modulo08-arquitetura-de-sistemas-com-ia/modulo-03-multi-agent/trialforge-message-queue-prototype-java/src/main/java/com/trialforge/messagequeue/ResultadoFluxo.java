package com.trialforge.messagequeue;

import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Porte do objeto de retorno de rodarFluxoTrialForge/rodar_fluxo_trialforge
 * (JS/Python) — tudo que a demonstracao e os testes precisam inspecionar.
 */
public final class ResultadoFluxo {

    final DadoProtocolo dadoProtocolo;
    final Reacao reacao;
    final String decisaoSupervisor;
    final ConcurrentHashMap<String, DocumentoRegistro> documentosGerados;
    final int tentativasICF;
    final int tentativasRetry;
    final List<String> ordemDeExecucao;
    final EstadoProtocolo estadoProtocolo;
    final Verificacao verificacao;

    ResultadoFluxo(DadoProtocolo dadoProtocolo, Reacao reacao, String decisaoSupervisor,
                   ConcurrentHashMap<String, DocumentoRegistro> documentosGerados, int tentativasICF,
                   int tentativasRetry, List<String> ordemDeExecucao, EstadoProtocolo estadoProtocolo,
                   Verificacao verificacao) {
        this.dadoProtocolo = dadoProtocolo;
        this.reacao = reacao;
        this.decisaoSupervisor = decisaoSupervisor;
        this.documentosGerados = documentosGerados;
        this.tentativasICF = tentativasICF;
        this.tentativasRetry = tentativasRetry;
        this.ordemDeExecucao = ordemDeExecucao;
        this.estadoProtocolo = estadoProtocolo;
        this.verificacao = verificacao;
    }

    public DadoProtocolo getDadoProtocolo() {
        return dadoProtocolo;
    }

    public boolean isOk() {
        return reacao.ok;
    }

    public ResultadoICF getResultadoICF() {
        return reacao.resultadoICF;
    }

    public ResultadoCSR getResultadoCSR() {
        return reacao.resultadoCSR;
    }

    public String getErroICF() {
        return reacao.erroICF;
    }

    public boolean isResultadoCSRPreservado() {
        return reacao.resultadoCSRPreservado;
    }

    public String getDecisaoSupervisor() {
        return decisaoSupervisor;
    }

    public int getDocumentosGeradosSize() {
        return documentosGerados.size();
    }

    public int getTentativasICF() {
        return tentativasICF;
    }

    public int getTentativasRetry() {
        return tentativasRetry;
    }

    public List<String> getOrdemDeExecucao() {
        return ordemDeExecucao;
    }

    public EstadoProtocolo getEstadoProtocolo() {
        return estadoProtocolo;
    }

    public Verificacao getVerificacao() {
        return verificacao;
    }
}
