package com.trialforge.messagequeue;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Callable;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ThreadFactory;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Porte de rodarFluxoTrialForge/rodar_fluxo_trialforge (JS/Python) — o fluxo
 * completo: Sequential (Protocolo) -&gt; Parallel (ICF+CSR reagindo ao evento
 * "protocolo:pronto") -&gt; Supervisor (retry CAP + consistencia/compensacao Saga).
 *
 * <p>Fila de mensagens: em vez do EventEmitter nativo do Node.js (JS) ou de um
 * equivalente minimo sobre asyncio (Python), este porte usa um
 * {@link ExecutorService} (threads reais) como runtime de concorrencia e o
 * {@link Barramento} (porte do EventEmitter) como camada de pub/sub por cima
 * dele — o padrao observavel (Sequential+Parallel, timeout real via corrida,
 * retry com limite, idempotencia, consistencia e compensacao Saga) e
 * identico ao original.</p>
 */
public final class TrialForgeFluxo {

    // Tempos de simulacao de trabalho, proporcionais aos timeouts reais definidos
    // no Slide 2 (30s/45s/60s) mas escalados / 100 pra manter a demo rapida — o que
    // importa aqui e a proporcao entre os tres, nao o valor absoluto em segundos.
    static final long TEMPO_PROTOCOLO_MS = 300;
    static final long TEMPO_ICF_MS = 450;
    static final long TEMPO_CSR_MS = 600;

    // A emenda do comite de etica dispara 500ms depois do protocolo:pronto —
    // depois que o ICF ja terminou (450ms), mas antes do CSR terminar (600ms).
    static final long TEMPO_EMENDA_MS = 500;

    static final Map<String, Integer> CRITERIOS_INICIAIS = Map.of("idadeMinima", 13);

    // Timeout explicito de verdade (Slide 2: 30s/45s/60s / 100), com margem de
    // seguranca sobre o tempo normal de trabalho.
    private static final double MARGEM_DE_SEGURANCA = 1.5;
    static final long TIMEOUT_PROTOCOLO_MS = Math.round(TEMPO_PROTOCOLO_MS * MARGEM_DE_SEGURANCA);
    static final long TIMEOUT_ICF_MS = Math.round(TEMPO_ICF_MS * MARGEM_DE_SEGURANCA);
    static final long TIMEOUT_CSR_MS = Math.round(TEMPO_CSR_MS * MARGEM_DE_SEGURANCA);

    // Simula um agente travado: demora bem mais que o proprio timeout, entao
    // nunca chega a responder a tempo — bem diferente de "lancou uma excecao".
    static final long TEMPO_ICF_TRAVADO_MS = TIMEOUT_ICF_MS * 3;

    // "ate tres tentativas" (Slide 2) — mesmo limite aplicado de forma
    // consistente na politica de retry do ICF.
    static final int MAX_TENTATIVAS = 3;

    private static final ExecutorService EXECUTOR = Executors.newCachedThreadPool(daemonFactory("trialforge-agente-"));
    private static final ScheduledExecutorService SCHEDULER =
            Executors.newScheduledThreadPool(2, daemonFactory("trialforge-scheduler-"));

    private TrialForgeFluxo() {
    }

    private static ThreadFactory daemonFactory(String prefixo) {
        AtomicLong contador = new AtomicLong();
        return runnable -> {
            Thread thread = new Thread(runnable, prefixo + contador.incrementAndGet());
            thread.setDaemon(true);
            return thread;
        };
    }

    /** Promise.race real contra um temporizador — submete o callable e corre contra o timeout. */
    static <T> T comTimeout(Callable<T> callable, long timeoutMs, String nomeAgente) {
        Future<T> future = EXECUTOR.submit(callable);
        return getComTimeout(future, timeoutMs, nomeAgente);
    }

    private static <T> T getComTimeout(Future<T> future, long timeoutMs, String nomeAgente) {
        try {
            return future.get(timeoutMs, TimeUnit.MILLISECONDS);
        } catch (TimeoutException e) {
            future.cancel(true);
            throw new ErroDeTimeoutException(nomeAgente, timeoutMs);
        } catch (ExecutionException e) {
            Throwable causa = e.getCause();
            if (causa instanceof RuntimeException runtimeException) {
                throw runtimeException;
            }
            throw new RuntimeException(causa);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException(e);
        }
    }

    // ---------- Reacao Parallel ao evento: PROMISE_ALL (bug) vs PROMISE_ALL_SETTLED (correcao) ----------

