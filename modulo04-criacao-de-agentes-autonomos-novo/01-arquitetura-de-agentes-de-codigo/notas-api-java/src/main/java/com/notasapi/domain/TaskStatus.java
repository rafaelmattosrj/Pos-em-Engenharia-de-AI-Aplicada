package com.notasapi.domain;

public enum TaskStatus {
    OPEN("open"),
    DONE("done");

    private final String wireValue;

    TaskStatus(String wireValue) {
        this.wireValue = wireValue;
    }

    public String wireValue() {
        return wireValue;
    }

    public static TaskStatus fromWireValue(String value) {
        for (TaskStatus status : values()) {
            if (status.wireValue.equals(value)) {
                return status;
            }
        }
        throw new IllegalArgumentException("Invalid task status: " + value);
    }
}
