package com.iadeva.guardrails.model;

public record GuardrailResult(boolean safe, String reason, String analysis) {}
