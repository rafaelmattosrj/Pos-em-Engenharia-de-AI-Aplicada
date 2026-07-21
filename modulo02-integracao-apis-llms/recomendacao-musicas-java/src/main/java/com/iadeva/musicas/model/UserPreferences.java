package com.iadeva.musicas.model;

import jakarta.persistence.*;

// Preferências musicais do usuário — equivalente ao PostgresStore do LangGraph
@Entity
@Table(name = "user_preferences")
public class UserPreferences {

    @Id
    private String userId;

    @Column(columnDefinition = "TEXT")
    private String preferences; // JSON string com preferências extraídas

    @Column(columnDefinition = "TEXT")
    private String conversationSummary;

    public UserPreferences() {}

    public UserPreferences(String userId) {
        this.userId = userId;
        this.preferences = "{}";
    }

    public String getUserId() { return userId; }
    public String getPreferences() { return preferences; }
    public void setPreferences(String preferences) { this.preferences = preferences; }
    public String getConversationSummary() { return conversationSummary; }
    public void setConversationSummary(String summary) { this.conversationSummary = summary; }
}
