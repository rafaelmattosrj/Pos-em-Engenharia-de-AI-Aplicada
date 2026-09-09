package tools

import (
	"fmt"
	"strings"

	"opspilot/domain"
	"opspilot/store"
)

// Porte de agents/tools.ts -- ferramentas concretas usadas pelos papéis
// analista (leitura) e executor (mutação de incidentes), conforme a partição
// estrutural documentada em team-strategy.ts (spec 018 FR-007).

func ListAlerts(s store.OpsStore) Tool {
	return &basicTool{
		name:        "list_alerts",
		description: "Lista os alertas disparando ou resolvidos.",
		exec: func(args map[string]any) (string, error) {
			status, err := optionalAlertStatus(args)
			if err != nil {
				return "", err
			}
			alerts := s.GetAlerts(status)
			lines := make([]string, 0, len(alerts))
			for _, a := range alerts {
				lines = append(lines, fmt.Sprintf("[%s] %s (%s): %s", a.Severity, a.Service, a.Status, a.Description))
			}
			return strings.Join(lines, "\n"), nil
		},
	}
}

func ListIncidents(s store.OpsStore) Tool {
	return &basicTool{
		name:        "list_incidents",
		description: "Lista incidentes abertos ou resolvidos.",
		exec: func(args map[string]any) (string, error) {
			status, err := optionalIncidentStatus(args)
			if err != nil {
				return "", err
			}
			incidents := s.GetIncidents(status)
			if len(incidents) == 0 {
				return "(nenhum incidente)", nil
			}
			lines := make([]string, 0, len(incidents))
			for _, inc := range incidents {
				lines = append(lines, fmt.Sprintf("%s [%s/%s] %s (%s)", inc.ID, inc.Severity, inc.Status, inc.Title, inc.Service))
			}
			return strings.Join(lines, "\n"), nil
		},
	}
}

func OpenIncident(s store.OpsStore) Tool {
	return &basicTool{
		name:        "open_incident",
		description: "Abre um novo incidente (title, service, severity).",
		exec: func(args map[string]any) (string, error) {
			severity, err := requiredSeverity(args)
			if err != nil {
				return "", err
			}
			title := stringArg(args, "title")
			service := stringArg(args, "service")
			incident := s.CreateIncident(title, service, severity)
			return "Incidente aberto: " + incident.ID, nil
		},
	}
}

func ResolveIncident(s store.OpsStore) Tool {
	return &basicTool{
		name:        "resolve_incident",
		description: "Resolve um incidente existente (id, summary).",
		exec: func(args map[string]any) (string, error) {
			id := stringArg(args, "id")
			var summary *string
			if raw, ok := args["summary"]; ok && raw != nil {
				s := fmt.Sprintf("%v", raw)
				summary = &s
			}
			incident, err := s.ResolveIncident(id, summary)
			if err != nil {
				return "", err
			}
			return "Incidente resolvido: " + incident.ID, nil
		},
	}
}

func ConsultarRunbook(s store.OpsStore) Tool {
	return &basicTool{
		name:        "consultar_runbook",
		description: "Consulta o runbook de um serviço.",
		exec: func(args map[string]any) (string, error) {
			runbook, err := s.GetRunbook(stringArg(args, "service"))
			if err != nil {
				return "", err
			}
			return runbook.Content, nil
		},
	}
}

// CheckProviderStatus é o porte simplificado de check-provider-status.ts: o
// original faz uma checagem HTTP real contra status pages de provedores.
// Aqui é um stub plugável (sem chamada de rede real, ver README) para manter
// o núcleo testável sem depender de rede externa.
func CheckProviderStatus() Tool {
	return &basicTool{
		name:        "check_provider_status",
		description: "Verifica o status operacional dos provedores externos.",
		exec: func(args map[string]any) (string, error) {
			return "Todos os provedores monitorados reportam operacional (stub, sem chamada de rede real).", nil
		},
	}
}

func stringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func optionalAlertStatus(args map[string]any) (*domain.AlertStatus, error) {
	raw, ok := args["status"]
	if !ok || raw == nil {
		return nil, nil
	}
	status := domain.AlertStatus(strings.ToUpper(fmt.Sprintf("%v", raw)))
	if status != domain.AlertFiring && status != domain.AlertResolved {
		return nil, fmt.Errorf("status de alerta desconhecido: %q", raw)
	}
	return &status, nil
}

func optionalIncidentStatus(args map[string]any) (*domain.IncidentStatus, error) {
	raw, ok := args["status"]
	if !ok || raw == nil {
		return nil, nil
	}
	status := domain.IncidentStatus(strings.ToUpper(fmt.Sprintf("%v", raw)))
	if status != domain.IncidentOpen && status != domain.IncidentResolved {
		return nil, fmt.Errorf("status de incidente desconhecido: %q", raw)
	}
	return &status, nil
}

func requiredSeverity(args map[string]any) (domain.Severity, error) {
	raw := stringArg(args, "severity")
	severity := domain.Severity(strings.ToUpper(raw))
	switch severity {
	case domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow:
		return severity, nil
	default:
		return "", fmt.Errorf("severidade desconhecida: %q", raw)
	}
}
