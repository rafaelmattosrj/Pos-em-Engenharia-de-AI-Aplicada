package com.iadeva.musicas.model;

import jakarta.persistence.*;
import java.time.Instant;

// Mensagem persistida — equivalente ao checkpoint do LangGraph no PostgreSQL
@Entity
@Table(name = "conversation_messages")
public class ConversationMessage {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String sessionId;

    @Column(nullable = false)
    private String role; // "user" | "assistant"

    @Column(columnDefinition = "TEXT", nullable = false)
    private String content;

    @Column(nullable = false)
    private Instant createdAt = Instant.now();

    public ConversationMessage() {}

    public ConversationMessage(String sessionId, String role, String content) {
        this.sessionId = sessionId;
        this.role = role;
        this.content = content;
    }

    public Long getId() { return id; }
    public String getSessionId() { return sessionId; }
    public String getRole() { return role; }
    public String getContent() { return content; }
    public Instant getCreatedAt() { return createdAt; }
}
