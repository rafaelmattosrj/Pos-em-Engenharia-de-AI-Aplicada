package com.opspilot.store;

import com.opspilot.domain.Alert;
import com.opspilot.domain.AlertStatus;
import com.opspilot.domain.Incident;
import com.opspilot.domain.IncidentStatus;
import com.opspilot.domain.Runbook;
import com.opspilot.domain.Severity;

import java.util.List;

/** Porte da interface OpsStore (domain/types.ts). */
public interface OpsStore {

    List<Alert> getAlerts(AlertStatus status);

    List<Incident> getIncidents(IncidentStatus status);

    Incident createIncident(String title, String service, Severity severity);

    Incident resolveIncident(String id, String summary);

    Runbook getRunbook(String service);

    void seedAlert(Alert alert);

    void seedRunbook(Runbook runbook);
}
