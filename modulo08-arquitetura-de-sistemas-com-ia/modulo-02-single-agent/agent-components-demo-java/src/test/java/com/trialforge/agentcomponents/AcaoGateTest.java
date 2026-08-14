package com.trialforge.agentcomponents;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre o mesmo cenario de testarGate() em agent-components-demo.js / .py:
 * acao sem gate executa direto; acao com gate nunca chega a executar.
 */
class AcaoGateTest {

    @Test
    void acaoSemGate_executaDireto() {
        AcaoGate.ResultadoAcao resultado =
                AcaoGate.executarOuGatear(new AcaoGate.AcaoProposta("x", false, () -> "feito"));

        assertThat(resultado.status()).isEqualTo("executada");
        assertThat(resultado.resultado()).isEqualTo("feito");
        assertThat(resultado.mensagem()).isNull();
    }

    @Test
    void acaoComGate_nuncaChegaAExecutar() {
        boolean[] executou = {false};
        AcaoGate.ResultadoAcao resultado = AcaoGate.executarOuGatear(new AcaoGate.AcaoProposta("y", true, () -> {
            executou[0] = true;
            return "nunca deveria rodar";
        }));

        assertThat(resultado.status()).isEqualTo("aguardando_aprovacao");
        assertThat(resultado.resultado()).isNull();
        assertThat(resultado.mensagem()).contains("y").contains("Approval Gate");
        assertThat(executou[0]).isFalse();
    }
}
