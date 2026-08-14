// Package graph orquestra o pipeline de análise: IntentNode → ExecutorNode —
// equivalente a AnalysisOrchestrator.java (que por sua vez espelha o grafo
// LangGraph do curso).
package graph

import (
	"context"

	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/node"
)

// Orchestrator conduz o pipeline de análise de vendas.
type Orchestrator struct {
	IntentNode   *node.IntentNode
	ExecutorNode *node.ExecutorNode
}

// Analyze extrai a intenção estruturada e executa a análise com tools —
// equivalente a AnalysisOrchestrator.analyze.
func (o *Orchestrator) Analyze(ctx context.Context, request model.AnalysisRequest) (model.AnalysisResult, error) {
	intent := o.IntentNode.Extract(ctx, request)
	return o.ExecutorNode.Execute(ctx, intent)
}
