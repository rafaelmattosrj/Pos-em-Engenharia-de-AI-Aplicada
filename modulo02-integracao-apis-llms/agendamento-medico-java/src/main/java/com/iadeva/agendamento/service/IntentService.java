package com.iadeva.agendamento.service;

import com.iadeva.agendamento.ResilientChatClient;
import com.iadeva.agendamento.model.IntentResult;
import com.iadeva.agendamento.model.Professional;
import org.springframework.ai.converter.BeanOutputConverter;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

// Identifica a intenção da mensagem do paciente usando LLM com structured output
// Equivalente ao identifyIntentNode.ts com generateStructured()
@Service
public class IntentService {

    private final ResilientChatClient fallbackClient;

    public IntentService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public IntentResult identifyIntent(String userMessage, List<Professional> professionals) {
        String profList = professionals.stream()
                .map(p -> "ID %d: %s (%s)".formatted(p.id(), p.name(), p.specialty()))
                .collect(Collectors.joining("\n"));

        String systemPrompt = """
                Você é um assistente de agendamento médico. Analise a mensagem do paciente e extraia as informações.

                Profissionais disponíveis:
                %s

                Retorne um JSON com os campos:
                - intent: "schedule", "cancel" ou "unknown"
                - patientName: nome do paciente (ou null)
                - professionalId: ID do profissional (ou null)
                - professionalName: nome do profissional (ou null)
                - datetime: data e hora em ISO 8601 (ou null)
                - reason: motivo da consulta (ou null)

                Responda APENAS com o JSON, sem explicações.
                """.formatted(profList);

        BeanOutputConverter<IntentResult> converter = new BeanOutputConverter<>(IntentResult.class);

        String response = fallbackClient.call(systemPrompt, userMessage);

        try {
            return converter.convert(response);
        } catch (Exception e) {
            return new IntentResult("unknown", null, null, null, null, null);
        }
    }
}
