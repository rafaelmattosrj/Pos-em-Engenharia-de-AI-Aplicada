package com.iadeva.agent.core;

import com.iadeva.agent.model.PlanDecision;
import com.iadeva.agent.model.ToolResult;
import com.iadeva.agent.tool.AgentTool;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;

/**
 * Executor — despacha para a ferramenta correta e retorna resultado.
 * Equivalente ao executor.py — despacha para a ferramenta correta e retorna resultado.
 *
 * Recebe um PlanDecision, localiza a tool pelo nome e executa.
 * Retorna ToolResult com sucesso ou erro.
 */
@Component
public class Executor {

    private static final Logger log = LoggerFactory.getLogger(Executor.class);

    // Mapa de tools indexado pelo nome — injetado automaticamente pelo Spring
    private final Map<String, AgentTool> toolRegistry;

    public Executor(List<AgentTool> tools) {
        this.toolRegistry = tools.stream()
                .collect(Collectors.toMap(AgentTool::getName, Function.identity()));
        log.info("Tools registradas: {}", toolRegistry.keySet());
    }

    /**
     * Executa a tool indicada no PlanDecision.
     *
     * @param plan Decisão do Planner com nome da tool e argumentos
     * @return ToolResult com o resultado da execução
     */
    public ToolResult execute(PlanDecision plan) {
        if (plan.tool() == null || plan.tool().isBlank()) {
            return ToolResult.fail("none", "Nenhuma tool especificada no plano");
        }

        AgentTool tool = toolRegistry.get(plan.tool());

        if (tool == null) {
            String msg = String.format(
                    "Tool '%s' não encontrada. Tools disponíveis: %s",
                    plan.tool(), toolRegistry.keySet()
            );
            log.warn(msg);
            return ToolResult.fail(plan.tool(), msg);
        }

        log.info("Executando tool '{}' com args: {}", plan.tool(), plan.args());

        try {
            String output = tool.execute(plan.args() != null ? plan.args() : Map.of());
            return ToolResult.ok(plan.tool(), output);
        } catch (Exception e) {
            log.error("Erro ao executar tool '{}': {}", plan.tool(), e.getMessage());
            return ToolResult.fail(plan.tool(), "Erro de execução: " + e.getMessage());
        }
    }
}
