package com.notasapi.service;

import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.store.TaskStore;

import java.util.List;

public class TaskServiceImpl implements TaskService {

    private final TaskStore store;

    public TaskServiceImpl(TaskStore store) {
        this.store = store;
    }

    @Override
    public Task createTask(String title) {
        String parsedTitle = requireNonBlank(title, "Task title is required");
        return store.create(parsedTitle);
    }

    @Override
    public List<Task> listTasks(TaskListFilter filter) {
        return store.list(filter == null ? TaskListFilter.ALL : filter);
    }

    @Override
    public Task completeTask(String id) {
        String parsedId = requireNonBlank(id, "Task id is required");
        return store.complete(parsedId).orElseThrow(() -> new TaskNotFoundException(parsedId));
    }

    @Override
    public void removeTask(String id) {
        String parsedId = requireNonBlank(id, "Task id is required");
        if (!store.remove(parsedId)) {
            throw new TaskNotFoundException(parsedId);
        }
    }

    private static String requireNonBlank(String value, String message) {
        String trimmed = value == null ? "" : value.trim();
        if (trimmed.isEmpty()) {
            throw new TaskValidationException(message);
        }
        return trimmed;
    }
}
