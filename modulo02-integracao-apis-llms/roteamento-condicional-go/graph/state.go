// Package graph implementa o fluxo de roteamento condicional — porte do
// pacote graph/ da versão Java, que por sua vez substitui o StateGraph do
// LangGraph por um pipeline explícito de nós + aresta condicional.
package graph

// Command identifica a intenção detectada no input — equivalente ao enum
// GraphState.Command.
type Command int

const (
	CommandUnknown Command = iota
	CommandUppercase
	CommandLowercase
)

// State é o estado imutável que flui pelo grafo — equivalente ao record
// GraphState. Cada nó recebe um State e retorna um novo State (sem mutação).
type State struct {
	Messages []string
	Output   string
	Command  Command
}

// InitialState cria o estado inicial a partir do input do usuário —
// equivalente a GraphState.initial(input).
func InitialState(input string) State {
	return State{
		Messages: []string{input},
		Output:   input,
		Command:  CommandUnknown,
	}
}

// WithOutput retorna uma cópia do estado com um novo Output.
func (s State) WithOutput(output string) State {
	s.Output = output
	return s
}

// WithCommand retorna uma cópia do estado com um novo Command.
func (s State) WithCommand(command Command) State {
	s.Command = command
	return s
}