    private static CompletableFuture<Reacao> inscreverReacaoParalela(
            Barramento barramento, Estrategia estrategia, OpcoesFluxo opcoesFalha,
            ConcurrentHashMap<String, DocumentoRegistro> documentosGerados, List<String> ordemDeExecucao,
            EstadoProtocolo estadoProtocolo) {
        CompletableFuture<Reacao> resultadoFuture = new CompletableFuture<>();

        barramento.once("protocolo:pronto", dadoProtocolo -> {
            Future<ResultadoICF> futureICF = EXECUTOR.submit(
                    () -> Agentes.agenteICF(dadoProtocolo, estadoProtocolo, documentosGerados, ordemDeExecucao, opcoesFalha));
            Future<ResultadoCSR> futureCSR = EXECUTOR.submit(
                    () -> Agentes.agenteCSR(dadoProtocolo, estadoProtocolo, ordemDeExecucao));

            if (estrategia == Estrategia.PROMISE_ALL) {
                try {
                    ResultadoICF icf = getComTimeout(futureICF, TIMEOUT_ICF_MS, "Agente ICF");
                    ResultadoCSR csr = getComTimeout(futureCSR, TIMEOUT_CSR_MS, "Agente CSR");
                    resultadoFuture.complete(new Reacao(true, icf, csr, null, true));
                } catch (RuntimeException erro) {
                    // BUG (TP, paragrafos 68-71): equivalente a Promise.all — rejeita tudo a
                    // primeira falha, mesmo que o outro agente tenha terminado bem.
                    resultadoFuture.complete(new Reacao(false, null, null, erro.getMessage(), false));
                }
            } else {
                ResultadoICF icf = null;
                String erroICF = null;
                try {
                    icf = getComTimeout(futureICF, TIMEOUT_ICF_MS, "Agente ICF");
                } catch (RuntimeException erro) {
                    erroICF = erro.getMessage();
                }
                ResultadoCSR csr = null;
                try {
                    csr = getComTimeout(futureCSR, TIMEOUT_CSR_MS, "Agente CSR");
                } catch (RuntimeException erro) {
                    // Igual ao original: so o erro do ICF e exposto no campo erroICF.
                }
                boolean ok = icf != null && csr != null;
                resultadoFuture.complete(new Reacao(ok, icf, csr, erroICF, csr != null));
            }
        });

        return resultadoFuture;
    }

    // ---------- Fluxo completo ----------

    public static ResultadoFluxo rodarFluxo(String estudo) throws InterruptedException {
        return rodarFluxo(estudo, OpcoesFluxo.padrao());
    }

    public static ResultadoFluxo rodarFluxo(String estudo, OpcoesFluxo opcoes) throws InterruptedException {
        // Escopo por fluxo, nao classe: cada chamada tem seu proprio "banco de
        // documentos", seu proprio registro de ordem de execucao, e seu proprio
        // estado de protocolo — nao compartilhados entre estudos.
        ConcurrentHashMap<String, DocumentoRegistro> documentosGerados = new ConcurrentHashMap<>();
        List<String> ordemDeExecucao = Collections.synchronizedList(new ArrayList<>());
        EstadoProtocolo estadoProtocolo = new EstadoProtocolo(CRITERIOS_INICIAIS);
        Barramento barramento = new Barramento(EXECUTOR);

        CompletableFuture<Reacao> reacaoFuture = inscreverReacaoParalela(
                barramento, opcoes.getEstrategia(), opcoes, documentosGerados, ordemDeExecucao, estadoProtocolo); // inscreve ANTES do emit

        DadoProtocolo dadoProtocolo = comTimeout(
                () -> Agentes.agenteProtocolo(barramento, estudo, estadoProtocolo, ordemDeExecucao, opcoes, SCHEDULER),
                TIMEOUT_PROTOCOLO_MS, "Agente Protocolo"); // Sequential

        Reacao reacao = reacaoFuture.join(); // Parallel ja rodou concorrente enquanto isso resolvia
        String decisaoSupervisor = Supervisor.decidirEstrategiaDeRetry(reacao);
        int tentativasRetry = 0;

        if ("retry_icf_apenas".equals(decisaoSupervisor)) {
            // O retry de verdade, COM LIMITE (CAP): reaproveita a MESMA chave de
            // idempotencia em cada tentativa, provando que nenhuma delas duplica o documento.
            ResultadoRetry resultadoRetry = Supervisor.executarICFComRetry(
                    dadoProtocolo, estadoProtocolo, documentosGerados, ordemDeExecucao, opcoes);
            tentativasRetry = resultadoRetry.tentativasFeitas();
            if (resultadoRetry.sucesso()) {
                reacao.resultadoICF = resultadoRetry.resultado();
                reacao.ok = true;
                decisaoSupervisor = "retry_icf_executado_com_sucesso";
            } else {
                // CAP: o limite de tentativas esgotou — segue em frente sem esse resultado
                // (Disponibilidade), em vez de esperar pra sempre (Consistencia).
                reacao.ok = false;
                reacao.erroICF = resultadoRetry.erro();
                decisaoSupervisor = "retry_esgotado_seguindo_sem_icf";
            }
        }

        DocumentoRegistro registroICF = documentosGerados.get(dadoProtocolo.versao() + ":icf");
        int tentativasICF = registroICF != null ? registroICF.getTentativas() : 0;

        // Mecanismo 2: so faz sentido verificar consistencia quando os dois
        // resultados de conteudo existem de verdade (depois do mecanismo 1 ja ter
        // resolvido qualquer falha tecnica).
        Verificacao verificacao = null;
        if (reacao.ok && reacao.resultadoICF != null && reacao.resultadoCSR != null) {
            verificacao = Supervisor.verificarConsistencia(reacao.resultadoICF, reacao.resultadoCSR);
            if (!verificacao.consistente()) {
                CompensacaoResultado compensacaoSaga = Supervisor.compensarDivergencia(
                        verificacao.agentesDivergentes(), estadoProtocolo, estudo, documentosGerados, ordemDeExecucao);
                if (compensacaoSaga.resultadoICF() != null) {
                    reacao.resultadoICF = compensacaoSaga.resultadoICF();
                }
                if (compensacaoSaga.resultadoCSR() != null) {
                    reacao.resultadoCSR = compensacaoSaga.resultadoCSR();
                }
                decisaoSupervisor = "saga_compensacao_"
                        + String.join("_e_", verificacao.agentesDivergentes()).toLowerCase();
            }
        }

        return new ResultadoFluxo(dadoProtocolo, reacao, decisaoSupervisor, documentosGerados, tentativasICF,
                tentativasRetry, List.copyOf(ordemDeExecucao), estadoProtocolo, verificacao);
    }
}
