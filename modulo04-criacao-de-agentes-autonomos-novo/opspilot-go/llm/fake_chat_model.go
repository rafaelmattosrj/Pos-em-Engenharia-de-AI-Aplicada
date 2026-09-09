package llm

import (
	"errors"
	"sync"

	"opspilot/tools"
)

// Responder simula comportamento dependente do input, equivalente ao
// `Function<List<ChatMessage>, ModelResponse>` do porte Java.
type Responder func(messages []ChatMessage) (ModelResponse, error)

// FakeChatModel é o test double sem chamada de rede. O UNIDADE.md da unidade
// 9 registra que "o código passa no typecheck e nos testes com fakes" --
// este é o equivalente Go desses fakes: uma fila de respostas programadas, ou
// uma função arbitrária.
type FakeChatModel struct {
	mu          sync.Mutex
	scripted    []ModelResponse
	responder   Responder
	callHistory [][]ChatMessage
}

var _ ChatModel = (*FakeChatModel)(nil)

// NewFakeChatModelScripted cria um FakeChatModel que consome a fila de
// respostas programadas em ordem, uma por chamada.
func NewFakeChatModelScripted(responses []ModelResponse) *FakeChatModel {
	scripted := make([]ModelResponse, len(responses))
	copy(scripted, responses)
	return &FakeChatModel{scripted: scripted}
}

// NewFakeChatModelFunc cria um FakeChatModel cuja resposta é calculada por
// uma função arbitrária a partir das mensagens recebidas.
func NewFakeChatModelFunc(responder Responder) *FakeChatModel {
	return &FakeChatModel{responder: responder}
}

func (f *FakeChatModel) Invoke(messages []ChatMessage, _ []tools.Tool) (ModelResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callHistory = append(f.callHistory, messages)

	if f.responder != nil {
		return f.responder(messages)
	}
	if len(f.scripted) == 0 {
		return ModelResponse{}, errors.New("FakeChatModel ran out of scripted responses")
	}
	next := f.scripted[0]
	f.scripted = f.scripted[1:]
	return next, nil
}

// CallCount é o porte de FakeChatModel.callCount() (porte Java).
func (f *FakeChatModel) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.callHistory)
}
