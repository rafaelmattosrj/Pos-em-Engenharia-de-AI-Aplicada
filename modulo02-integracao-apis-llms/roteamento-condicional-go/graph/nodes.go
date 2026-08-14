package graph

import "strings"

// IdentifyIntent analisa o último input e define o comando — equivalente a
// IdentifyIntentNode.process.
func IdentifyIntent(state State) State {
	input := strings.ToLower(state.Messages[len(state.Messages)-1])

	var command Command
	switch {
	case strings.Contains(input, "upper"):
		command = CommandUppercase
	case strings.Contains(input, "lower"):
		command = CommandLowercase
	default:
		command = CommandUnknown
	}

	return state.WithCommand(command)
}

// UpperCase converte o output para maiúsculas — equivalente a UpperCaseNode.
func UpperCase(state State) State {
	return state.WithOutput(strings.ToUpper(state.Output))
}

// LowerCase converte o output para minúsculas — equivalente a LowerCaseNode.
func LowerCase(state State) State {
	return state.WithOutput(strings.ToLower(state.Output))
}

// ChatResponse é o nó final: o output já foi transformado pelos nós
// anteriores — equivalente a ChatResponseNode (identidade).
func ChatResponse(state State) State {
	return state
}
