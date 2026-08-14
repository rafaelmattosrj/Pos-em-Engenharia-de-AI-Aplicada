package com.trialforge.gateway;

/**
 * Contrato do Approval Gate (Modulo 4.4) — implementado por ApprovalGate em
 * producao; testes de Processor usam um stub que nao depende de stdin.
 */
@FunctionalInterface
public interface Approver {
    boolean pedirAprovacaoHumana(String rascunho) throws Exception;
}
