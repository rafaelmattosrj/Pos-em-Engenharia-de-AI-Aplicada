package tiering

import "context"

// Gateway abstrai as duas operações do Ollama usadas neste protótipo: chat
// com streaming (o original usa stream:true em gerarComTier) e embeddings.
// *ollama.Client satisfaz esta interface. Extraída para permitir testar
// CascadeGateway com um dublê determinístico, sem depender de um Ollama
// local rodando durante `go test`.
type Gateway interface {
	ChatStream(ctx context.Context, model, systemPrompt, userPrompt string, onChunk func(string)) (string, error)
	Embed(ctx context.Context, model, texto string) ([]float64, error)
}

// ApprovalPrompt é o Approval Gate (Módulo 4.4): pede aprovação humana para
// um rascunho antes de virar oficial. Extraída como interface para permitir
// testar CascadeGateway sem esperar entrada real de stdin durante `go test`.
type ApprovalPrompt interface {
	Approve(rascunho string) (bool, error)
}

// ApprovalPromptFunc adapta uma função comum a ApprovalPrompt.
type ApprovalPromptFunc func(rascunho string) (bool, error)

func (f ApprovalPromptFunc) Approve(rascunho string) (bool, error) {
	return f(rascunho)
}
