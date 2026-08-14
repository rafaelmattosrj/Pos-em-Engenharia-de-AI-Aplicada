package com.iadeva.cfp.model;

/**
 * Equivalente à interface EventDTO de shared-types/src/lib/event.dto.ts.
 */
public record Event(String id, String nome, String endereco, int capacidade, String data) {}
