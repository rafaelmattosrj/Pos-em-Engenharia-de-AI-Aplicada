package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"agent-loop-framework/model"
)

// ChatClient é a única operação de que o Planner depende — extraída como
// interface para permitir um duplo de teste simples, equivalente ao mock de
// ChatClient em Java.
type ChatClient interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

const plannerPromptTemplate = `## Contexto atual (Percepção)
%s

## Situação
- Step atual: %d de %d
- Tools já usadas: %s
- Tempo decorrido: %d segundos

## Tools disponíveis
- getMetrics: Obtém métricas de latência (p50, p95, p99) de um serviço. Args: { "service": "nome-do-servico" }
- getLogs: Obtém últimas linhas de log de um serviço. Args: { "service": "nome-do-servico", "lines": 20 }
- getDeployHistory: Obtém histórico de deploys recentes. Args: { "service": "nome-do-servico" }
- saveIncident: Salva um incidente. Args: { "title": "...", "severity": "high|medium|low", "description": "..." }

## Instrução
Analise o contexto e decida a próxima ação. Responda APENAS com JSON válido no formato abaixo, sem markdown:

{
  "reasoning": "Seu raciocínio sobre a situação",
  "action": "Descrição clara da ação",
  "tool": "nomeDaTool ou null se tarefa concluída",
  "args": { "chave": "valor" },
  "successCriteria": "Como saber que o resultado foi satisfatório",
  "done": false
}

Se a tarefa estiver concluída, use "done": true e "tool": null.
`

// Planner consulta o LLM e retorna uma PlanDecision estruturada —
// equivalente a Planner.java.
type Planner struct {
	Client ChatClient
}

// Plan gera a próxima decisão de planejamento com base na percepção atual.
func (p *Planner) Plan(ctx context.Context, perception string, state *AgentState, contract AgentContract) model.PlanDecision {
	prompt := p.buildPrompt(perception, state, contract)

	response, err := p.Client.Chat(ctx, contract.SystemPrompt(), prompt)
	if err != nil {
		log.Printf("[Planner] Erro ao consultar LLM: %v", err)
		return model.PlanDecision{
			Reasoning: "Erro ao consultar LLM: " + err.Error(),
			Action:    "Encerrar execução por erro",
			Args:      map[string]any{},
			Done:      true,
		}
	}

	return p.parseResponse(response)
}

func (p *Planner) buildPrompt(perception string, state *AgentState, contract AgentContract) string {
	perceptionText := perception
	if strings.TrimSpace(perception) == "" {
		perceptionText = "Nenhum contexto ainda — início da execução."
	}

	toolsText := "nenhuma"
	if len(state.LastToolsUsed()) > 0 {
		toolsText = strings.Join(state.LastToolsUsed(), ", ")
	}

	return fmt.Sprintf(plannerPromptTemplate,
		perceptionText,
		state.CurrentStep(),
		contract.MaxSteps,
		toolsText,
		state.ElapsedSeconds(),
	)
}

func (p *Planner) parseResponse(response string) model.PlanDecision {
	cleaned := strings.TrimSpace(response)
	cleaned = strings.ReplaceAll(cleaned, "```json", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	cleaned = strings.TrimSpace(cleaned)

	var decision model.PlanDecision
	if err := json.Unmarshal([]byte(cleaned), &decision); err != nil {
		log.Printf("[Planner] Falha ao parsear resposta do LLM: %v. Resposta: %s", err, response)
		return model.PlanDecision{
			Reasoning: "Resposta inválida do LLM",
			Action:    "INVALID",
			Args:      map[string]any{},
			Done:      false,
		}
	}
	if decision.Args == nil {
		decision.Args = map[string]any{}
	}
	return decision
}
