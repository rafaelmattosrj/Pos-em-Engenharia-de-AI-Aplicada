package com.iadeva.evals.model;

import java.util.List;
import java.util.Map;

/**
 * Caso de teste para avaliação de seleção de ferramentas.
 * Define o contexto, a ferramenta esperada e as ferramentas proibidas.
 */
public record ToolSelectionCase(
        String id,
        String context,
        int loopStep,
        String expectedTool,
        Map<String, Object> expectedArgs,
        List<String> forbiddenTools
) {}
