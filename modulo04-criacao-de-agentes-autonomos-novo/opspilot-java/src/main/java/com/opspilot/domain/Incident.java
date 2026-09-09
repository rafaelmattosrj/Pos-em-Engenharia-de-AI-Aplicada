package com.opspilot.domain;

public record Incident(
        String id,
        String title,
        String service,
        Severity severity,
        IncidentStatus status,
        long createdAt,
        Long resolvedAt,
        String summary) {

    public Incident resolved(long resolvedAt, String summary) {
        return new Incident(id, title, service, severity, IncidentStatus.RESOLVED, createdAt, resolvedAt, summary);
    }
}
