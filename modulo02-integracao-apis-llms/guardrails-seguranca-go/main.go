// Guardrails de segurança contra prompt injection — verifica cada mensagem
// com um modelo dedicado antes de encaminhá-la ao chat principal. Porte Go
// do projeto Spring Boot guardrails-seguranca-java.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"guardrails-seguranca/graph"
	"guardrails-seguranca/llm"
	"guardrails-seguranca/openrouter"
	"guardrails-seguranca/service"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.7)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 1000)
	guardrailsModel := getEnv("APP_GUARDRAILS_MODEL", "openai/gpt-oss-safeguard-20b")
	guardrailsEnabled := getEnv("APP_GUARDRAILS_ENABLED", "true") == "true"
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	openrouterClient := openrouter.NewClient(apiKey)

	orchestrator := &graph.Orchestrator{
		Guardrails: &service.GuardrailsService{
			Client:  openrouterClient,
			Model:   guardrailsModel,
			Enabled: guardrailsEnabled,
		},
		Chat: &service.ChatService{
			Client: &llm.ResilientClient{
				OpenRouter:     openrouterClient,
				FallbackModels: fallbackModels,
				Temperature:    temperature,
				MaxTokens:      maxTokens,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", chatHandler(orchestrator))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("guardrails-seguranca ouvindo em %s (guardrails habilitado: %v)\n", addr, guardrailsEnabled)
	log.Fatal(http.ListenAndServe(addr, mux))
}
