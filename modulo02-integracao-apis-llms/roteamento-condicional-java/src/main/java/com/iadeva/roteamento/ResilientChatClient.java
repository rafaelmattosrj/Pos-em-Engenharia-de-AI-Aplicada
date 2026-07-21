package com.iadeva.roteamento;

import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.model.ChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class ResilientChatClient {

    private final ChatModel chatModel;
    private final List<String> fallbackModels;

    public ResilientChatClient(
            ChatModel chatModel,
            @Value("${app.fallback-models}") String fallbackModelsConfig) {
        this.chatModel = chatModel;
        this.fallbackModels = List.of(fallbackModelsConfig.split(","));
    }

    public String call(String systemPrompt, String userPrompt) {
        Exception lastException = null;
        for (String model : fallbackModels) {
            try {
                return buildClient(model).prompt()
                        .system(systemPrompt)
                        .user(userPrompt)
                        .call()
                        .content();
            } catch (Exception e) {
                lastException = e;
            }
        }
        throw new RuntimeException("All fallback models failed", lastException);
    }

    private ChatClient buildClient(String model) {
        return ChatClient.builder(chatModel)
                .defaultOptions(OpenAiChatOptions.builder().model(model).build())
                .build();
    }
}
