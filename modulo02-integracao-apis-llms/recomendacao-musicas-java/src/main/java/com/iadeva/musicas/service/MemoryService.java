package com.iadeva.musicas.service;

import com.iadeva.musicas.model.ConversationMessage;
import com.iadeva.musicas.repository.ConversationRepository;
import org.springframework.stereotype.Service;

import java.util.List;

// Gerencia o histórico de conversas persistido no banco
// Equivalente ao PostgresSaver (checkpointer) do LangGraph
@Service
public class MemoryService {

    private final ConversationRepository repository;

    public MemoryService(ConversationRepository repository) {
        this.repository = repository;
    }

    public void addMessage(String sessionId, String role, String content) {
        repository.save(new ConversationMessage(sessionId, role, content));
    }

    public List<ConversationMessage> getHistory(String sessionId) {
        return repository.findBySessionIdOrderByCreatedAtAsc(sessionId);
    }

    public long countMessages(String sessionId) {
        return repository.countBySessionId(sessionId);
    }
}
