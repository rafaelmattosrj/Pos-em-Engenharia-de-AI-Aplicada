package com.trialforge.gateway;

import java.util.List;

/** Simula o Approval Gate com uma fila de respostas pré-definida, sem
 * depender de stdin. */
class FakeApprover implements Approver {

    private final List<Boolean> respostas;
    int chamadas = 0;

    FakeApprover(Boolean... respostas) {
        this.respostas = List.of(respostas);
    }

    @Override
    public boolean pedirAprovacaoHumana(String rascunho) {
        boolean r = chamadas < respostas.size() && respostas.get(chamadas);
        chamadas++;
        return r;
    }
}
