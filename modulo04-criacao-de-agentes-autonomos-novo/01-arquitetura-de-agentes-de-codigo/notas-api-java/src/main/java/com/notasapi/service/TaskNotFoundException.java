package com.notasapi.service;

public class TaskNotFoundException extends RuntimeException {

    private final String taskId;

    public TaskNotFoundException(String taskId) {
        super("Task \"" + taskId + "\" was not found");
        this.taskId = taskId;
    }

    public String taskId() {
        return taskId;
    }
}
