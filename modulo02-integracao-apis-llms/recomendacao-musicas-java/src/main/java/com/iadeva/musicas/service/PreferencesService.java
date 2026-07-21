package com.iadeva.musicas.service;

import com.iadeva.musicas.ResilientChatClient;
import com.iadeva.musicas.model.UserPreferences;
import com.iadeva.musicas.repository.PreferencesRepository;
import org.springframework.stereotype.Service;

import java.util.List;

// Gerencia preferências musicais persistidas
// Equivalente ao PostgresStore (store) do LangGraph
@Service
public class PreferencesService {

    private final PreferencesRepository repository;
    private final ResilientChatClient fallbackClient;

    public PreferencesService(PreferencesRepository repository, ResilientChatClient fallbackClient) {
        this.repository = repository;
        this.fallbackClient = fallbackClient;
    }

    public UserPreferences getOrCreate(String userId) {
        return repository.findByUserId(userId)
                .orElseGet(() -> repository.save(new UserPreferences(userId)));
    }

    // Extrai e salva preferências da conversa usando LLM
    public void extractAndSavePreferences(String userId, List<String> conversationTexts) {
        String conversation = String.join("\n", conversationTexts);
        String extracted = fallbackClient.call(
                """
                Extraia preferências musicais desta conversa em JSON.
                Inclua: gêneros, artistas, músicas, humor preferido.
                Responda APENAS com o JSON, sem explicações.
                """,
                conversation);

        UserPreferences prefs = getOrCreate(userId);
        prefs.setPreferences(extracted);
        repository.save(prefs);
    }

    // Sumariza a conversa quando fica muito longa
    public String summarize(String userId, List<String> conversationTexts) {
        String conversation = String.join("\n", conversationTexts);
        String summary = fallbackClient.call(
                "Resuma esta conversa sobre preferências musicais em 3-5 frases.",
                conversation);

        UserPreferences prefs = getOrCreate(userId);
        prefs.setConversationSummary(summary);
        repository.save(prefs);
        return summary;
    }
}
