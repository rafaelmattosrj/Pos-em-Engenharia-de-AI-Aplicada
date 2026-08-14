// Package service implementa a geração do Brag Document — equivalente a
// bragGeneratorFlow de flows.ts.
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"brag-bot/model"
)

// GeminiClient é a única operação de que BragService depende.
type GeminiClient interface {
	GenerateJSON(ctx context.Context, prompt string, temperature float64) (string, error)
}

const promptTemplate = `Persona: Você deve atuar como um "Senior Career Consultant" focado em Planos de Desenvolvimento Individual (IDP) para Engenheiros de Software.
Objetivo: Transformar o rascunho informal do usuário em um "Brag Document" executivo.

Regra 1: Usar tom profissional, objetivo e focado em impacto, sem adjetivos emocionais.
Regra 2: Se não existirem métricas exatas, infira a natureza da métrica baseada na ação tomada de forma plausível (ex: "tempo de execução não especificado mas otimizado").
Regra 3: Responda APENAS com um JSON válido, sem markdown, com os campos exatos:
  title (string): Ação principal + Resultado de alto nível.
  context (string): Situação ou Problema original. O que estava quebrado, lento, o desafio, etc.
  actionTaken (string): Ação técnica ou estratégica passo a passo tomada para resolver o problema.
  businessImpact (string): Qual o impacto de negócio. Tempo ganho, redução de falhas, etc.
  metrics (array de strings): Apenas dados estritamente quantificáveis. Ex: "50% reduction", "10ms latency".
  technologiesUsed (array de strings): Ferramentas, linguagens, bibliotecas e plataformas mencionadas ou inferidas.
Regra 4: O output deve respeitar o idioma original do input.

Aqui está o rascunho informal do usuário:
`

const generationTemperature = 0.8

// BragService gera um BragDocument a partir do rascunho informal do
// usuário — equivalente a bragGeneratorFlow.
type BragService struct {
	Client GeminiClient
}

// Generate monta o prompt, chama o Gemini e retorna o BragDocument com um
// novo id gerado pelo servidor — equivalente a `{...output, id: uuidv4()}`
// no flow original.
func (s *BragService) Generate(ctx context.Context, definition string) (model.BragDocument, error) {
	prompt := promptTemplate + definition

	raw, err := s.Client.GenerateJSON(ctx, prompt, generationTemperature)
	if err != nil {
		return model.BragDocument{}, err
	}

	var parsed struct {
		Title            string   `json:"title"`
		Context          string   `json:"context"`
		ActionTaken      string   `json:"actionTaken"`
		BusinessImpact   string   `json:"businessImpact"`
		Metrics          []string `json:"metrics"`
		TechnologiesUsed []string `json:"technologiesUsed"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return model.BragDocument{}, fmt.Errorf("falha ao gerar o conteudo: nenhum output valido foi retornado pela IA: %w", err)
	}

	return model.BragDocument{
		ID:               uuid.New().String(),
		Title:            parsed.Title,
		Context:          parsed.Context,
		ActionTaken:      parsed.ActionTaken,
		BusinessImpact:   parsed.BusinessImpact,
		Metrics:          parsed.Metrics,
		TechnologiesUsed: parsed.TechnologiesUsed,
	}, nil
}
