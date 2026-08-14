// Agente analisador de vendas: pipeline IntentNode → ExecutorNode com
// function calling (tool local csvToJson). Porte Go do projeto Spring Boot
// mcp-sales-analyzer-java.
package main

import (
	"fmt"
	"log"
	"net/http"

	"mcp-sales-analyzer/graph"
	"mcp-sales-analyzer/node"
	"mcp-sales-analyzer/openai"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	apiKey := getEnv("OPENAI_API_KEY", "sk-placeholder")
	model := getEnv("OPENAI_MODEL", "gpt-4o-mini")

	client := openai.NewClient(apiKey, model)

	orchestrator := &graph.Orchestrator{
		IntentNode:   &node.IntentNode{Client: client},
		ExecutorNode: &node.ExecutorNode{Client: client},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /analyze", analyzeHandler(orchestrator))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("mcp-sales-analyzer ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
