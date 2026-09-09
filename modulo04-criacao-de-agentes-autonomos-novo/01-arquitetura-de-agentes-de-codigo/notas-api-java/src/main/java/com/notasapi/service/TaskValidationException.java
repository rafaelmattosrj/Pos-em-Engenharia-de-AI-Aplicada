package com.notasapi.service;

import java.util.List;

public class TaskValidationException extends RuntimeException {

    private final List<String> issues;

    public TaskValidationException(List<String> issues) {
        super(String.join("; ", issues));
        this.issues = List.copyOf(issues);
    }

    public TaskValidationException(String issue) {
        this(List.of(issue));
    }

    public List<String> issues() {
        return issues;
    }
}
