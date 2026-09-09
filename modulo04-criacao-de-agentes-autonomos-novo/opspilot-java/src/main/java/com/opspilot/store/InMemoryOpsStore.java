package com.opspilot.store;

import com.opspilot.domain.Alert;
import com.opspilot.domain.AlertStatus;
import com.opspilot.domain.Exceptions;
import com.opspilot.domain.Incident;
import com.opspilot.domain.IncidentStatus;
import com.opspilot.domain.Runbook;
import com.opspilot.domain.Severity;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Porte de InMemoryStore (store/in-memory-store.ts) -- o próprio original usa
 * essa implementação em memória em testes; aqui ela é a implementação
 * default (o original usa SQLite em produção via SqliteOpsStore, fora do
 * escopo deste porte -- ver README).
 */
public class InMemoryOpsStore implements OpsStore {

    private final List<Alert> alerts = new ArrayList<>();
    private final List<Incident> incidents = new ArrayList<>();
    private final Map<String, Runbook> runbooks = new LinkedHashMap<>();

    @Override
    public List<Alert> getAlerts(AlertStatus status) {
        return alerts.stream()
                .filter(alert -> status == null || alert.status() == status)
                .collect(Collectors.toList());
    }

    @Override
    public List<Incident> getIncidents(IncidentStatus status) {
        return incidents.stream()
                .filter(incident -> status == null || incident.status() == status)
                .sorted((a, b) -> Long.compare(a.createdAt(), b.createdAt()))
                .collect(Collectors.toList());
    }

    @Override
    public Incident createIncident(String title, String service, Severity severity) {
        Incident incident = new Incident(
                "inc-" + System.currentTimeMillis() + "-" + UUID.randomUUID().toString().substring(0, 4),
                title, service, severity, IncidentStatus.OPEN, System.currentTimeMillis(), null, null);
        incidents.add(incident);
        return incident;
    }

    @Override
    public Incident resolveIncident(String id, String summary) {
        for (int i = 0; i < incidents.size(); i++) {
            if (incidents.get(i).id().equals(id)) {
                Incident resolved = incidents.get(i).resolved(System.currentTimeMillis(), summary);
                incidents.set(i, resolved);
                return resolved;
            }
        }
        throw new Exceptions.IncidentNotFoundException(id);
    }

    @Override
    public Runbook getRunbook(String service) {
        Runbook runbook = runbooks.get(service);
        if (runbook == null) {
            throw new Exceptions.RunbookNotFoundException(service);
        }
        return runbook;
    }

    @Override
    public void seedAlert(Alert alert) {
        alerts.add(alert);
    }

    @Override
    public void seedRunbook(Runbook runbook) {
        runbooks.put(runbook.service(), runbook);
    }
}
