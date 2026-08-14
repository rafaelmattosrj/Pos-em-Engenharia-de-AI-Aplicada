// Pipeline RAG completo (Retrieval-Augmented Generation) sobre um PDF,
// usando Neo4j como vector store e um LLM via OpenRouter — porte Go de
// pdf-rag-knowledge-base-java.
package main

import (
	"context"
	"fmt"
	"strings"

	"pdf-rag-knowledge-base/embeddings"
	"pdf-rag-knowledge-base/openrouter"
	"pdf-rag-knowledge-base/pdfx"
	"pdf-rag-knowledge-base/splitter"
	"pdf-rag-knowledge-base/store"
)

const promptTemplate = `Você é um assistente especializado em TensorFlow.js e machine learning.

**Contexto e Regras:**
- Tarefa: Responder perguntas sobre TensorFlow.js e machine learning de forma educacional
- Tom de voz: educacional e amigável
- Idioma: pt-BR
- Formato de resposta: texto natural com exemplos

**Instruções importantes:**
1. Use APENAS as informações do contexto fornecido para responder
2. Se o contexto não contiver informação suficiente, diga que não encontrou a informação
3. Seja claro, objetivo e use exemplos quando apropriado
4. Mantenha um tom educacional e amigável
5. Se houver código ou exemplos no contexto, inclua-os na resposta
6. Responda em português de forma natural e conversacional
7. Estruture sua resposta em parágrafos quando necessário
8. Use analogias e exemplos práticos para facilitar o entendimento

**Pergunta do usuário:**
%s

**Contexto recuperado do documento:**
%s

**Resposta:**
Forneça uma resposta clara, educacional e em português. Use exemplos do contexto quando disponível.
`

const minScore = 0.5

func main() {
	ctx := context.Background()

	fmt.Println("🚀 Inicializando sistema de RAG com Neo4j (Go)...")
	fmt.Println()

	loadDotEnv(".env")
	neo4jURI := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := getEnv("NEO4J_USER", "neo4j")
	neo4jPassword := getEnv("NEO4J_PASSWORD", "password")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	nlpModel := getEnv("NLP_MODEL", "")
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_EMBED_MODEL", "all-minilm")

	// ─── ETAPA 1: Carrega e divide o PDF ─────────────────────────────────
	fmt.Println("📄 Carregando PDF (tensores.pdf)...")
	text, err := pdfx.ExtractText("tensores.pdf")
	if err != nil {
		fmt.Println("❌ Erro ao carregar o PDF:", err)
		fmt.Println("   Copie o arquivo: cp ../pdf-rag-knowledge-base-java/tensores.pdf ./tensores.pdf")
		return
	}
	fmt.Println("✅ PDF carregado com sucesso")

	chunks := splitter.Recursive(text, 1000, 200)
	fmt.Printf("✂️  Dividido em %d chunks\n\n", len(chunks))

	// ─── ETAPA 2: Carrega modelo de embeddings local (via Ollama) ────────
	fmt.Println("🧠 Carregando modelo de embeddings local (Ollama: all-minilm)...")
	embedClient := embeddings.NewClient(ollamaURL, ollamaModel)
	fmt.Println("✅ Modelo de embeddings carregado")
	fmt.Println()

	// ─── ETAPA 3: Conecta ao Neo4j e limpa dados anteriores ──────────────
	fmt.Println("🗑️  Removendo documentos existentes no Neo4j...")
	vectorStore, err := store.NewStore(ctx, neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		fmt.Println("❌ Erro ao conectar ao Neo4j:", err)
		fmt.Println("   Suba o Neo4j com: docker-compose up -d")
		return
	}
	defer vectorStore.Close(ctx)

	if err := vectorStore.Clear(ctx); err != nil {
		fmt.Println("❌ Erro ao limpar dados anteriores:", err)
		return
	}
	if err := vectorStore.EnsureIndex(ctx, embeddings.Dimension); err != nil {
		fmt.Println("❌ Erro ao criar indice vetorial:", err)
		return
	}
	fmt.Println("✅ Dados anteriores removidos")
	fmt.Println()

	// ─── ETAPA 4: Gera embeddings e armazena no Neo4j ────────────────────
	fmt.Println("📥 Gerando embeddings e armazenando no Neo4j...")
	for i, chunk := range chunks {
		vector, err := embedClient.Embed(ctx, chunk)
		if err != nil {
			fmt.Println("❌ Erro ao gerar embedding:", err)
			return
		}
		if err := vectorStore.Add(ctx, chunk, vector); err != nil {
			fmt.Println("❌ Erro ao armazenar chunk:", err)
			return
		}
		fmt.Printf("   ✅ Chunk %d/%d armazenado\n", i+1, len(chunks))
	}
	fmt.Println("\n✅ Base de dados populada com sucesso!")
	fmt.Println()

	// ─── ETAPA 5: Configura o LLM via OpenRouter ─────────────────────────
	chatClient := openrouter.NewClient(apiKey, nlpModel)

	// ─── ETAPA 6: Pipeline RAG — perguntas e respostas ───────────────────
	questions := []string{
		"Como converter objetos JavaScript em tensores?",
		"O que é normalização de dados e por que é necessária?",
		"Como funciona uma rede neural no TensorFlow.js?",
		"O que significa treinar uma rede neural?",
		"o que é hot enconding e quando usar?",
	}

	fmt.Println("🔍 ETAPA 2: Executando buscas por similaridade...")
	fmt.Println()

	for _, question := range questions {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("📌 PERGUNTA:", question)
		fmt.Println(strings.Repeat("=", 80))

		questionEmbedding, err := embedClient.Embed(ctx, question)
		if err != nil {
			fmt.Println("❌ Erro ao gerar embedding da pergunta:", err)
			continue
		}

		fmt.Println("🔍 Buscando no vector store do Neo4j...")
		matches, err := vectorStore.Search(ctx, questionEmbedding, 3)
		if err != nil {
			fmt.Println("❌ Erro na busca:", err)
			continue
		}

		if len(matches) == 0 {
			fmt.Println("⚠️  Nenhum resultado encontrado na base de conhecimento.")
			fmt.Println()
			continue
		}

		fmt.Printf("✅ Encontrados %d resultados relevantes (melhor score: %.3f)\n", len(matches), matches[0].Score)

		var relevant []string
		for _, m := range matches {
			if m.Score > minScore {
				relevant = append(relevant, m.Text)
			}
		}

		if len(relevant) == 0 {
			fmt.Printf("⚠️  Nenhum resultado com score suficiente (> %.1f).\n\n", minScore)
			continue
		}

		ragContext := strings.Join(relevant, "\n\n---\n\n")

		fmt.Println("🤖 Gerando resposta com IA...")
		prompt := fmt.Sprintf(promptTemplate, question, ragContext)
		answer, err := chatClient.Chat(ctx, prompt)
		if err != nil {
			msg := err.Error()
			if len(msg) > 200 {
				msg = msg[:200]
			}
			fmt.Println("❌ Erro ao chamar LLM:", msg)
			fmt.Println()
			continue
		}
		fmt.Println("\n" + answer)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ Processamento concluído com sucesso!")
	fmt.Println()
}
