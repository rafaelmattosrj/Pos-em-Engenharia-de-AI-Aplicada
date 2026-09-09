package routing

import (
	_ "embed"
	"fmt"
	"strings"

	"psp-routing-intelligence/domain"
)

// promptTemplate e o template do prompt RAG, extraido 1:1 de IDEIA.md e
// embutido no binario em tempo de compilacao (go:embed) — equivalente a
// carregar resources/prompts/routing-prompt.txt do classpath na versao Java.
//
//go:embed routing-prompt.txt
var promptTemplate string

// BuildPrompt monta o prompt final substituindo {transaction} e
// {similarCases} no template — equivalente a RoutingPromptBuilder.build da
// versao Java.
func BuildPrompt(tx domain.Transaction, similarCases []domain.SimilarCase) string {
	prompt := strings.ReplaceAll(promptTemplate, "{transaction}", SerializeTransaction(tx))
	prompt = strings.ReplaceAll(prompt, "{similarCases}", describeSimilarCases(similarCases))
	return prompt
}

func describeSimilarCases(similarCases []domain.SimilarCase) string {
	if len(similarCases) == 0 {
		return "(nenhum caso historico similar encontrado)"
	}

	var b strings.Builder
	for i, c := range similarCases {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "- PSP=%s, status=%s, valor=R$%.2f, metodo=%s, similaridade=%.2f",
			c.PSP, c.Status, c.Amount, c.Method, c.Similarity)
	}
	return b.String()
}
