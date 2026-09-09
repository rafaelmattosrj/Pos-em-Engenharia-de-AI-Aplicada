package llm

import "opspilot/tools"

// ChatModel é o porte da interface ChatModel / fachada OpsChatModel
// (agents/model.ts). Diferente do porte Java (que usa exceções não
// verificadas), Invoke retorna um error explícito -- idiomático em Go.
type ChatModel interface {
	Invoke(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error)
}

// ChatModelFunc adapta uma função comum para a interface ChatModel, no
// espírito de http.HandlerFunc -- útil para modelos fake definidos inline em
// testes (equivalente ao lambda `(messages, tools) -> ...` usado no porte
// Java para ChatModel).
type ChatModelFunc func(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error)

func (f ChatModelFunc) Invoke(messages []ChatMessage, availableTools []tools.Tool) (ModelResponse, error) {
	return f(messages, availableTools)
}
