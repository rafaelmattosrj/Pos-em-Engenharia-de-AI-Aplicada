// Package service contém a geração/correção de Cypher e a resposta
// analítica final — equivalentes a CypherGeneratorService.java e
// AnalyticalResponseService.java.
package service

import (
	"context"
	"fmt"
	"strings"
)

// ResilientCaller é satisfeito por llm.ResilientClient — abstraído aqui para
// permitir fakes nos testes do orchestrator sem depender do OpenRouter real.
type ResilientCaller interface {
	Call(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

const cypherGeneratorSystemPrompt = `Você é um especialista em Neo4j Cypher. Converta perguntas em linguagem natural para queries Cypher.

Schema do banco:
%s

Regras:
- Retorne APENAS a query Cypher, sem explicações, sem markdown
- Use MATCH, WHERE, RETURN conforme necessário
- Para relações de compra, use: (s:Student)-[:ENROLLED_IN]->(c:Course)
- Prefira RETURN com propriedades específicas em vez de nós inteiros`

const cypherCorrectionSystemPrompt = `Corrija a query Cypher que falhou. Retorne APENAS a query corrigida.

Schema: %s`

// CypherGeneratorService gera e corrige queries Cypher a partir de linguagem
// natural, via um ResilientCaller — equivalente a CypherGeneratorService.java.
type CypherGeneratorService struct {
	Client ResilientCaller
}

// GenerateCypher converte a pergunta em Cypher — equivalente a
// CypherGeneratorService.generateCypher.
func (s *CypherGeneratorService) GenerateCypher(ctx context.Context, question, schema string) (string, error) {
	raw, err := s.Client.Call(ctx, fmt.Sprintf(cypherGeneratorSystemPrompt, schema), question)
	if err != nil {
		return "", err
	}
	return cleanCypher(raw), nil
}

// CorrectCypher pede ao LLM para corrigir uma query que falhou — equivalente
// a CypherGeneratorService.correctCypher.
func (s *CypherGeneratorService) CorrectCypher(ctx context.Context, failedQuery, errMessage, schema string) (string, error) {
	userPrompt := fmt.Sprintf("Query com erro:\n%s\n\nErro: %s\n\nRetorne a query corrigida:", failedQuery, errMessage)
	raw, err := s.Client.Call(ctx, fmt.Sprintf(cypherCorrectionSystemPrompt, schema), userPrompt)
	if err != nil {
		return "", err
	}
	return cleanCypher(raw), nil
}

// cleanCypher remove cercas markdown que o LLM eventualmente devolve, mesmo
// com a instrução explícita de não usá-las — mesma tolerância da versão Java.
func cleanCypher(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.ReplaceAll(cleaned, "```cypher", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	return strings.TrimSpace(cleaned)
}
