package core

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"agent-loop-framework/model"
	"agent-loop-framework/tool"
)

// Executor despacha para a tool correta e retorna o resultado — equivalente
// a Executor.java.
type Executor struct {
	registry map[string]tool.AgentTool
}

// NewExecutor cria um Executor indexando as tools pelo nome.
func NewExecutor(tools []tool.AgentTool) *Executor {
	registry := make(map[string]tool.AgentTool, len(tools))
	for _, t := range tools {
		registry[t.Name()] = t
	}
	log.Printf("Tools registradas: %v", toolNames(registry))
	return &Executor{registry: registry}
}

// Execute executa a tool indicada em plan.
func (e *Executor) Execute(plan model.PlanDecision) model.ToolResult {
	if strings.TrimSpace(plan.Tool) == "" {
		return model.FailResult("none", "Nenhuma tool especificada no plano")
	}

	t, ok := e.registry[plan.Tool]
	if !ok {
		msg := fmt.Sprintf("Tool '%s' não encontrada. Tools disponíveis: %v", plan.Tool, toolNames(e.registry))
		log.Println(msg)
		return model.FailResult(plan.Tool, msg)
	}

	log.Printf("Executando tool '%s' com args: %v", plan.Tool, plan.Args)

	args := plan.Args
	if args == nil {
		args = map[string]any{}
	}
	output := t.Execute(args)
	return model.OkResult(plan.Tool, output)
}

func toolNames(registry map[string]tool.AgentTool) []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
