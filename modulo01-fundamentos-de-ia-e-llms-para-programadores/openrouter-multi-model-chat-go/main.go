// Demonstração de chat com múltiplos modelos gratuitos via OpenRouter — porte
// Go do projeto Java equivalente (openrouter-multi-model-chat-java), que por
// sua vez replica o exemplo original em shell/curl (Exemplo 11).
//
// O OpenRouter (https://openrouter.ai) é um proxy unificado que expõe dezenas
// de modelos através de uma única API compatível com o formato OpenAI. Aqui,
// em vez de uma abstração como o LangChain4j, usamos um cliente HTTP mínimo
// (pacote openrouter/) — o ecossistema Go não tem uma biblioteca de LLM tão
// madura e amplamente adotada quanto o LangChain4j do Java, então a
// implementação idiomática é falar HTTP diretamente.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"openrouter-multi-model-chat/openrouter"
)

// pergunta enviada para todos os modelos.
const pergunta = "Me conte uma curiosidade sobre LLMs (Large Language Models)."

// modelos gratuitos disponíveis no OpenRouter.
var modelos = []string{
	"google/gemma-3-27b-it:free",
	"meta-llama/llama-3.2-3b-instruct:free",
	"mistralai/mistral-7b-instruct:free",
}

func main() {
	loadDotEnv(".env")

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ERRO: A variável OPENROUTER_API_KEY não foi encontrada.")
		fmt.Fprintln(os.Stderr, "Crie um arquivo .env na raiz do projeto com o conteúdo:")
		fmt.Fprintln(os.Stderr, "  OPENROUTER_API_KEY=sk-or-...")
		os.Exit(1)
	}

	client := openrouter.NewClient(apiKey)

	linha := strings.Repeat("=", 60)
	fmt.Println(linha)
	fmt.Println("  OpenRouter Multi-Model Chat em Go")
	fmt.Println(linha)
	fmt.Println("Pergunta enviada para todos os modelos:")
	fmt.Printf("  %q\n", pergunta)
	fmt.Println(linha)

	ctx := context.Background()

	for _, modelo := range modelos {
		fmt.Println()
		fmt.Println("Modelo:", modelo)
		fmt.Println(strings.Repeat("-", 60))

		resposta, err := client.Complete(ctx, modelo, pergunta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao chamar o modelo [%s]: %v\n", modelo, err)
		} else {
			fmt.Println(resposta)
		}

		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Println()
	fmt.Println(linha)
	fmt.Println("  Fim da demonstração.")
	fmt.Println(linha)
}
