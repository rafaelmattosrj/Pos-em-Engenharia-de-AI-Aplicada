// Agent Memory — agente com consciência de memória, integrando memória de
// curto prazo, longo prazo, episódica e contextual (embeddings), com um
// motor de reflexão evolutiva que extrai lições de execuções passadas.
// Porte Go do projeto Spring Boot agent-memory-java.
package main

import (
	"fmt"
	"log"
	"net/http"

	"agent-memory/agent"
	"agent-memory/engine"
	"agent-memory/memory"
	"agent-memory/openai"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	apiKey := getEnv("OPENAI_API_KEY", "sk-placeholder")
	chatModel := getEnv("OPENAI_MODEL", "gpt-4o-mini")
	embeddingModel := getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")
	memoryPath := getEnv("MEMORY_PATH", "./data")
	similarityThreshold := getEnvFloat("SIMILARITY_THRESHOLD", 0.7)

	client := openai.NewClient(apiKey, chatModel, embeddingModel)

	longTerm := memory.NewLongTermMemory(memoryPath)
	episodic := memory.NewEpisodicMemory(memoryPath)
	contextual := &memory.ContextualMemory{Embedder: client, SimilarityThreshold: similarityThreshold}
	reflection := engine.NewReflectionEngine(client, memoryPath)

	memoryAgent := &agent.MemoryAwareAgent{
		LongTerm:   longTerm,
		Episodic:   episodic,
		Contextual: contextual,
		Reflection: reflection,
		Client:     client,
	}

	srv := &server{
		memoryAgent: memoryAgent,
		episodic:    episodic,
		reflection:  reflection,
		chatClient:  client,
		memoryPath:  memoryPath,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /agent-memory/run", srv.runWithMemory)
	mux.HandleFunc("POST /agent-memory/run-without", srv.runWithoutMemory)
	mux.HandleFunc("GET /agent-memory/episodes", srv.getEpisodes)
	mux.HandleFunc("GET /agent-memory/lessons", srv.getLessons)
	mux.HandleFunc("DELETE /agent-memory/clear", srv.clearMemory)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("agent-memory ouvindo em %s (memoria em: %s)\n", addr, memoryPath)
	log.Fatal(http.ListenAndServe(addr, mux))
}
