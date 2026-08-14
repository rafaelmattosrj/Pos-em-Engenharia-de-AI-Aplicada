package tiering

import (
	"context"
	"strings"
	"sync"
)

// fakeGateway é um dublê determinístico de Gateway: embeddings e respostas de
// chat pré-configurados por texto/modelo — sem chamar nenhum Ollama de
// verdade, e sem latência de rede (essencial para os testes de concorrência).
type fakeGateway struct {
	mu              sync.Mutex
	embeddings      map[string][]float64
	embeddingPadrao []float64
	chats           map[string]string // chave: modelo + "|" + trecho do prompt
	chatPadrao      string
}

func newFakeGateway() *fakeGateway {
	return &fakeGateway{
		embeddings:      map[string][]float64{},
		embeddingPadrao: []float64{0, 0},
		chats:           map[string]string{},
		chatPadrao:      "resposta padrão",
	}
}

func (f *fakeGateway) comEmbedding(texto string, vetor []float64) *fakeGateway {
	f.embeddings[texto] = vetor
	return f
}

func (f *fakeGateway) comEmbeddingPadrao(vetor []float64) *fakeGateway {
	f.embeddingPadrao = vetor
	return f
}

func (f *fakeGateway) comChat(modelo, trechoDoPrompt, resposta string) *fakeGateway {
	f.chats[modelo+"|"+trechoDoPrompt] = resposta
	return f
}

func (f *fakeGateway) comChatPadrao(resposta string) *fakeGateway {
	f.chatPadrao = resposta
	return f
}

func (f *fakeGateway) ChatStream(ctx context.Context, modelo, systemPrompt, userPrompt string, onChunk func(string)) (string, error) {
	f.mu.Lock()
	resposta := f.chatPadrao
	for chave, r := range f.chats {
		partes := strings.SplitN(chave, "|", 2)
		if partes[0] == modelo && strings.Contains(userPrompt, partes[1]) {
			resposta = r
			break
		}
	}
	f.mu.Unlock()
	onChunk(resposta)
	return resposta, nil
}

func (f *fakeGateway) Embed(ctx context.Context, modelo, texto string) ([]float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if v, ok := f.embeddings[texto]; ok {
		return v, nil
	}
	return f.embeddingPadrao, nil
}

// fakeApprovalPrompt é um dublê determinístico de ApprovalPrompt: devolve uma
// resposta fixa, sem ler stdin.
type fakeApprovalPrompt struct {
	resposta       bool
	chamadas       int
	ultimoRascunho string
}

func newFakeApprovalPrompt(resposta bool) *fakeApprovalPrompt {
	return &fakeApprovalPrompt{resposta: resposta}
}

func (f *fakeApprovalPrompt) Approve(rascunho string) (bool, error) {
	f.chamadas++
	f.ultimoRascunho = rascunho
	return f.resposta, nil
}
