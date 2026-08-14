package graph

import "context"

// Orchestrator executa o fluxo:
// START → IdentifyIntent → (aresta condicional) → UpperCase/LowerCase/Fallback → ChatResponse → END
// Equivalente a WorkflowOrchestrator.invoke.
type Orchestrator struct {
	Fallback *FallbackNode
}

// Invoke roda o pipeline completo para o input do usuário.
func (o *Orchestrator) Invoke(ctx context.Context, input string) (State, error) {
	state := InitialState(input)
	state = IdentifyIntent(state)

	var err error
	switch state.Command {
	case CommandUppercase:
		state = UpperCase(state)
	case CommandLowercase:
		state = LowerCase(state)
	default:
		state, err = o.Fallback.Process(ctx, state)
		if err != nil {
			return state, err
		}
	}

	return ChatResponse(state), nil
}
