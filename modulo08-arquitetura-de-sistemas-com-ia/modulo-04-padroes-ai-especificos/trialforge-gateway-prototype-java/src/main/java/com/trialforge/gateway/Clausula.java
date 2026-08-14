package com.trialforge.gateway;

/**
 * Uma clausula regulatoria de um dos indices do RAG (ICF, Protocolo ou CSR).
 */
public record Clausula(String tema, String texto, String fonte) {
}
