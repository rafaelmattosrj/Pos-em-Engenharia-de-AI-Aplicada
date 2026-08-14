// Demonstração de Embeddings + Busca por Similaridade com Neo4j (sem LLM) —
// porte Go de embeddings-vector-search-java.
//
// Passos:
//  1. Carrega tensores.pdf e divide em chunks (1000 chars / 200 overlap)
//  2. Gera embeddings localmente via Ollama (modelo all-minilm, 384 dim)
//  3. Limpa dados anteriores no Neo4j e cria o índice vetorial
//  4. Armazena os vetores no Neo4j
//  5. Executa busca por similaridade para 5 perguntas
//  6. Exibe os 3 chunks mais relevantes com score e preview de texto
//     -- NENHUMA chamada a LLM --
package main

import (
	"context"
	"fmt"
	"strings"

	"embeddings-vector-search/embeddings"
	"embeddings-vector-search/pdfx"
	"embeddings-vector-search/splitter"
	"embeddings-vector-search/store"
)

const previewLength = 300

func main() {
	ctx := context.Background()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  Embeddings + Vector Search com Neo4j  (Go)")
	fmt.Println("  Modelo local via Ollama: all-minilm  |  Sem chamada a LLM")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	loadDotEnv(".env")
	neo4jURI := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnv("NEO4J_USER", "neo4j")
	neo4jPassword := getEnv("NEO4J_PASSWORD", "password")
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_EMBED_MODEL", "all-minilm")

	// ── ETAPA 1: Carrega e divide o PDF ─────────────────────────────────
	fmt.Println("ETAPA 1: Carregando PDF (tensores.pdf)...")
	text, err := pdfx.ExtractText("tensores.pdf")
	if err != nil {
		fmt.Println("  Erro ao carregar o PDF:", err)
		fmt.Println("  Copie o arquivo: cp ../pdf-rag-knowledge-base-java/tensores.pdf ./tensores.pdf")
		return
	}
	fmt.Println("  PDF carregado com sucesso.")

	chunks := splitter.Recursive(text, 1000, 200)
	fmt.Printf("  Dividido em %d chunks (tamanho=1000, overlap=200)\n\n", len(chunks))

	// ── ETAPA 2: Cliente de embeddings local (via Ollama) ───────────────
	fmt.Println("ETAPA 2: Conectando ao modelo de embeddings local (Ollama: all-minilm)...")
	embedClient := embeddings.NewClient(ollamaURL, ollamaModel)
	fmt.Printf("  Cliente pronto. Dimensao do vetor: %d\n\n", embeddings.Dimension)

	// ── ETAPA 3: Conecta ao Neo4j e limpa dados anteriores ──────────────
	fmt.Println("ETAPA 3: Limpando dados anteriores no Neo4j...")
	vectorStore, err := store.NewStore(ctx, neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		fmt.Println("  Erro ao conectar ao Neo4j:", err)
		fmt.Println("  Suba o Neo4j com: docker-compose up -d")
		return
	}
	defer vectorStore.Close(ctx)

	if err := vectorStore.Clear(ctx); err != nil {
		fmt.Println("  Erro ao limpar dados anteriores:", err)
		return
	}
	fmt.Println("  Dados anteriores removidos.")

	if err := vectorStore.EnsureIndex(ctx, embeddings.Dimension); err != nil {
		fmt.Println("  Erro ao criar indice vetorial:", err)
		return
	}
	fmt.Println()

	// ── ETAPA 4: Gera embeddings e armazena no Neo4j ────────────────────
	fmt.Println("ETAPA 4: Gerando embeddings e armazenando no Neo4j...")
	for i, chunk := range chunks {
		vector, err := embedClient.Embed(ctx, chunk)
		if err != nil {
			fmt.Println("  Erro ao gerar embedding:", err)
			return
		}
		if err := vectorStore.Add(ctx, chunk, vector); err != nil {
			fmt.Println("  Erro ao armazenar chunk:", err)
			return
		}
		fmt.Printf("  [%d/%d] chunk armazenado\n", i+1, len(chunks))
	}
	fmt.Println("\n  Base de dados populada com sucesso!")
	fmt.Println()

	// ── ETAPA 5: Busca por similaridade (sem LLM) ───────────────────────
	fmt.Println("ETAPA 5: Executando buscas por similaridade...")
	fmt.Println()

	questions := []string{
		"O que sao tensores e como sao representados em JavaScript?",
		"Como converter objetos JavaScript em tensores?",
		"O que e normalizacao de dados e por que e necessaria?",
		"Como funciona uma rede neural no TensorFlow.js?",
		"O que e hot encoding e quando usar?",
	}

	for _, question := range questions {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("PERGUNTA:", question)
		fmt.Println(strings.Repeat("=", 80))

		questionEmbedding, err := embedClient.Embed(ctx, question)
		if err != nil {
			fmt.Println("  Erro ao gerar embedding da pergunta:", err)
			continue
		}

		matches, err := vectorStore.Search(ctx, questionEmbedding, 3)
		if err != nil {
			fmt.Println("  Erro na busca:", err)
			continue
		}

		if len(matches) == 0 {
			fmt.Println("  Nenhum resultado encontrado.")
			fmt.Println()
			continue
		}

		fmt.Printf("  Encontrados %d resultados:\n\n", len(matches))

		for i, match := range matches {
			preview := match.Text
			if len(preview) > previewLength {
				preview = preview[:previewLength] + "..."
			}

			fmt.Printf("  Resultado #%d  |  Score: %.4f\n", i+1, match.Score)
			fmt.Println("  " + strings.Repeat("-", 60))
			for _, line := range strings.Split(preview, "\n") {
				fmt.Println("  " + line)
			}
			fmt.Println()
		}
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  Processamento concluido.")
	fmt.Println("  Nenhum LLM foi chamado — apenas embeddings + busca vetorial.")
	fmt.Println(strings.Repeat("=", 80))
}
