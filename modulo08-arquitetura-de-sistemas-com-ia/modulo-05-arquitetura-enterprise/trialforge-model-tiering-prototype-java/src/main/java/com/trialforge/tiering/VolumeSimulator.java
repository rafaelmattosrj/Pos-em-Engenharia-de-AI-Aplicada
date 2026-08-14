package com.trialforge.tiering;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

/**
 * Volume concorrente (extra, não faz parte da demo gravada — Missão Prática):
 * testa o Gateway sob pressão real, várias requisições disparadas ao mesmo
 * tempo (aqui, via {@link ExecutorService} com N threads reais, em vez do
 * {@code Promise.all} do event loop de Node) — confirma que
 * {@link OrcamentoManager#reservarOrcamento} segura o orçamento por estudo
 * mesmo com N requisições concorrentes pro mesmo estudo.
 *
 * Porte 1:1 de simularVolumeConcorrente em trialforge-model-tiering-prototype.js / .py,
 * adaptado para threads de verdade (ver comentário em {@link OrcamentoManager}).
 */
public final class VolumeSimulator {

    private VolumeSimulator() {
    }

    public record ResultadoEstudo(String estudoId, double gasto, double limite, boolean estourou) {
    }

    public static List<ResultadoEstudo> simular(CascadeGateway gateway, OrcamentoManager orcamento,
                                                  List<String> estudos, List<Callable<String>> requisicoes,
                                                  int paralelismo) throws InterruptedException {
        ExecutorService executor = Executors.newFixedThreadPool(paralelismo);
        try {
            List<Future<String>> futures = new ArrayList<>();
            for (Callable<String> requisicao : requisicoes) {
                futures.add(executor.submit(requisicao));
            }
            for (Future<String> future : futures) {
                try {
                    future.get();
                } catch (Exception e) {
                    throw new RuntimeException("Falha ao processar requisição concorrente: " + e.getMessage(), e);
                }
            }
        } finally {
            executor.shutdown();
        }

        List<ResultadoEstudo> resultados = new ArrayList<>();
        for (String estudoId : estudos) {
            double gasto = orcamento.gastoAtual(estudoId);
            double limite = orcamento.limite(estudoId);
            resultados.add(new ResultadoEstudo(estudoId, gasto, limite, gasto > limite + 1e-9));
        }
        return resultados;
    }
}
