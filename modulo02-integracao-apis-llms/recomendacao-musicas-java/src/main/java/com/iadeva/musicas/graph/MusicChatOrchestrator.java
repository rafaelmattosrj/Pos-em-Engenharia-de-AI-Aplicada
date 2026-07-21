package com.iadeva.musicas.graph;

import com.iadeva.musicas.ResilientChatClient;
import com.iadeva.musicas.model.ConversationMessage;
import com.iadeva.musicas.model.UserPreferences;
import com.iadeva.musicas.service.MemoryService;
import com.iadeva.musicas.service.PreferencesService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

// Orquestrador — equivalente ao StateGraph do LangGraph com checkpointer e store
// Fluxo: START → chat → (savePreferences? / summarize?) → END
@Component
public class MusicChatOrchestrator {

    private final ResilientChatClient fallbackClient;
    private final MemoryService memoryService;
    private final PreferencesService preferencesService;

    @Value("${app.summarize-after-messages:10}")
    private int summarizeAfterMessages;

    public MusicChatOrchestrator(
            ResilientChatClient fallbackClient,
            MemoryService memoryService,
            PreferencesService preferencesService) {
        this.fallbackClient = fallbackClient;
        this.memoryService = memoryService;
        this.preferencesService = preferencesService;
    }

    public String chat(String userId, String sessionId, String userMessage) {
        // Recupera contexto persistido
        UserPreferences prefs = preferencesService.getOrCreate(userId);
        List<ConversationMessage> history = memoryService.getHistory(sessionId);

        // Monta contexto da conversa
        String conversationContext = history.stream()
                .map(m -> m.getRole() + ": " + m.getContent())
                .collect(Collectors.joining("\n"));

        String userContext = prefs.getConversationSummary() != null
                ? "Resumo anterior: " + prefs.getConversationSummary() + "\n"
                : "";
        userContext += prefs.getPreferences() != null && !prefs.getPreferences().equals("{}")
                ? "Preferências: " + prefs.getPreferences()
                : "";

        // Nó chat: gera resposta com contexto completo
        String response = fallbackClient.call(
                """
                Você é um especialista em música. Faça recomendações personalizadas.
                Use o contexto do usuário para personalizar suas respostas.
                Contexto do usuário: %s
                """.formatted(userContext),
                conversationContext.isEmpty() ? userMessage :
                        "Histórico:\n" + conversationContext + "\n\nMensagem atual: " + userMessage);

        // Persiste mensagens
        memoryService.addMessage(sessionId, "user", userMessage);
        memoryService.addMessage(sessionId, "assistant", response);

        // Aresta condicional: precisa sumarizar?
        long messageCount = memoryService.countMessages(sessionId);
        if (messageCount >= summarizeAfterMessages) {
            List<String> allTexts = memoryService.getHistory(sessionId).stream()
                    .map(m -> m.getRole() + ": " + m.getContent())
                    .toList();
            preferencesService.summarize(userId, allTexts);
            preferencesService.extractAndSavePreferences(userId, allTexts);
        } else if (containsMusicPreference(userMessage)) {
            // Aresta: salvar preferências
            preferencesService.extractAndSavePreferences(userId,
                    List.of("user: " + userMessage, "assistant: " + response));
        }

        return response;
    }

    private boolean containsMusicPreference(String message) {
        String lower = message.toLowerCase();
        return lower.contains("gosto") || lower.contains("adoro") || lower.contains("prefiro")
                || lower.contains("curto") || lower.contains("favorit") || lower.contains("ouço");
    }
}
