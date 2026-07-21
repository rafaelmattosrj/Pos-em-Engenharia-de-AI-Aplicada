package com.iadeva.guardrails.service;

import com.iadeva.guardrails.ResilientChatClient;
import org.springframework.stereotype.Service;

// Serviço de chat principal — só recebe inputs aprovados pelo guardrails
@Service
public class ChatService {

    private final ResilientChatClient fallbackClient;

    public ChatService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public String chat(String systemPrompt, String userMessage) {
        return fallbackClient.call(systemPrompt, userMessage);
    }
}
