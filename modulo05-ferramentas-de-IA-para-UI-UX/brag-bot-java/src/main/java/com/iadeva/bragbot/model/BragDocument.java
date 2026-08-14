package com.iadeva.bragbot.model;

import java.util.List;

/**
 * Equivalente ao BragSchema (Zod) de flows.ts, com o campo "id" gerado pelo
 * servidor após a resposta estruturada do modelo (mesma ordem de
 * `{...output, id: uuidv4()}` no flow original).
 */
public record BragDocument(
        String id,
        String title,
        String context,
        String actionTaken,
        String businessImpact,
        List<String> metrics,
        List<String> technologiesUsed
) {}
