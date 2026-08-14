package com.trialforge.messagequeue;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Porte de criarEstadoProtocolo/revisarProtocolo (JS/Python) — o Protocolo e
 * um recurso VERSIONADO, nao um valor congelado: a versao anterior fica
 * preservada no historico, nunca apagada.
 */
class EstadoProtocoloTest {

    @Test
    void comecaNaVersao1ComOsCriteriosIniciais() {
        EstadoProtocolo estado = new EstadoProtocolo(Map.of("idadeMinima", 13));

        assertThat(estado.getVersao()).isEqualTo(1);
        assertThat(estado.getCriterios()).containsEntry("idadeMinima", 13);
        assertThat(estado.getHistorico()).isEmpty();
    }

    @Test
    void revisarIncrementaVersaoEPreservaHistoricoDaAnterior() {
        EstadoProtocolo estado = new EstadoProtocolo(Map.of("idadeMinima", 13));

        estado.revisar(Map.of("idadeMinima", 12), "Comitê de ética corrigiu a idade mínima.");

        assertThat(estado.getVersao()).isEqualTo(2);
        assertThat(estado.getCriterios()).containsEntry("idadeMinima", 12);
        assertThat(estado.getHistorico()).hasSize(1);
        assertThat(estado.getHistorico().get(0).versaoAnterior()).isEqualTo(1);
        assertThat(estado.getHistorico().get(0).criteriosAnteriores()).containsEntry("idadeMinima", 13);
        assertThat(estado.getHistorico().get(0).motivo()).isEqualTo("Comitê de ética corrigiu a idade mínima.");
    }
}
