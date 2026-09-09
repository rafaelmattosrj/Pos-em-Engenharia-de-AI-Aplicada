package com.opspilot.domain;

public record Alert(String id, String service, String description, Severity severity, AlertStatus status) {
}
