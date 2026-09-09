package com.opspilot.domain;

public final class Exceptions {
    private Exceptions() {
    }

    public static final class IncidentNotFoundException extends RuntimeException {
        public IncidentNotFoundException(String id) {
            super("Incident \"" + id + "\" was not found");
        }
    }

    public static final class RunbookNotFoundException extends RuntimeException {
        public RunbookNotFoundException(String service) {
            super("Runbook for service \"" + service + "\" was not found");
        }
    }

    public static final class ModelUnavailableException extends RuntimeException {
        public ModelUnavailableException(String message) {
            super(message);
        }
    }
}
