package com.notasapi;

import com.notasapi.domain.TaskListFilter;
import com.notasapi.store.JsonFileTaskStore;
import com.notasapi.store.TaskStorePersistenceException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class JsonFileTaskStoreTest {

    @TempDir
    Path tempDir;

    @Test
    void persistsAcrossInstances(@TempDir Path dir) {
        Path storePath = dir.resolve("store.json");
        AtomicInteger counter = new AtomicInteger();
        JsonFileTaskStore first = new JsonFileTaskStore(storePath, () -> "id-" + counter.incrementAndGet());
        first.create("Persistente");

        JsonFileTaskStore second = new JsonFileTaskStore(storePath);
        assertThat(second.list(TaskListFilter.ALL)).hasSize(1);
        assertThat(second.list(TaskListFilter.ALL).get(0).title()).isEqualTo("Persistente");
    }

    @Test
    void missingFileStartsEmpty() {
        Path storePath = tempDir.resolve("nao-existe.json");
        JsonFileTaskStore store = new JsonFileTaskStore(storePath);

        assertThat(store.list(TaskListFilter.ALL)).isEmpty();
    }

    @Test
    void invalidJsonThrowsPersistenceException() throws Exception {
        Path storePath = tempDir.resolve("invalido.json");
        Files.writeString(storePath, "{ not valid json");

        assertThatThrownBy(() -> new JsonFileTaskStore(storePath))
                .isInstanceOf(TaskStorePersistenceException.class);
    }

    @Test
    void missingTasksFieldThrowsPersistenceException() throws Exception {
        Path storePath = tempDir.resolve("sem-tasks.json");
        Files.writeString(storePath, "{}");

        assertThatThrownBy(() -> new JsonFileTaskStore(storePath))
                .isInstanceOf(TaskStorePersistenceException.class);
    }
}
