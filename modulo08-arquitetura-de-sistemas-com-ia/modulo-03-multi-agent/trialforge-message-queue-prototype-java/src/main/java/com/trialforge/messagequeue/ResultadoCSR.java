package com.trialforge.messagequeue;

import java.util.Map;

/** Resultado do Agente CSR. Sem registro idempotente: o CSR nao grava documento. */
public record ResultadoCSR(String agente, String sintese, Map<String, Integer> criterioUsado,
                            int versaoUsada, int versaoAoConcluir) {
}
