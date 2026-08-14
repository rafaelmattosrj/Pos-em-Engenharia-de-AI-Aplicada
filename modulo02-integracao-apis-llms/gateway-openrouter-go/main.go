// Smart Model Router Gateway — servidor HTTP que expõe POST /chat e roteia
// a pergunta por uma cadeia de modelos de fallback via OpenRouter. Porte Go
// do projeto Spring Boot gateway-openrouter-java.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"gateway-openrouter/gateway"
	"gateway-openrouter/openrouter"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	systemPrompt := getEnv("APP_SYSTEM_PROMPT", "You are a helpful assistant.")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.2)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 100)
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	client := &gateway.ResilientClient{
		OpenRouter:     openrouter.NewClient(apiKey),
		FallbackModels: fallbackModels,
		SystemPrompt:   systemPrompt,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", gateway.ChatHandler(client))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("gateway-openrouter ouvindo em %s (fallback: %v)\n", addr, fallbackModels)
	log.Fatal(http.ListenAndServe(addr, mux))
}
