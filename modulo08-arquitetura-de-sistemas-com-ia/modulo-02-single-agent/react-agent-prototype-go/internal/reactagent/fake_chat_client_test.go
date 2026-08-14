package reactagent

import "context"

// evento e uma resposta enfileirada no fakeChatClient: ou uma ModeloResposta
// de sucesso, ou um erro a devolver.
type evento struct {
	resposta ModeloResposta
	erro     error
	ehErro   bool
}

// fakeChatClient e um dublê determinístico de ChatClient, usado nos testes
// de AgenteICF e RetryingChatClient — evita depender de um Ollama local
// rodando durante `go test`.
type fakeChatClient struct {
	eventos  []evento
	Chamadas int
}

func (f *fakeChatClient) comResposta(resposta ModeloResposta) *fakeChatClient {
	f.eventos = append(f.eventos, evento{resposta: resposta})
	return f
}

func (f *fakeChatClient) comFalha(erro error) *fakeChatClient {
	f.eventos = append(f.eventos, evento{erro: erro, ehErro: true})
	return f
}

func (f *fakeChatClient) Chat(_ context.Context, _ []Mensagem, _ []Mensagem) (ModeloResposta, error) {
	f.Chamadas++
	if len(f.eventos) == 0 {
		panic("fakeChatClient: nenhum evento enfileirado")
	}
	ev := f.eventos[0]
	f.eventos = f.eventos[1:]
	if ev.ehErro {
		return ModeloResposta{}, ev.erro
	}
	return ev.resposta, nil
}
