// Agent Loop Framework — implementa o ciclo Percepção→Planejamento→Ação→
// Avaliação com critérios de parada explícitos (maxSteps, timeout,
// no-progress, circuit breaker), exposto via POST /agent/run. Porte Go do
// projeto Spring Boot agent-loop-framework-java.
package main

import (
	"fmt"
	"log"
	"net/http"

	"agent-loop-framework/agentloop"
	"agent-loop-framework/core"
	"agent-loop-framework/observability"
	"agent-loop-framework/openai"
	"agent-loop-framework/tool"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	apiKey := getEnv("OPENAI_API_KEY", "sk-placeholder")
	model := getEnv("OPENAI_MODEL", "gpt-4o-mini")

	contract := core.NewAgentContract()
	contract.MaxSteps = getEnvInt("AGENT_MAX_STEPS", contract.MaxSteps)
	contract.MaxTokens = getEnvInt("AGENT_MAX_TOKENS", contract.MaxTokens)
	contract.MaxTimeSeconds = getEnvInt("AGENT_MAX_TIME", contract.MaxTimeSeconds)
	contract.NoProgressSteps = getEnvInt("AGENT_NO_PROGRESS_STEPS", contract.NoProgressSteps)

	chatClient := openai.NewClient(apiKey, model)

	tools := []tool.AgentTool{
		tool.GetMetricsTool{},
		tool.GetLogsTool{},
		tool.GetDeployHistoryTool{},
		&tool.SaveIncidentTool{},
	}

	loop := &agentloop.AgentLoop{
		Planner:        &core.Planner{Client: chatClient},
		Executor:       core.NewExecutor(tools),
		CircuitBreaker: core.NewCircuitBreaker(),
		Telemetry:      &observability.AgentTelemetry{},
		Contract:       contract,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /agent/run", runHandler(loop))
	mux.HandleFunc("GET /agent/health", healthHandler)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("agent-loop-framework ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
