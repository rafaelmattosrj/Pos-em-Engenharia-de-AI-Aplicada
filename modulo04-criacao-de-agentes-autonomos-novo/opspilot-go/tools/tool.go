// Package tools contém o contrato de ferramenta (porte de
// DynamicStructuredTool, agents/tools.ts, simplificado) e as ferramentas
// concretas usadas pelos papéis do time (analista/executor) e pela estratégia
// ReAct de referência.
package tools

// Tool é o porte simplificado de DynamicStructuredTool (agents/tools.ts).
// Execute retorna um error explícito para entrada inválida (ex.: severidade
// desconhecida) -- diferente do porte Java, que deixa uma
// IllegalArgumentException não tratada estourar a chamada inteira; aqui o
// chamador (ReactStrategy / RoleRunner) transforma o erro numa observação de
// trace, mantendo o loop do agente vivo.
type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]any) (string, error)
}

// basicTool é uma implementação funcional de Tool, equivalente aos objetos
// anônimos usados no porte Java (OpsTools).
type basicTool struct {
	name        string
	description string
	exec        func(args map[string]any) (string, error)
}

func (t *basicTool) Name() string        { return t.name }
func (t *basicTool) Description() string { return t.description }
func (t *basicTool) Execute(args map[string]any) (string, error) {
	return t.exec(args)
}
