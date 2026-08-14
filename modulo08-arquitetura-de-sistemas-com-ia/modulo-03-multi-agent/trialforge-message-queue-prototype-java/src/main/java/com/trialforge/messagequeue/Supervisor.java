package com.trialforge.messagequeue;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Porte dos dois mecanismos do Supervisor (JS/Python):
 *
 * <p>Mecanismo 1 — decide e executa a estrategia de retry (Teorema CAP,
 * Modulo 3.4): timeout explicito, retry com limite (ate MAX_TENTATIVAS),
 * idempotencia. NAO e o padrao Saga: o ICF falhou DENTRO da propria
 * execucao, nao e um problema descoberto DEPOIS por uma etapa posterior.</p>
 *
 * <p>Mecanismo 2 — verifica consistencia (Modulo 3.2, paragrafo 88) e
 * compensa via Saga de verdade (Modulo 3.4): so o agente que ficou defasado
 * e regenerado, o outro nunca e retrabalhado.</p>
 */
final class Supervisor {

    private Supervisor() {
    }

    static String decidirEstrategiaDeRetry(Reacao reacao) {
        if (reacao.ok) {
            return "nenhum_retry_necessario";
        }
        if (reacao.resultadoCSRPreservado) {
            return "retry_icf_apenas"; // so refaz quem falhou, preserva o CSR que ja terminou
        }
        return "retry_ambos"; // sem visibilidade de quem terminou bem, precisa refazer tudo
    }

    static ResultadoRetry executarICFComRetry(DadoProtocolo dadoProtocolo, EstadoProtocolo estadoProtocolo,
                                               ConcurrentHashMap<String, DocumentoRegistro> documentosGerados,
                                               List<String> ordemDeExecucao, OpcoesFluxo opcoesFalha) {
        boolean persistente = opcoesFalha.isPersistirFalhaNoRetry();
        String ultimoErro = null;
        for (int tentativa = 1; tentativa <= TrialForgeFluxo.MAX_TENTATIVAS; tentativa++) {
            OpcoesFluxo opcoesDaTentativa = persistente
                    ? OpcoesFluxo.apenasFalhasICF(opcoesFalha.isForcarFalhaICF(), opcoesFalha.isForcarTravamentoICF())
                    : OpcoesFluxo.padrao();
            try {
                ResultadoICF resultado = TrialForgeFluxo.comTimeout(
                        () -> Agentes.agenteICF(dadoProtocolo, estadoProtocolo, documentosGerados, ordemDeExecucao, opcoesDaTentativa),
                        TrialForgeFluxo.TIMEOUT_ICF_MS, "Agente ICF");
                return new ResultadoRetry(true, resultado, null, tentativa);
            } catch (Exception erro) {
                ultimoErro = erro.getMessage();
            }
        }
        return new ResultadoRetry(false, null, ultimoErro, TrialForgeFluxo.MAX_TENTATIVAS);
    }

    static Verificacao verificarConsistencia(ResultadoICF resultadoICF, ResultadoCSR resultadoCSR) {
        boolean icfConsistente = resultadoICF.versaoUsada() == resultadoICF.versaoAoConcluir();
        boolean csrConsistente = resultadoCSR.versaoUsada() == resultadoCSR.versaoAoConcluir();
        List<String> agentesDivergentes = new ArrayList<>();
        if (!icfConsistente) {
            agentesDivergentes.add("ICF");
        }
        if (!csrConsistente) {
            agentesDivergentes.add("CSR");
        }
        return new Verificacao(agentesDivergentes.isEmpty(), icfConsistente, csrConsistente, agentesDivergentes);
    }

    static CompensacaoResultado compensarDivergencia(List<String> agentesDivergentes, EstadoProtocolo estadoProtocolo,
                                                       String estudo, ConcurrentHashMap<String, DocumentoRegistro> documentosGerados,
                                                       List<String> ordemDeExecucao) throws InterruptedException {
        DadoProtocolo dadoAtual = new DadoProtocolo(estadoProtocolo.getVersao(), estadoProtocolo.getCriterios(), estudo);
        ResultadoICF resultadoICF = null;
        ResultadoCSR resultadoCSR = null;
        if (agentesDivergentes.contains("ICF")) {
            resultadoICF = Agentes.agenteICF(dadoAtual, estadoProtocolo, documentosGerados, ordemDeExecucao, OpcoesFluxo.padrao());
        }
        if (agentesDivergentes.contains("CSR")) {
            resultadoCSR = Agentes.agenteCSR(dadoAtual, estadoProtocolo, ordemDeExecucao);
        }
        return new CompensacaoResultado(resultadoICF, resultadoCSR);
    }
}
