package com.notasapi.domain;

/** Porte 1:1 do schema `taskSchema` (domain/task.ts) — id, title e status validados. */
public record Task(String id, String title, TaskStatus status) {

    public Task {
        if (id == null || id.isBlank()) {
            throw new IllegalArgumentException("Task id is required");
        }
        if (title == null || title.isBlank()) {
            throw new IllegalArgumentException("Task title is required");
        }
    }

    public Task withStatus(TaskStatus newStatus) {
        return new Task(id, title, newStatus);
    }
}
