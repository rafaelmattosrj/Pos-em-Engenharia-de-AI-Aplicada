package com.trialforge.messagequeue;

import java.util.List;

/**
 * Porte de verificarConsistencia (JS/Python) — Modulo 3.2, paragrafo 88:
 * "o Supervisor verifica se os critérios citados em cada um batem entre si."
 */
public record Verificacao(boolean consistente, boolean icfConsistente, boolean csrConsistente,
                           List<String> agentesDivergentes) {
}
