package com.iadeva.musicas.repository;

import com.iadeva.musicas.model.ConversationMessage;
import org.springframework.data.jpa.repository.JpaRepository;
import java.util.List;

public interface ConversationRepository extends JpaRepository<ConversationMessage, Long> {
    List<ConversationMessage> findBySessionIdOrderByCreatedAtAsc(String sessionId);
    long countBySessionId(String sessionId);
}
