package com.iadeva.musicas.service;

import com.iadeva.musicas.model.ConversationMessage;
import com.iadeva.musicas.repository.ConversationRepository;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

// Equivalente a TestMemoryService_AddAndRetrieveHistory (service_test.go)
@DataJpaTest
class MemoryServiceTest {

    @Autowired
    private ConversationRepository repository;

    @Test
    void addMessage_thenGetHistoryAndCountMessages() {
        MemoryService service = new MemoryService(repository);

        service.addMessage("s1", "user", "quero recomendacoes de rock");
        service.addMessage("s1", "assistant", "que tal Led Zeppelin?");

        List<ConversationMessage> history = service.getHistory("s1");
        assertThat(history).hasSize(2);

        assertThat(service.countMessages("s1")).isEqualTo(2);
    }
}
