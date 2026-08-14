package com.trialforge.messagequeue;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * Porte de agenteProtocolo/agenteICF/agenteCSR/registrarDocumento (JS/Python)
 * — Slide 2: A Arquitetura Completa.
 *
 * <p>Cada agente aqui e um metodo bloqueante comum (usa {@link Thread#sleep}
 * para simular trabalho assincrono) — quem da a eles concorrencia real e o
 * {@link java.util.concurrent.ExecutorService} que os submete como
 * {@link java.util.concurrent.Callable}, em {@link TrialForgeFluxo}.</p>
 */
final class Agentes {

    private Agentes() {
    }

    static DadoProtocolo agenteProtocolo(Barramento barramento, String estudo, EstadoProtocolo estadoProtocolo,
                                          List<String> ordemDeExecucao, OpcoesFluxo opcoes,
                                          ScheduledExecutorService scheduler) throws InterruptedException {
        Thread.sleep(TrialForgeFluxo.TEMPO_PROTOCOLO_MS);
        if (opcoes.isForcarFalhaProtocolo()) {
            throw new RuntimeException("Agente Protocolo: falha simulada ao gerar critérios.");
        }
        ordemDeExecucao.add("protocolo"); // registrado ANTES do emit: prova causal de que Sequential foi respeitado

        // Evento rico: copia do estado atual, nao uma referencia — revisoes futuras
        // no estadoProtocolo nao mudam retroativamente o que ja foi publicado aqui.
        DadoProtocolo dado = new DadoProtocolo(estadoProtocolo.getVersao(), estadoProtocolo.getCriterios(), estudo);
        barramento.emit("protocolo:pronto", dado); // nao-bloqueante: publica e segue em frente

        if (opcoes.isComEmendaEtica()) {
            // Simula uma emenda chegando de forma assincrona, independente do fluxo
            // principal — exatamente como um comite de etica revisaria um criterio
            // dias ou semanas depois, sem nenhuma relacao com o timing de ICF/CSR.
            scheduler.schedule(
                    () -> estadoProtocolo.revisar(
                            Map.of("idadeMinima", 12),
                            "Comitê de ética corrigiu a idade mínima de 13 para 12 anos."),
                    TrialForgeFluxo.TEMPO_EMENDA_MS, TimeUnit.MILLISECONDS);
        }

        return dado;
    }

    static DocumentoRegistro registrarDocumento(ConcurrentHashMap<String, DocumentoRegistro> documentosGerados,
                                                 String chave, ResultadoICFBase resultado) {
        return documentosGerados.compute(chave, (k, existente) -> {
            if (existente != null) {
                existente.incrementarTentativas();
                return existente;
            }
            return new DocumentoRegistro(resultado);
        });
    }

    static ResultadoICF agenteICF(DadoProtocolo dadoProtocolo, EstadoProtocolo estadoProtocolo,
                                   ConcurrentHashMap<String, DocumentoRegistro> documentosGerados,
                                   List<String> ordemDeExecucao, OpcoesFluxo opcoes) throws InterruptedException {
        ordemDeExecucao.add("icf:inicio"); // registrado ANTES do sleep: prova que so comeca depois do protocolo:pronto
        // forcarTravamentoICF simula o agente travado (demora alem do timeout) — bem
        // diferente de forcarFalhaICF, que simula uma excecao lancada dentro do proprio agente.
        long tempo = opcoes.isForcarTravamentoICF() ? TrialForgeFluxo.TEMPO_ICF_TRAVADO_MS : TrialForgeFluxo.TEMPO_ICF_MS;
        Thread.sleep(tempo);

        String chave = dadoProtocolo.versao() + ":icf";
        int idadeMinima = dadoProtocolo.criterios().get("idadeMinima");
        ResultadoICFBase base = new ResultadoICFBase(
                "ICF",
                "Assentimento gerado com idade mínima de " + idadeMinima + " anos (a partir da versão "
                        + dadoProtocolo.versao() + " do protocolo).",
                dadoProtocolo.criterios(),
                dadoProtocolo.versao(),
                // Conferido na hora de concluir, nao na hora de comecar — e isso que permite
                // o Supervisor perceber se uma emenda chegou enquanto o ICF ainda trabalhava.
                estadoProtocolo.getVersao());

        // O documento e gravado ANTES da falha simulada: o cenario real que retry
        // precisa tratar nao e "a tarefa nunca rodou", e "a tarefa rodou, mas a
        // confirmacao se perdeu" — daí a necessidade de idempotencia, nao so de tentar de novo.
        DocumentoRegistro registro = registrarDocumento(documentosGerados, chave, base);
        if (opcoes.isForcarFalhaICF()) {
            throw new RuntimeException("Agente ICF: falha simulada na confirmação, depois do documento já ter sido gravado.");
        }
        return base.comTentativas(registro.getTentativas());
    }

    static ResultadoCSR agenteCSR(DadoProtocolo dadoProtocolo, EstadoProtocolo estadoProtocolo,
                                   List<String> ordemDeExecucao) throws InterruptedException {
        ordemDeExecucao.add("csr:inicio"); // registrado ANTES do sleep: prova que so comeca depois do protocolo:pronto
        Thread.sleep(TrialForgeFluxo.TEMPO_CSR_MS);
        int idadeMinima = dadoProtocolo.criterios().get("idadeMinima");
        return new ResultadoCSR(
                "CSR",
                "Conformidade ICH E3 verificada com idade mínima de " + idadeMinima + " anos (a partir da versão "
                        + dadoProtocolo.versao() + " do protocolo).",
                dadoProtocolo.criterios(),
                dadoProtocolo.versao(),
                estadoProtocolo.getVersao());
    }
}
