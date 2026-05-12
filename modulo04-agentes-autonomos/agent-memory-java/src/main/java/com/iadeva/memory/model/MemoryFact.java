package com.iadeva.memory.model;

import java.time.Instant;

/**
 * Representa um fato persistido na memória de longo prazo.
 * Equivalente a um registro de conhecimento factual do agente.
 */
public record MemoryFact(
        String id,
        String content,
        String source,
        Instant confirmedAt,
        Instant expiresAt
) {}
