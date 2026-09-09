package com.notasapi.domain;

/** Porte de `taskListFilterSchema` — "all" nao filtra, os demais casam com TaskStatus. */
public enum TaskListFilter {
    ALL("all"),
    OPEN("open"),
    DONE("done");

    private final String wireValue;

    TaskListFilter(String wireValue) {
        this.wireValue = wireValue;
    }

    public String wireValue() {
        return wireValue;
    }

    public static TaskListFilter fromWireValue(String value) {
        for (TaskListFilter filter : values()) {
            if (filter.wireValue.equals(value)) {
                return filter;
            }
        }
        throw new IllegalArgumentException("Invalid task list filter: " + value);
    }
}
