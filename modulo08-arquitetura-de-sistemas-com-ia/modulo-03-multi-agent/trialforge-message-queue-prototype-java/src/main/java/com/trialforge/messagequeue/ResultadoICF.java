package com.trialforge.messagequeue;

import java.util.Map;

/** Resultado final do Agente ICF, ja com o contador de tentativas do registro idempotente. */
public record ResultadoICF(String agente, String secao, Map<String, Integer> criterioUsado,
                            int versaoUsada, int versaoAoConcluir, int tentativas) {
}
