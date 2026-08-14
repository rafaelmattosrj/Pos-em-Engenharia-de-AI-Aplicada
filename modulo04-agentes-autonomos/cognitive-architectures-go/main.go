// Cognitive Architectures — 3 arquiteturas cognitivas (ReAct, Plan-and-
// Execute, Reflection) expostas via POST /agents/run, roteadas pelo campo
// "architecture" do request. Porte Go do projeto Spring Boot
// cognitive-architectures-java.
package main

import (
	"fmt"
	"log"
	"net/http"

	"cognitive-architectures/agent"
	"cognitive-architectures/openai"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	apiKey := getEnv("OPENAI_API_KEY", "sk-placeholder")
	model := getEnv("OPENAI_MODEL", "gpt-4o-mini")
	reflectionThreshold := getEnvFloat("REFLECTION_THRESHOLD", 0.7)
	reflectionMaxCycles := getEnvInt("REFLECTION_MAX_CYCLES", 3)
	reactMaxSteps := getEnvInt("REACT_MAX_STEPS", 10)

	client := openai.NewClient(apiKey, model)

	critiqueEvaluator := &agent.CritiqueEvaluator{Client: client, Threshold: reflectionThreshold}

	srv := &server{
		react:       &agent.ReactAgent{Client: client, MaxSteps: reactMaxSteps},
		planExecute: &agent.PlanExecuteAgent{Client: client},
		reflection:  &agent.ReflectionAgent{Client: client, Critique: critiqueEvaluator, MaxCycles: reflectionMaxCycles},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /agents/run", srv.run)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("cognitive-architectures ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
