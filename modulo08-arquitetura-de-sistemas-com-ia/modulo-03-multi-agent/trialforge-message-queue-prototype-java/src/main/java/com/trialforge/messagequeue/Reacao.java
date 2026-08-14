package com.trialforge.messagequeue;

/**
 * Resultado da reacao Parallel (ICF + CSR) ao evento "protocolo:pronto".
 * Mutavel de proposito: o Supervisor reatribui campos apos o retry e apos a
 * compensacao Saga, igual aos spreads {@code {...reacao, ...}} do JS/Python.
 */
final class Reacao {

    boolean ok;
    ResultadoICF resultadoICF;
    ResultadoCSR resultadoCSR;
    String erroICF;
    boolean resultadoCSRPreservado;

    Reacao(boolean ok, ResultadoICF resultadoICF, ResultadoCSR resultadoCSR, String erroICF, boolean resultadoCSRPreservado) {
        this.ok = ok;
        this.resultadoICF = resultadoICF;
        this.resultadoCSR = resultadoCSR;
        this.erroICF = erroICF;
        this.resultadoCSRPreservado = resultadoCSRPreservado;
    }
}
