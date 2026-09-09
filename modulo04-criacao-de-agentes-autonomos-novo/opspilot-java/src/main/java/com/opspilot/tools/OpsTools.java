package com.opspilot.tools;

import com.opspilot.domain.AlertStatus;
import com.opspilot.domain.IncidentStatus;
import com.opspilot.domain.Severity;
import com.opspilot.store.OpsStore;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Porte de agents/tools.ts -- ferramentas concretas usadas pelos papeis
 * analista (leitura) e executor (mutacao de incidentes), conforme a
 * particao estrutural documentada em team-strategy.ts (spec 018 FR-007).
 */
public final class OpsTools {

    private OpsTools() {
    }

    public static Tool listAlerts(OpsStore store) {
        return new Tool() {
            public String name() {
                return "list_alerts";
            }

            public String description() {
                return "Lista os alertas disparando ou resolvidos.";
            }

            public String execute(Map<String, Object> args) {
                AlertStatus status = args.containsKey("status")
                        ? AlertStatus.valueOf(String.valueOf(args.get("status")).toUpperCase())
                        : null;
                return store.getAlerts(status).stream()
                        .map(alert -> "[%s] %s (%s): %s".formatted(alert.severity(), alert.service(), alert.status(), alert.description()))
                        .collect(Collectors.joining("\n"));
            }
        };
    }

    public static Tool listIncidents(OpsStore store) {
        return new Tool() {
            public String name() {
                return "list_incidents";
            }

            public String description() {
                return "Lista incidentes abertos ou resolvidos.";
            }

            public String execute(Map<String, Object> args) {
                IncidentStatus status = args.containsKey("status")
                        ? IncidentStatus.valueOf(String.valueOf(args.get("status")).toUpperCase())
                        : null;
                List<String> lines = store.getIncidents(status).stream()
                        .map(incident -> "%s [%s/%s] %s (%s)".formatted(
                                incident.id(), incident.severity(), incident.status(), incident.title(), incident.service()))
                        .collect(Collectors.toList());
                return lines.isEmpty() ? "(nenhum incidente)" : String.join("\n", lines);
            }
        };
    }

    public static Tool openIncident(OpsStore store) {
        return new Tool() {
            public String name() {
                return "open_incident";
            }

            public String description() {
                return "Abre um novo incidente (title, service, severity).";
            }

            public String execute(Map<String, Object> args) {
                var incident = store.createIncident(
                        String.valueOf(args.get("title")),
                        String.valueOf(args.get("service")),
                        Severity.valueOf(String.valueOf(args.get("severity")).toUpperCase()));
                return "Incidente aberto: " + incident.id();
            }
        };
    }

    public static Tool resolveIncident(OpsStore store) {
        return new Tool() {
            public String name() {
                return "resolve_incident";
            }

            public String description() {
                return "Resolve um incidente existente (id, summary).";
            }

            public String execute(Map<String, Object> args) {
                var incident = store.resolveIncident(
                        String.valueOf(args.get("id")),
                        args.containsKey("summary") ? String.valueOf(args.get("summary")) : null);
                return "Incidente resolvido: " + incident.id();
            }
        };
    }

    public static Tool consultarRunbook(OpsStore store) {
        return new Tool() {
            public String name() {
                return "consultar_runbook";
            }

            public String description() {
                return "Consulta o runbook de um servico.";
            }

            public String execute(Map<String, Object> args) {
                return store.getRunbook(String.valueOf(args.get("service"))).content();
            }
        };
    }

    /**
     * Porte simplificado de check-provider-status.ts: o original faz uma
     * checagem HTTP real contra status pages de provedores. Aqui e um stub
     * plugavel (sem chamada de rede real, ver README) para manter o nucleo
     * testavel sem depender de rede externa.
     */
    public static Tool checkProviderStatus() {
        return new Tool() {
            public String name() {
                return "check_provider_status";
            }

            public String description() {
                return "Verifica o status operacional dos provedores externos.";
            }

            public String execute(Map<String, Object> args) {
                return "Todos os provedores monitorados reportam operacional (stub, sem chamada de rede real).";
            }
        };
    }
}
