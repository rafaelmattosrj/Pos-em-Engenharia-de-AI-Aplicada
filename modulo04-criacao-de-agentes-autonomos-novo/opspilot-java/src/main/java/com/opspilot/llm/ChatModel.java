package com.opspilot.llm;

import com.opspilot.tools.Tool;

import java.util.List;

/**
 * Porte da fachada OpsChatModel (agents/model.ts): abstração sobre "invocar um
 * LLM", desacoplada de qualquer provedor/biblioteca especifica (o original usa
 * ChatOpenAI/LangChain sobre OpenRouter). Uma implementação real ligaria isto
 * a um provedor OpenAI-compatible via HttpClient -- fora do escopo deste porte
 * arquitetural (ver README).
 */
public interface ChatModel {
    ModelResponse invoke(List<ChatMessage> messages, List<Tool> tools);

    default ModelResponse invoke(List<ChatMessage> messages) {
        return invoke(messages, List.of());
    }
}
