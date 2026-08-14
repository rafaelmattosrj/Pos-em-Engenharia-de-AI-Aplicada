// Roteamento condicional — pipeline com nós + aresta condicional (identifica
// intenção, roteia para UPPERCASE/LOWERCASE/fallback via LLM), exposto via
// POST /chat. Porte Go do projeto Spring Boot roteamento-condicional-java
// (que por sua vez substitui o StateGraph do LangGraph por nós explícitos).
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"roteamento-condicional/graph"
	"roteamento-condicional/openrouter"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.2)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 200)
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	orchestrator := &graph.Orchestrator{
		Fallback: &graph.FallbackNode{
			OpenRouter:     openrouter.NewClient(apiKey),
			FallbackModels: fallbackModels,
			Temperature:    temperature,
			MaxTokens:      maxTokens,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", chatHandler(orchestrator))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("roteamento-condicional ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
