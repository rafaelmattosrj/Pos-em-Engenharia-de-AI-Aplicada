package com.iadeva.musicas.repository;

import com.iadeva.musicas.model.ConversationMessage;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

// Equivalente a TestConversationRepository_SaveAndFind (repository_test.go)
@DataJpaTest
class ConversationRepositoryTest {

    @Autowired
    private ConversationRepository repository;

    @Test
    void save_and_findBySessionId_ordersByCreatedAtAscending() throws InterruptedException {
        repository.save(new ConversationMessage("session-1", "user", "ola"));
        Thread.sleep(5); // garante createdAt distinto para ordenação determinística
        repository.save(new ConversationMessage("session-1", "assistant", "oi, tudo bem?"));
        repository.save(new ConversationMessage("session-2", "user", "mensagem de outra sessao"));

        List<ConversationMessage> messages = repository.findBySessionIdOrderByCreatedAtAsc("session-1");

        assertThat(messages).hasSize(2);
        assertThat(messages.get(0).getRole()).isEqualTo("user");
        assertThat(messages.get(1).getRole()).isEqualTo("assistant");
    }

    @Test
    void countBySessionId_countsOnlyMessagesFromThatSession() {
        repository.save(new ConversationMessage("s1", "user", "a"));
        repository.save(new ConversationMessage("s1", "assistant", "b"));
        repository.save(new ConversationMessage("s2", "user", "c"));

        assertThat(repository.countBySessionId("s1")).isEqualTo(2);
        assertThat(repository.countBySessionId("s2")).isEqualTo(1);
    }
}
