package com.trialforge.messagequeue;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;

import java.util.concurrent.TimeUnit;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Porte dos 9 cenarios de rodarTestes()/rodar_testes() (JS/Python) — os
 * mesmos casos, um por metodo JUnit em vez de um contador manual de
 * passou/total.
 */
@Timeout(value = 10, unit = TimeUnit.SECONDS)
class TrialForgeFluxoTest {

    @Test
    @DisplayName("1) evento 'protocolo:pronto' carrega versao+criterios estruturados (evento rico)")
    void eventoProtocoloProntoEhRico() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo fase II");

        assertThat(resultado.getDadoProtocolo().versao()).isEqualTo(1);
        assertThat(resultado.getDadoProtocolo().criterios()).containsEntry("idadeMinima", 13);
    }

    @Test
    @DisplayName("2) caminho feliz: ICF+CSR ok, sem retry, consistência confirmada")
    void caminhoFeliz() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo fase II");

        assertThat(resultado.isOk()).isTrue();
        assertThat(resultado.getDecisaoSupervisor()).isEqualTo("nenhum_retry_necessario");
        assertThat(resultado.getVerificacao()).isNotNull();
        assertThat(resultado.getVerificacao().consistente()).isTrue();
    }

    @Test
    @DisplayName("3) BUG: com PROMISE_ALL, falha do ICF perde o resultado do CSR (parágrafo 68-71)")
    void bugPromiseAllPerdeResultadoDoCSR() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo fase II",
                OpcoesFluxo.padrao().comEstrategia(Estrategia.PROMISE_ALL).forcarFalhaICF());

        assertThat(resultado.isOk()).isFalse();
        assertThat(resultado.isResultadoCSRPreservado()).isFalse();
    }

    @Test
    @DisplayName("4) CORREÇÃO: com PROMISE_ALL_SETTLED, CSR preservado e retry do ICF executado com sucesso")
    void correcaoPromiseAllSettledPreservaCSR() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo fase II",
                OpcoesFluxo.padrao().comEstrategia(Estrategia.PROMISE_ALL_SETTLED).forcarFalhaICF());

        assertThat(resultado.isResultadoCSRPreservado()).isTrue();
        assertThat(resultado.getDecisaoSupervisor()).isEqualTo("retry_icf_executado_com_sucesso");
        assertThat(resultado.isOk()).isTrue();
    }

    @Test
    @DisplayName("4b) IDEMPOTÊNCIA: 1 documento só (não 2) após falha + retry, com 2 tentativas registradas")
    void idempotenciaDoRegistroDeDocumento() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo fase II",
                OpcoesFluxo.padrao().comEstrategia(Estrategia.PROMISE_ALL_SETTLED).forcarFalhaICF());

        assertThat(resultado.getDocumentosGeradosSize()).isEqualTo(1);
        assertThat(resultado.getTentativasICF()).isEqualTo(2);
    }

    @Test
    @DisplayName("5) Sequential respeitado (prova causal): protocolo aparece antes de icf:inicio e csr:inicio")
    void sequentialRespeitado() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo de ordem");

        var ordem = resultado.getOrdemDeExecucao();
        int idxProtocolo = ordem.indexOf("protocolo");
        int idxICF = ordem.indexOf("icf:inicio");
        int idxCSR = ordem.indexOf("csr:inicio");

        assertThat(idxProtocolo).isNotEqualTo(-1);
        assertThat(idxProtocolo).isLessThan(idxICF);
        assertThat(idxProtocolo).isLessThan(idxCSR);
    }

    @Test
    @DisplayName("6) DIVERGÊNCIA DETECTADA: ICF consistente, CSR defasado (só ele) após emenda no meio do caminho")
    void divergenciaDetectadaAposEmenda() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo com emenda",
                OpcoesFluxo.padrao().comEmendaEtica());

        Verificacao verificacao = resultado.getVerificacao();
        assertThat(verificacao).isNotNull();
        assertThat(verificacao.consistente()).isFalse();
        assertThat(verificacao.icfConsistente()).isTrue();
        assertThat(verificacao.csrConsistente()).isFalse();
        assertThat(verificacao.agentesDivergentes()).containsExactly("CSR");
    }

    @Test
    @DisplayName("7) COMPENSAÇÃO SAGA: só o CSR é regenerado com idadeMinima=12 (versão 2); ICF nunca tocado")
    void compensacaoSagaRegeneraApenasCSR() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo com emenda",
                OpcoesFluxo.padrao().comEmendaEtica());

        assertThat(resultado.getResultadoCSR().criterioUsado()).containsEntry("idadeMinima", 12);
        assertThat(resultado.getResultadoCSR().versaoUsada()).isEqualTo(2);
        assertThat(resultado.getDecisaoSupervisor()).isEqualTo("saga_compensacao_csr");
        assertThat(resultado.getEstadoProtocolo().getHistorico()).hasSize(1);
    }

    @Test
    @DisplayName("8) TIMEOUT REAL detectado e recuperado: ICF trava, retry resolve na 1ª tentativa")
    void timeoutRealDetectadoERecuperadoNoRetry() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo com agente travado",
                OpcoesFluxo.padrao().forcarTravamentoICF());

        assertThat(resultado.getErroICF()).contains("timeout");
        assertThat(resultado.getDecisaoSupervisor()).isEqualTo("retry_icf_executado_com_sucesso");
        assertThat(resultado.getTentativasRetry()).isEqualTo(1);
    }

    @Test
    @DisplayName("9) RETRY ESGOTA O LIMITE: falha persistente esgota MAX_TENTATIVAS, Supervisor segue sem o ICF")
    void retryEsgotaLimiteEmFalhaPersistente() throws InterruptedException {
        ResultadoFluxo resultado = TrialForgeFluxo.rodarFluxo("Estudo com falha persistente",
                OpcoesFluxo.padrao().forcarTravamentoICF().persistirFalhaNoRetry());

        assertThat(resultado.getDecisaoSupervisor()).isEqualTo("retry_esgotado_seguindo_sem_icf");
        assertThat(resultado.getTentativasRetry()).isEqualTo(TrialForgeFluxo.MAX_TENTATIVAS);
        assertThat(resultado.isOk()).isFalse();
    }

    @Test
    @DisplayName("Cada chamada a rodarFluxo tem escopo isolado (documentos/ordem/estado não vazam entre estudos)")
    void escopoIsoladoPorFluxo() throws InterruptedException {
        ResultadoFluxo primeiro = TrialForgeFluxo.rodarFluxo("Estudo A");
        ResultadoFluxo segundo = TrialForgeFluxo.rodarFluxo("Estudo B");

        assertThat(primeiro.getEstadoProtocolo()).isNotSameAs(segundo.getEstadoProtocolo());
        assertThat(primeiro.getDadoProtocolo().versao()).isEqualTo(1);
        assertThat(segundo.getDadoProtocolo().versao()).isEqualTo(1);
    }
}
