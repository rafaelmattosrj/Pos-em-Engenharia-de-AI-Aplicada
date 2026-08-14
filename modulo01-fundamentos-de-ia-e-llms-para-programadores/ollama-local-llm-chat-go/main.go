// Demonstração de chat com LLM local via Ollama — porte Go do projeto Java
// equivalente (ollama-local-llm-chat-java). O Ollama expõe uma API
// OpenAI-compatible em http://localhost:11434/v1.
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ollama-local-llm-chat/ollama"
)

var perguntas = []string{
	"Explique o que e inteligencia artificial em 3 frases simples.",
	"Qual e a diferenca entre machine learning e deep learning?",
	"O que e um LLM (Large Language Model) e como ele funciona?",
	"Cite 3 casos de uso praticos de IA na engenharia de software.",
	"O que e RAG (Retrieval-Augmented Generation) e para que serve?",
}

func main() {
	loadDotEnv(".env")

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Ollama Local LLM Chat - Go")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	baseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434/v1")
	modelName := getEnv("OLLAMA_MODEL", "llama3.2")

	fmt.Println("Configuracao:")
	fmt.Println("  Base URL :", baseURL)
	fmt.Println("  Modelo   :", modelName)
	fmt.Println()

	client := ollama.NewClient(baseURL, modelName)
	ctx := context.Background()

	for i, pergunta := range perguntas {
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("[%d/%d] PERGUNTA:\n%s\n\n", i+1, len(perguntas), pergunta)

		inicio := time.Now()
		resposta, err := client.Chat(ctx, pergunta)
		duracao := time.Since(inicio)

		if err != nil {
			fmt.Println("ERRO ao chamar o modelo:", err)
			fmt.Println()
			fmt.Println("Verifique se o Ollama esta rodando:")
			fmt.Println("  ollama serve")
			fmt.Println("  ollama pull", modelName)
		} else {
			fmt.Println("RESPOSTA:")
			fmt.Println(resposta)
			fmt.Printf("\n(tempo: %d ms)\n", duracao.Milliseconds())
		}

		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Sessao de chat encerrada.")
	fmt.Println(strings.Repeat("=", 70))
}
