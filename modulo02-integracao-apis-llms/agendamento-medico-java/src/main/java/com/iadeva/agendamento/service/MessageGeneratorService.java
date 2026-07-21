package com.iadeva.agendamento.service;

import com.iadeva.agendamento.ResilientChatClient;
import org.springframework.stereotype.Service;

// Gera mensagens amigáveis para o paciente — equivalente ao messageGeneratorNode.ts
@Service
public class MessageGeneratorService {

    private final ResilientChatClient fallbackClient;

    public MessageGeneratorService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public String generateSuccessMessage(String action, String patientName, String details) {
        return fallbackClient.call(
                "Você é um assistente de clínica médica. Gere uma mensagem amigável e profissional em português.",
                "Gere uma confirmação de %s para o paciente %s. Detalhes: %s".formatted(action, patientName, details));
    }

    public String generateErrorMessage(String error) {
        return fallbackClient.call(
                "Você é um assistente de clínica médica. Gere uma mensagem amigável em português.",
                "Informe ao paciente que houve um problema: %s".formatted(error));
    }

    public String generateUnknownIntentMessage(String originalMessage) {
        return fallbackClient.call(
                "Você é um assistente de clínica médica. Responda de forma amigável em português.",
                originalMessage);
    }
}
