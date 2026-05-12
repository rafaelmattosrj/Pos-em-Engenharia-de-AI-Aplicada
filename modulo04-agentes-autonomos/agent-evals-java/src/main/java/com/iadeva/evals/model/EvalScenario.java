package com.iadeva.evals.model;

import java.util.List;

/**
 * Representa um cenário de avaliação do dataset.
 * Equivalente a uma linha do eval_dataset.py Python.
 */
public record EvalScenario(
        String id,
        String input,
        String difficulty,
        List<String> expectedTools
) {}
