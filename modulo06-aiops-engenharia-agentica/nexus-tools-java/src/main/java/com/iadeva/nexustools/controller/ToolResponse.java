package com.iadeva.nexustools.controller;

/**
 * Envelope de resposta comum a todos os endpoints de ferramentas —
 * equivalente ao retorno String simples de cada função decorada com @tool
 * na versão Python.
 */
public record ToolResponse(String result) {}
