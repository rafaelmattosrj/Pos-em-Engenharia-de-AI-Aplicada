package guardrail

import (
	"context"

	"manipulation-guardrail-prototype/ollama"
)

// OllamaClassifier é a implementação real de ClassifierClient, via Ollama local.
type OllamaClassifier struct {
	Client *ollama.Client
	Model  string
}

func NewOllamaClassifier(client *ollama.Client, model string) *OllamaClassifier {
	return &OllamaClassifier{Client: client, Model: model}
}

func (o *OllamaClassifier) Classify(ctx context.Context, pergunta string) (string, error) {
	return o.Client.Chat(ctx, o.Model, []ollama.Message{
		{Role: "system", Content: InstrucaoClassificador},
		{Role: "user", Content: pergunta},
	})
}
