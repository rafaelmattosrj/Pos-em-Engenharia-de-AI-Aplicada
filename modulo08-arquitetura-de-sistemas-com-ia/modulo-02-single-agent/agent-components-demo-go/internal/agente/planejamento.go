package agente

import (
	"context"
	"time"
)

// ChatClient e a abstracao do modelo de chat usado pela secao de
// Planejamento. Extraida como interface (o original em JS/Python chama o
// SDK do Ollama diretamente) para manter a chamada de rede isolada e
// substituivel — nao porque os testes precisem de um dublê: a secao de
// Planejamento nao tem teste automatizado, igual ao original (so
// observacao ao vivo).
type ChatClient interface {
	Chat(ctx context.Context, mensagens []Mensagem) (string, error)
}

// ResultadoPlanejamento contem o texto da resposta, quanto tempo levou, e
// quantas chamadas ao modelo foram feitas.
type ResultadoPlanejamento struct {
	Texto     string
	DuracaoMs int64
	Chamadas  int
}

// ChainOfThought pede pro modelo pensar passo a passo antes de responder,
// em uma unica chamada.
func ChainOfThought(ctx context.Context, client ChatClient, pergunta string) (ResultadoPlanejamento, error) {
	inicio := time.Now()
	resposta, err := client.Chat(ctx, []Mensagem{
		{Role: "user", Content: "Pense passo a passo, de forma breve, antes de responder: " + pergunta},
	})
	if err != nil {
		return ResultadoPlanejamento{}, err
	}
	return ResultadoPlanejamento{
		Texto:     resposta,
		DuracaoMs: time.Since(inicio).Milliseconds(),
		Chamadas:  1,
	}, nil
}

// ChainOfThoughtMaisReflexao faz uma segunda chamada, pedindo pro modelo
// criticar a propria resposta — chamada inteira a mais, nao so raciocinio
// mais longo na mesma chamada.
func ChainOfThoughtMaisReflexao(ctx context.Context, client ChatClient, pergunta string) (ResultadoPlanejamento, error) {
	inicio := time.Now()

	primeira, err := client.Chat(ctx, []Mensagem{{Role: "user", Content: pergunta}})
	if err != nil {
		return ResultadoPlanejamento{}, err
	}

	critica, err := client.Chat(ctx, []Mensagem{
		{Role: "user", Content: pergunta},
		{Role: "assistant", Content: primeira},
		{Role: "user", Content: "Releia sua resposta acima. Ela tem algum erro ou imprecisão? Responda só \"sim\" ou \"não\" e, se sim, qual."},
	})
	if err != nil {
		return ResultadoPlanejamento{}, err
	}

	return ResultadoPlanejamento{
		Texto:     critica,
		DuracaoMs: time.Since(inicio).Milliseconds(),
		Chamadas:  2,
	}, nil
}
