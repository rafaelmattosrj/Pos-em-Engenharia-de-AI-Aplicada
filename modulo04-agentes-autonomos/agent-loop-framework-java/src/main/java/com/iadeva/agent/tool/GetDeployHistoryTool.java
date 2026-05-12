package com.iadeva.agent.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

/**
 * Retorna o histórico de deploys recentes simulado de um serviço.
 * Tool simulada — em produção integraria com CI/CD pipeline (Jenkins, GitHub Actions, ArgoCD).
 */
@Component
public class GetDeployHistoryTool implements AgentTool {

    private final ObjectMapper objectMapper;

    public GetDeployHistoryTool(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Override
    public String getName() {
        return "getDeployHistory";
    }

    @Override
    public String execute(Map<String, Object> args) {
        String service = (String) args.getOrDefault("service", "api-gateway");

        // Histórico de deploys simulado — em produção viria do CI/CD pipeline
        List<Map<String, Object>> deploys = List.of(
                Map.of(
                        "id", "deploy-20240115-001",
                        "timestamp", "2024-01-15T14:15:00Z",
                        "service", service,
                        "version", "v2.3.1",
                        "previous_version", "v2.3.0",
                        "status", "SUCCESS",
                        "deployed_by", "ci-pipeline",
                        "changes", "Refatoração do connection pool e aumento de timeout para queries de pagamento",
                        "rollback_available", true
                ),
                Map.of(
                        "id", "deploy-20240114-003",
                        "timestamp", "2024-01-14T22:30:00Z",
                        "service", service,
                        "version", "v2.3.0",
                        "previous_version", "v2.2.9",
                        "status", "SUCCESS",
                        "deployed_by", "rafael.mattos",
                        "changes", "Correção de bug no fluxo de autenticação OAuth",
                        "rollback_available", true
                ),
                Map.of(
                        "id", "deploy-20240113-002",
                        "timestamp", "2024-01-13T10:00:00Z",
                        "service", service,
                        "version", "v2.2.9",
                        "previous_version", "v2.2.8",
                        "status", "ROLLED_BACK",
                        "deployed_by", "ci-pipeline",
                        "changes", "Migração de dependência do driver JDBC",
                        "rollback_available", false,
                        "rollback_reason", "Aumento de 300% na latência do banco após deploy"
                )
        );

        try {
            return objectMapper.writeValueAsString(Map.of(
                    "service", service,
                    "total_deploys", deploys.size(),
                    "last_deploy", deploys.getFirst().get("timestamp"),
                    "deploys", deploys
            ));
        } catch (Exception e) {
            return "{\"error\": \"Falha ao serializar histórico de deploys\"}";
        }
    }
}
