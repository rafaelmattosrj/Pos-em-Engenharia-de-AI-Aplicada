package com.iadeva.memory.model;

import java.time.Instant;

/**
 * Representa uma lição aprendida extraída pela ReflectionEngine após uma execução.
 * Lições generalizáveis são reutilizadas em execuções futuras similares.
 */
public record Lesson(
        String id,
        String situation,
        String action,
        String result,
        String learning,
        String generalizability,
        Instant extractedAt
) {}
