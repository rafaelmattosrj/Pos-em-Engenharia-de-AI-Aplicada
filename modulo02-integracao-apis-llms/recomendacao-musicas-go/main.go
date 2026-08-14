// Chat de recomendação de músicas com memória de conversa e preferências
// persistidas em SQLite. Porte Go do projeto Spring Boot
// recomendacao-musicas-java.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"recomendacao-musicas/graph"
	"recomendacao-musicas/llm"
	"recomendacao-musicas/openrouter"
	"recomendacao-musicas/repository"
	"recomendacao-musicas/service"
)

func main() {
	ctx := context.Background()
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.7)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 500)
	summarizeAfter := int64(getEnvInt("APP_SUMMARIZE_AFTER_MESSAGES", 10))
	dbPath := getEnv("DB_PATH", "./data/musicdb.sqlite")
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	db, err := repository.Open(ctx, dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	resilientClient := &llm.ResilientClient{
		OpenRouter:     openrouter.NewClient(apiKey),
		FallbackModels: fallbackModels,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
	}

	orchestrator := &graph.Orchestrator{
		Client: resilientClient,
		MemoryService: &service.MemoryService{
			Repository: &repository.ConversationRepository{DB: db},
		},
		PreferencesService: &service.PreferencesService{
			Repository: &repository.PreferencesRepository{DB: db},
			Client:     resilientClient,
		},
		SummarizeAfterCount: summarizeAfter,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", chatHandler(orchestrator))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("recomendacao-musicas ouvindo em %s (db: %s)\n", addr, dbPath)
	log.Fatal(http.ListenAndServe(addr, mux))
}
