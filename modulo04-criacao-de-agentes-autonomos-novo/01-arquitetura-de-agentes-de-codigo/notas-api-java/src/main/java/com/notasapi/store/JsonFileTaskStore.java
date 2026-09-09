package com.notasapi.store;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.domain.TaskStatus;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Collectors;

/**
 * Porte 1:1 de JsonFileTaskStore (store/json-file-task-store.ts): persiste em
 * {tasks: [...]}, escrevendo num arquivo temporario e fazendo rename atomico
 * (mesma estrategia do original, evita arquivo corrompido em escrita parcial).
 */
public class JsonFileTaskStore implements TaskStore {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final Path filePath;
    private final Supplier<String> generateId;
    private final Map<String, Task> tasks;

    public JsonFileTaskStore(Path filePath) {
        this(filePath, () -> UUID.randomUUID().toString());
    }

    public JsonFileTaskStore(Path filePath, Supplier<String> generateId) {
        this.filePath = filePath;
        this.generateId = generateId;
        this.tasks = loadTasksFromFile(filePath);
    }

    @Override
    public Task create(String title) {
        Task task = new Task(generateId.get(), title, TaskStatus.OPEN);
        tasks.put(task.id(), task);
        persistTasks();
        return task;
    }

    @Override
    public List<Task> list(TaskListFilter filter) {
        return tasks.values().stream()
                .filter(task -> switch (filter) {
                    case ALL -> true;
                    case OPEN -> task.status() == TaskStatus.OPEN;
                    case DONE -> task.status() == TaskStatus.DONE;
                })
                .collect(Collectors.toList());
    }

    @Override
    public Optional<Task> getById(String id) {
        return Optional.ofNullable(tasks.get(id));
    }

    @Override
    public Optional<Task> complete(String id) {
        Task task = tasks.get(id);
        if (task == null) {
            return Optional.empty();
        }
        Task completed = task.status() == TaskStatus.DONE ? task : task.withStatus(TaskStatus.DONE);
        tasks.put(id, completed);
        persistTasks();
        return Optional.of(completed);
    }

    @Override
    public boolean remove(String id) {
        Task removed = tasks.remove(id);
        if (removed == null) {
            return false;
        }
        persistTasks();
        return true;
    }

    private static Map<String, Task> loadTasksFromFile(Path filePath) {
        if (!Files.exists(filePath)) {
            return new LinkedHashMap<>();
        }

        String rawContent;
        try {
            rawContent = Files.readString(filePath, StandardCharsets.UTF_8);
        } catch (IOException e) {
            throw new TaskStorePersistenceException(
                    "Failed to read task store file at \"" + filePath + "\"", e);
        }

        if (rawContent.isBlank()) {
            return new LinkedHashMap<>();
        }

        ObjectNode root;
        try {
            root = (ObjectNode) MAPPER.readTree(rawContent);
        } catch (IOException e) {
            throw new TaskStorePersistenceException(
                    "Task store file at \"" + filePath + "\" contains invalid JSON", e);
        }

        ArrayNode tasksNode = (ArrayNode) root.get("tasks");
        if (tasksNode == null) {
            throw new TaskStorePersistenceException(
                    "Task store file at \"" + filePath + "\" has an invalid structure", null);
        }

        Map<String, Task> result = new LinkedHashMap<>();
        for (var node : tasksNode) {
            try {
                Task task = new Task(
                        node.get("id").asText(),
                        node.get("title").asText(),
                        TaskStatus.fromWireValue(node.get("status").asText()));
                result.put(task.id(), task);
            } catch (RuntimeException e) {
                throw new TaskStorePersistenceException(
                        "Task store file at \"" + filePath + "\" has an invalid structure", e);
            }
        }
        return result;
    }

    private void persistTasks() {
        Path directory = filePath.toAbsolutePath().getParent();
        Path tempFilePath = filePath.resolveSibling(
                filePath.getFileName() + "." + ProcessHandle.current().pid() + "." + System.nanoTime() + ".tmp");

        ArrayNode tasksNode = MAPPER.createArrayNode();
        for (Task task : tasks.values()) {
            ObjectNode node = MAPPER.createObjectNode();
            node.put("id", task.id());
            node.put("title", task.title());
            node.put("status", task.status().wireValue());
            tasksNode.add(node);
        }
        ObjectNode root = MAPPER.createObjectNode();
        root.set("tasks", tasksNode);

        try {
            if (directory != null) {
                Files.createDirectories(directory);
            }
            String fileContent = MAPPER.writerWithDefaultPrettyPrinter().writeValueAsString(root);
            Files.writeString(tempFilePath, fileContent + "\n", StandardCharsets.UTF_8);
            Files.move(tempFilePath, filePath, StandardCopyOption.REPLACE_EXISTING, StandardCopyOption.ATOMIC_MOVE);
        } catch (IOException e) {
            throw new TaskStorePersistenceException(
                    "Failed to persist task store file at \"" + filePath + "\"", e);
        }
    }
}
