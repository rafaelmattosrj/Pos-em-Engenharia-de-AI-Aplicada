package com.iadeva.agent.controller;

import com.iadeva.agent.core.AgentLoop;
import com.iadeva.agent.model.AgentTrace;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * Endpoint REST para executar o agente loop via HTTP.
 * Recebe a tarefa inicial e retorna o resultado e trace completo da execução.
 */
@RestController
@RequestMapping("/agent")
public class AgentController {

    private static final Logger log = LoggerFactory.getLogger(AgentController.class);

    private final AgentLoop agentLoop;

    public AgentController(AgentLoop agentLoop) {
        this.agentLoop = agentLoop;
    }

    /**
     * Executa o agente loop com a tarefa fornecida.
     *
     * POST /agent/run
     * Body: { "input": "Diagnosticar degradação no serviço api-gateway" }
     *
     * @param body Map contendo o campo "input" com a tarefa
     * @return resultado final e trace completo da execução
     */
    @PostMapping("/run")
    public ResponseEntity<Map<String, Object>> run(@RequestBody Map<String, String> body) {
        String input = body.get("input");

        if (input == null || input.isBlank()) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "Campo 'input' é obrigatório e não pode estar vazio"));
        }

        log.info("[AgentController] Recebendo tarefa: {}", input);

        try {
            AgentTrace trace = agentLoop.run(input);

            return ResponseEntity.ok(Map.of(
                    "result", trace.getFinalResult(),
                    "trace", trace
            ));
        } catch (Exception e) {
            log.error("[AgentController] Erro durante execução do agente: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", "Erro durante execução do agente: " + e.getMessage()));
        }
    }

    /**
     * Endpoint de health check.
     * GET /agent/health
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of(
                "status", "UP",
                "service", "agent-loop-framework"
        ));
    }
}
