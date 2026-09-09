package com.notasapi;

import com.notasapi.cli.TaskCli;
import com.notasapi.domain.Task;
import com.notasapi.service.TaskService;
import com.notasapi.service.TaskServiceImpl;
import com.notasapi.store.InMemoryTaskStore;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

import static org.assertj.core.api.Assertions.assertThat;

class TaskCliTest {

    private TaskService taskService;
    private ByteArrayOutputStream out;
    private ByteArrayOutputStream err;
    private PrintStream stdout;
    private PrintStream stderr;

    @BeforeEach
    void setUp() {
        taskService = new TaskServiceImpl(new InMemoryTaskStore());
        out = new ByteArrayOutputStream();
        err = new ByteArrayOutputStream();
        stdout = new PrintStream(out, true, StandardCharsets.UTF_8);
        stderr = new PrintStream(err, true, StandardCharsets.UTF_8);
    }

    @Test
    void createPrintsCreatedTask() {
        int exitCode = TaskCli.run(new String[]{"create", "--title", "Estudar"}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(0);
        assertThat(out.toString(StandardCharsets.UTF_8)).contains("Created task:").contains("Estudar");
    }

    @Test
    void listWithoutTasksPrintsMessage() {
        int exitCode = TaskCli.run(new String[]{"list"}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(0);
        assertThat(out.toString(StandardCharsets.UTF_8)).contains("No tasks found.");
    }

    @Test
    void missingCommandIsUsageError() {
        int exitCode = TaskCli.run(new String[]{}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(1);
        assertThat(err.toString(StandardCharsets.UTF_8)).contains("Missing command").contains("Usage:");
    }

    @Test
    void completeUnknownIdIsNotFoundError() {
        int exitCode = TaskCli.run(new String[]{"complete", "--id", "ghost"}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(1);
        assertThat(err.toString(StandardCharsets.UTF_8)).contains("was not found");
    }

    @Test
    void taskPrefixIsStripped() {
        Task task = taskService.createTask("Prefixo");

        int exitCode = TaskCli.run(
                new String[]{"task", "complete", "--id", task.id()}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(0);
        assertThat(out.toString(StandardCharsets.UTF_8)).contains("Completed task:");
    }

    @Test
    void unknownCommandIsUsageError() {
        int exitCode = TaskCli.run(new String[]{"unknown"}, taskService, stdout, stderr);

        assertThat(exitCode).isEqualTo(1);
        assertThat(err.toString(StandardCharsets.UTF_8)).contains("Unknown command");
    }
}
