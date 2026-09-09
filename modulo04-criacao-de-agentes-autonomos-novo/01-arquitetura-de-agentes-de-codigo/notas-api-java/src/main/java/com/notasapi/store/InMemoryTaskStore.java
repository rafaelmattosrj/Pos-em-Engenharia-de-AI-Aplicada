package com.notasapi.store;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.domain.TaskStatus;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Collectors;

/** Porte 1:1 de InMemoryTaskStore (store/in-memory-task-store.ts). */
public class InMemoryTaskStore implements TaskStore {

    private final Map<String, Task> tasks = new LinkedHashMap<>();
    private final Supplier<String> generateId;

    public InMemoryTaskStore() {
        this(() -> UUID.randomUUID().toString());
    }

    public InMemoryTaskStore(Supplier<String> generateId) {
        this.generateId = generateId;
    }

    @Override
    public Task create(String title) {
        Task task = new Task(generateId.get(), title, TaskStatus.OPEN);
        tasks.put(task.id(), task);
        return task;
    }

    @Override
    public List<Task> list(TaskListFilter filter) {
        return tasks.values().stream()
                .filter(task -> matchesFilter(task, filter))
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
        return Optional.of(completed);
    }

    @Override
    public boolean remove(String id) {
        return tasks.remove(id) != null;
    }

    private static boolean matchesFilter(Task task, TaskListFilter filter) {
        return switch (filter) {
            case ALL -> true;
            case OPEN -> task.status() == TaskStatus.OPEN;
            case DONE -> task.status() == TaskStatus.DONE;
        };
    }
}
