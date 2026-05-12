package com.iadeva.memory.model;

import java.time.Instant;
import java.util.List;

/**
 * Representa um episódio de execução do agente — uma interação completa com input, passos e resultado.
 * Armazenado na memória episódica para consulta futura.
 */
public record Episode(
        String id,
        String input,
        List<String> steps,
        String outcome,
        List<String> lessonsExtracted,
        Instant timestamp
) {}
