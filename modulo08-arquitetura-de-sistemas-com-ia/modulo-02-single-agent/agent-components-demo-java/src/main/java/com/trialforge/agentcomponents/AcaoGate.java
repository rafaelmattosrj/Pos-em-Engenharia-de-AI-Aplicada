package com.trialforge.agentcomponents;

import java.util.function.Supplier;

/**
 * 4 e 5. ACAO + APPROVAL GATE — nem toda acao executa direto.
 *
 * Porte 1:1 de agent-components-demo.js / agent_components_demo.py (secao 4/5).
 */
public final class AcaoGate {

    private AcaoGate() {
    }

    /**
     * Adaptacao: no original, {@code executar} e uma funcao/lambda sem argumentos
     * embutida no objeto literal da acao proposta. Aqui vira um {@link Supplier}
     * dentro de um record — mesmo formato observavel (so roda se nao houver gate).
     */
    public record AcaoProposta(String tipo, boolean requerAprovacao, Supplier<String> executar) {
    }

    public record ResultadoAcao(String status, String resultado, String mensagem) {
        static ResultadoAcao executada(String resultado) {
            return new ResultadoAcao("executada", resultado, null);
        }

        static ResultadoAcao aguardandoAprovacao(String mensagem) {
            return new ResultadoAcao("aguardando_aprovacao", null, mensagem);
        }
    }

    public static ResultadoAcao executarOuGatear(AcaoProposta acaoProposta) {
        if (acaoProposta.requerAprovacao()) {
            return ResultadoAcao.aguardandoAprovacao(
                    "Ação \"" + acaoProposta.tipo() + "\" NÃO executada. Aguardando Approval Gate "
                            + "(Módulo 1.3, Pergunta 2: erro caro e irreversível)."
            );
        }
        return ResultadoAcao.executada(acaoProposta.executar().get());
    }
}
