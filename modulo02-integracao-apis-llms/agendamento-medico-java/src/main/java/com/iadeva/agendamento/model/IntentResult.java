package com.iadeva.agendamento.model;

// Output estruturado do LLM para identificação de intenção
// Equivalente ao IntentSchema do Zod no TypeScript
public record IntentResult(
        String intent,           // "schedule" | "cancel" | "unknown"
        String patientName,
        Integer professionalId,
        String professionalName,
        String datetime,         // ISO 8601
        String reason
) {}
