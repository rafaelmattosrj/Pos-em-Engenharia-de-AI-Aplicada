package com.trialforge.tiering;

import org.junit.jupiter.api.Test;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.offset;

class OrcamentoManagerTest {

    @Test
    void verificarOrcamento_comFolga_retornaTrue() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-teste", 0.05);
        orcamento.registrarGasto("estudo-teste", 0.045);

        // 0.045 + 0.004 = 0.049 <= 0.05
        assertThat(orcamento.verificarOrcamento("estudo-teste", 0.004)).isTrue();
    }

    @Test
    void verificarOrcamento_estourandoLimite_retornaFalse() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-teste", 0.05);
        orcamento.registrarGasto("estudo-teste", 0.045);

        // 0.045 + 0.006 = 0.051 > 0.05
        assertThat(orcamento.verificarOrcamento("estudo-teste", 0.006)).isFalse();
    }

    @Test
    void verificarOrcamento_naoMutaEstado() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-teste", 0.05);
        orcamento.registrarGasto("estudo-teste", 0.045);

        orcamento.verificarOrcamento("estudo-teste", 0.004);

        assertThat(orcamento.gastoAtual("estudo-teste")).isCloseTo(0.045, offset(1e-9));
    }

    @Test
    void reservarOrcamento_comFolga_debitaEDevolveTrue() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);

        boolean reservou = orcamento.reservarOrcamento("estudo-A", 0.011);

        assertThat(reservou).isTrue();
        assertThat(orcamento.gastoAtual("estudo-A")).isCloseTo(0.011, offset(1e-9));
    }

    @Test
    void reservarOrcamento_estourandoLimite_naoDebitaEDevolveFalse() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-B", 0.005);

        boolean reservou = orcamento.reservarOrcamento("estudo-B", 0.011);

        assertThat(reservou).isFalse();
        assertThat(orcamento.gastoAtual("estudo-B")).isCloseTo(0.0, offset(1e-9));
    }

    @Test
    void liberarSobra_devolveDiferenca() {
        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-A", 0.05);
        orcamento.reservarOrcamento("estudo-A", 0.011); // reserva pior caso (Tier1+Tier2)

        orcamento.liberarSobra("estudo-A", 0.011 - 0.001); // Tier 1 resolveu sozinho, custou só 0.001

        assertThat(orcamento.gastoAtual("estudo-A")).isCloseTo(0.001, offset(1e-9));
    }

    @Test
    void estudoDesconhecido_lancaExcecao() {
        OrcamentoManager orcamento = new OrcamentoManager();
        org.junit.jupiter.api.Assertions.assertThrows(IllegalArgumentException.class,
                () -> orcamento.verificarOrcamento("estudo-fantasma", 0.001));
    }

    /**
     * Réplica do cenário "estudo-F" de simularVolumeConcorrente: orçamento cabe
     * exatamente 2 reservas do pior caso, mas N threads reais disputam a mesma
     * conta ao mesmo tempo. reservarOrcamento deve continuar sendo a única
     * porta de entrada — nenhuma combinação de threads pode fazer o gasto
     * ultrapassar o limite.
     */
    @Test
    void reservarOrcamento_sobConcorrenciaReal_nuncaEstouraOLimite() throws InterruptedException {
        double custoPiorCaso = 0.011;
        int numeroDeThreads = 20;
        double limite = 2 * custoPiorCaso + 0.0005; // cabe exatamente 2 reservas

        OrcamentoManager orcamento = new OrcamentoManager();
        orcamento.definirOrcamento("estudo-F", limite);

        ExecutorService executor = Executors.newFixedThreadPool(numeroDeThreads);
        CountDownLatch largada = new CountDownLatch(1);
        CountDownLatch chegada = new CountDownLatch(numeroDeThreads);
        AtomicInteger reservasBemSucedidas = new AtomicInteger();

        for (int i = 0; i < numeroDeThreads; i++) {
            executor.submit(() -> {
                try {
                    largada.await();
                    if (orcamento.reservarOrcamento("estudo-F", custoPiorCaso)) {
                        reservasBemSucedidas.incrementAndGet();
                    }
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                } finally {
                    chegada.countDown();
                }
            });
        }

        largada.countDown(); // solta todas as threads ao mesmo tempo
        chegada.await();
        executor.shutdown();

        assertThat(orcamento.gastoAtual("estudo-F")).isLessThanOrEqualTo(limite + 1e-9);
        assertThat(reservasBemSucedidas.get()).isEqualTo(2); // exatamente 2 das 20 threads conseguem reservar
    }
}
