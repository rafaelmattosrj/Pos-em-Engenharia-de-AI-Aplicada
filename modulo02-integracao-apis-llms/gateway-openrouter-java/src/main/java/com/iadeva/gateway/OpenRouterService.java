package com.iadeva.gateway;

import org.springframework.ai.chat.model.ChatResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class OpenRouterService {

    private final ResilientChatClient fallbackClient;

    @Value("${app.system-prompt}")
    private String systemPrompt;

    public OpenRouterService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public LlmResponse generate(String prompt) {
        ChatResponse response = fallbackClient.callForResponse(systemPrompt, prompt);

        String content = response.getResult().getOutput().getText();
        String model = response.getMetadata() != null
                ? response.getMetadata().getModel()
                : "unknown";

        return new LlmResponse(model, content);
    }
}
