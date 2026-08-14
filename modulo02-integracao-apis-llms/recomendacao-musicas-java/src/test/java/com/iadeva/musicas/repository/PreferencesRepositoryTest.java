package com.iadeva.musicas.repository;

import com.iadeva.musicas.model.UserPreferences;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;

// Equivalente a TestPreferencesRepository_SaveAndFind (repository_test.go)
@DataJpaTest
class PreferencesRepositoryTest {

    @Autowired
    private PreferencesRepository repository;

    @Test
    void findByUserId_unknownUser_returnsEmpty() {
        assertThat(repository.findByUserId("usuario-inexistente")).isEmpty();
    }

    @Test
    void save_thenFindByUserId_returnsPersistedPreferences() {
        UserPreferences prefs = new UserPreferences("user-1");
        prefs.setPreferences("{\"genero\":\"rock\"}");

        repository.save(prefs);

        Optional<UserPreferences> found = repository.findByUserId("user-1");
        assertThat(found).isPresent();
        assertThat(found.get().getPreferences()).isEqualTo("{\"genero\":\"rock\"}");
    }

    @Test
    void save_isUpsert_updatesExistingRecordInsteadOfDuplicating() {
        UserPreferences prefs = new UserPreferences("user-1");
        repository.save(prefs);

        UserPreferences found = repository.findByUserId("user-1").orElseThrow();
        found.setConversationSummary("resumo da conversa");
        repository.save(found);

        UserPreferences updated = repository.findByUserId("user-1").orElseThrow();
        assertThat(updated.getConversationSummary()).isEqualTo("resumo da conversa");
        assertThat(repository.count()).isEqualTo(1);
    }
}
