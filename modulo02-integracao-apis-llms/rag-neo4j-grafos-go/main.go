// RAG com Neo4j Knowledge Graph e geração de Cypher via LLM. Porte Go do
// projeto Spring Boot rag-neo4j-grafos-java: converte perguntas em
// linguagem natural em Cypher, executa contra o Neo4j com autocorreção via
// LLM em caso de erro de sintaxe, e gera a resposta final em PT-BR.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rag-neo4j-grafos/graph"
	"rag-neo4j-grafos/llm"
	"rag-neo4j-grafos/neo4jservice"
	"rag-neo4j-grafos/openrouter"
	"rag-neo4j-grafos/service"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "4000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.2)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 500)
	maxCorrectionAttempts := getEnvInt("APP_MAX_CORRECTION_ATTEMPTS", 1)
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	neo4jURI := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnv("NEO4J_USERNAME", "neo4j")
	neo4jPassword := getEnv("NEO4J_PASSWORD", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	neo4jSvc, err := neo4jservice.NewService(ctx, neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		log.Fatalf("falha ao conectar ao Neo4j: %v", err)
	}
	defer neo4jSvc.Close(context.Background())

	resilientClient := &llm.ResilientClient{
		OpenRouter:     openrouter.NewClient(apiKey),
		FallbackModels: fallbackModels,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
	}

	orchestrator := &graph.Orchestrator{
		Neo4j:                 neo4jSvc,
		CypherGenerator:       &service.CypherGeneratorService{Client: resilientClient},
		AnalyticalResponse:    &service.AnalyticalResponseService{Client: resilientClient},
		MaxCorrectionAttempts: maxCorrectionAttempts,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", chatHandler(orchestrator))
	mux.HandleFunc("/seed", seedHandler(neo4jSvc))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("rag-neo4j-grafos ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
