package com.iadeva.agent.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Persiste incidentes em memória durante a execução da aplicação.
 * Tool simulada — em produção persistiria no banco de dados de incidentes (PagerDuty, Jira, ServiceNow).
 */
@Component
public class SaveIncidentTool implements AgentTool {

    private static final Logger log = LoggerFactory.getLogger(SaveIncidentTool.class);

    // Armazenamento em memória — em produção seria um repositório JPA ou cliente de API
    private final List<Map<String, Object>> incidentStore = new CopyOnWriteArrayList<>();

    private final ObjectMapper objectMapper;

    public SaveIncidentTool(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Override
    public String getName() {
        return "saveIncident";
    }

    @Override
    public String execute(Map<String, Object> args) {
        String id = "INC-" + UUID.randomUUID().toString().substring(0, 8).toUpperCase();
        String title = (String) args.getOrDefault("title", "Incidente sem título");
        String severity = (String) args.getOrDefault("severity", "medium");
        String description = (String) args.getOrDefault("description", "Sem descrição");

        Map<String, Object> incident = Map.of(
                "id", id,
                "title", title,
                "severity", severity,
                "description", description,
                "status", "OPEN",
                "created_at", Instant.now().toString(),
                "created_by", "DevOps Agent"
        );

        incidentStore.add(incident);
        log.info("[SaveIncidentTool] Incidente criado: {} — {}", id, title);

        try {
            return objectMapper.writeValueAsString(Map.of(
                    "success", true,
                    "incident_id", id,
                    "message", "Incidente registrado com sucesso",
                    "incident", incident
            ));
        } catch (Exception e) {
            return "{\"success\": false, \"error\": \"Falha ao serializar confirmação\"}";
        }
    }

    /** Retorna todos os incidentes salvos em memória (útil para testes e auditoria) */
    public List<Map<String, Object>> getAllIncidents() {
        return new ArrayList<>(incidentStore);
    }
}
