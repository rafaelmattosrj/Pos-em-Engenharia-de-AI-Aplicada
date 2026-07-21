package com.iadeva.guardrails.graph;

import com.iadeva.guardrails.model.GuardrailResult;
import com.iadeva.guardrails.model.User;
import com.iadeva.guardrails.service.ChatService;
import com.iadeva.guardrails.service.GuardrailsService;
import org.springframework.stereotype.Component;

import java.util.Map;

// Orquestrador do fluxo de segurança — equivalente ao StateGraph do LangGraph
// Fluxo: START → guardrails_check → (blocked | chat) → END
@Component
public class SafeguardOrchestrator {

    private final GuardrailsService guardrailsService;
    private final ChatService chatService;
    private final Map<String, User> users;

    public SafeguardOrchestrator(
            GuardrailsService guardrailsService,
            ChatService chatService,
            Map<String, User> users) {
        this.guardrailsService = guardrailsService;
        this.chatService = chatService;
        this.users = users;
    }

    public ChatResult process(String username, String userMessage) {
        User user = users.getOrDefault(username,
                new User(username, "member", java.util.List.of(), username));

        String systemPrompt = """
                You are a helpful assistant.
                User role: %s
                User name: %s
                Only users with admin role can access files.
                """.formatted(user.role(), user.displayName());

        // Nó 1: verificação de guardrails
        GuardrailResult guardrailResult = guardrailsService.check(
                userMessage, user.role(), user.displayName());

        System.out.printf("🔒 Guardrails check: safe=%s%n", guardrailResult.safe());

        // Aresta condicional: blocked vs chat
        if (!guardrailResult.safe()) {
            String blockedMessage = "⛔ Sua mensagem foi bloqueada pelo sistema de segurança. " +
                    "Motivo: " + guardrailResult.reason();
            return new ChatResult(false, blockedMessage);
        }

        // Nó chat: processa normalmente
        String response = chatService.chat(systemPrompt, userMessage);
        return new ChatResult(true, response);
    }

    public record ChatResult(boolean allowed, String message) {}
}
