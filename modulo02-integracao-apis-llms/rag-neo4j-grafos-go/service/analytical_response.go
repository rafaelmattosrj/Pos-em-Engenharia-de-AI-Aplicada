package service

import (
	"context"
	"fmt"
)

const analyticalResponseSystemPrompt = "Você é um analista de dados educacionais. Responda perguntas sobre alunos e cursos.\nSeja direto, use os dados fornecidos, responda em português."

// AnalyticalResponseService gera a resposta final em linguagem natural a
// partir dos resultados do Neo4j — equivalente a AnalyticalResponseService.java.
type AnalyticalResponseService struct {
	Client ResilientCaller
}

// GenerateResponse produz a resposta em PT-BR para a pergunta original, dado
// o resultado bruto da query — equivalente a
// AnalyticalResponseService.generateResponse.
func (s *AnalyticalResponseService) GenerateResponse(ctx context.Context, question string, dbResults []map[string]any) (string, error) {
	userPrompt := fmt.Sprintf("Pergunta: %s\n\nDados do banco: %v\n\nResponda de forma clara e útil.", question, dbResults)
	return s.Client.Call(ctx, analyticalResponseSystemPrompt, userPrompt)
}
