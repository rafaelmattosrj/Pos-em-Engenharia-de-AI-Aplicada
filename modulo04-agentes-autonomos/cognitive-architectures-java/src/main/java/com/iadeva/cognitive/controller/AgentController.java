package com.iadeva.cognitive.controller;

import com.iadeva.cognitive.agent.PlanExecuteAgent;
import com.iadeva.cognitive.agent.ReactAgent;
import com.iadeva.cognitive.agent.ReflectionAgent;
import com.iadeva.cognitive.model.AgentRequest;
import com.iadeva.cognitive.model.AgentResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * Controller responsável por receber requisições e rotear para a
 * arquitetura cognitiva correta com base no campo "architecture".
 *
 * POST /agents/run
 *   Body: { "architecture": "react|plan-execute|reflection", "input": "<tarefa>" }
 */
@RestController
@RequestMapping("/agents")
public class AgentController {

    private static final Logger log = LoggerFactory.getLogger(AgentController.class);

    private final ReactAgent reactAgent;
    private final PlanExecuteAgent planExecuteAgent;
    private final ReflectionAgent reflectionAgent;

    public AgentController(ReactAgent reactAgent,
                           PlanExecuteAgent planExecuteAgent,
                           ReflectionAgent reflectionAgent) {
        this.reactAgent = reactAgent;
        this.planExecuteAgent = planExecuteAgent;
        this.reflectionAgent = reflectionAgent;
    }

    /**
     * Executa um agente cognitivo com base na arquitetura solicitada.
     *
     * @param request contém o tipo de arquitetura e a tarefa (input)
     * @return AgentResponse com resultado e métricas de execução
     */
    @PostMapping("/run")
    public ResponseEntity<AgentResponse> run(@RequestBody AgentRequest request) {
        log.info("[Controller] Requisição recebida — architecture={}, input={}",
                request.architecture(), request.input());

        if (request.architecture() == null || request.architecture().isBlank()) {
            return ResponseEntity.badRequest()
                    .body(new AgentResponse("Campo 'architecture' é obrigatório.",
                            new AgentResponse.AgentMetrics(0, 0, 0)));
        }

        if (request.input() == null || request.input().isBlank()) {
            return ResponseEntity.badRequest()
                    .body(new AgentResponse("Campo 'input' é obrigatório.",
                            new AgentResponse.AgentMetrics(0, 0, 0)));
        }

        AgentResponse response = switch (request.architecture().toLowerCase().trim()) {
            case "react"         -> reactAgent.execute(request.input());
            case "plan-execute"  -> planExecuteAgent.execute(request.input());
            case "reflection"    -> reflectionAgent.execute(request.input());
            default -> new AgentResponse(
                    "Arquitetura desconhecida: '" + request.architecture() + "'. "
                    + "Valores válidos: react, plan-execute, reflection.",
                    new AgentResponse.AgentMetrics(0, 0, 0)
            );
        };

        log.info("[Controller] Execução concluída — steps={}, tokens={}, reflections={}",
                response.metrics().steps(),
                response.metrics().tokens(),
                response.metrics().reflections());

        return ResponseEntity.ok(response);
    }
}
