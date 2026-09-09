// PSP Routing Intelligence — servidor HTTP que recomenda o Payment Service
// Provider (PSP) com maior chance de aprovacao para uma transacao, usando
// Embeddings + RAG + LLM sobre transacoes historicas armazenadas no Neo4j.
//
// Porte Go de psp-routing-intelligence-java (Spring Boot + LangChain4j),
// que por sua vez implementa a especificacao de
// ../psp-routing-intelligence-java/IDEIA.md — ver README.md deste projeto
// para o detalhamento de paridade e adaptacoes.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"psp-routing-intelligence/embeddings"
	"psp-routing-intelligence/httpserver"
	"psp-routing-intelligence/openrouter"
	"psp-routing-intelligence/routing"
	"psp-routing-intelligence/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	neo4jURI := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnv("NEO4J_USER", "neo4j")
	neo4jPassword := getEnv("NEO4J_PASSWORD", "password")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	nlpModel := getEnv("NLP_MODEL", "google/gemma-3-27b-it:free")
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_EMBED_MODEL", "all-minilm")

	fmt.Println("Inicializando PSP Routing Intelligence (Go)...")

	vectorStore, err := store.NewStore(ctx, neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		log.Fatalf("erro ao conectar ao Neo4j: %v (suba o Neo4j com: docker-compose up -d)", err)
	}
	defer vectorStore.Close(ctx)

	if err := vectorStore.EnsureIndex(ctx, embeddings.Dimension); err != nil {
		log.Fatalf("erro ao criar indice vetorial: %v", err)
	}

	embeddingClient := embeddings.NewClient(ollamaURL, ollamaModel)
	chatClient := openrouter.NewClient(apiKey, nlpModel)

	routingService := &routing.Service{
		Embed:       embeddingClient.Embed,
		FindSimilar: vectorStore.FindSimilar,
		Chat:        chatClient.Chat,
	}

	seedService := &routing.SeedService{
		Clear: vectorStore.Clear,
		Embed: embeddingClient.Embed,
		Store: vectorStore.Store,
	}

	handler := httpserver.New(routingService, seedService)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		fmt.Printf("Servindo em http://localhost:%s (POST /api/routing/recommend, POST /api/routing/seed)\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro no servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nEncerrando...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
