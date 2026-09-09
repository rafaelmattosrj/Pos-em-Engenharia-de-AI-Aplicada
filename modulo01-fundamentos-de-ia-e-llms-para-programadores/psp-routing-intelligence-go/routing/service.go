package routing

import (
	"context"

	"psp-routing-intelligence/domain"
)

// TopK e o numero de casos similares recuperados por consulta — "5
// transacoes" no IDEIA.md (mesma constante de
// SimilaritySearchService.TOP_K na versao Java).
const TopK = 5

// Service orquestra o pipeline completo descrito em IDEIA.md para
// POST /api/routing/recommend — equivalente a
// com.psprouting.application.RoutingService, com as dependencias externas
// (embeddings, Neo4j, OpenRouter) recebidas como campos de funcao em vez de
// beans injetados, para ficar testavel sem nenhuma delas.
type Service struct {
	// Embed gera o vetor de embedding de um texto.
	Embed func(ctx context.Context, text string) ([]float64, error)
	// FindSimilar busca os TopK casos historicos mais similares no Neo4j.
	FindSimilar func(ctx context.Context, embedding []float64, topK int) ([]domain.SimilarCase, error)
	// Chat envia o prompt RAG ao LLM e retorna a resposta em texto.
	Chat func(ctx context.Context, prompt string) (string, error)
}

// Recommend executa os 4 passos do pipeline:
//
//	(1) serializa a transacao e gera o embedding
//	(2) busca os TopK casos historicos mais similares no Neo4j
//	(3) monta o prompt RAG e chama o LLM via OpenRouter
//	(4) combina a recomendacao do LLM com os casos similares encontrados
func (s *Service) Recommend(ctx context.Context, tx domain.Transaction) (domain.RoutingRecommendation, error) {
	text := SerializeTransaction(tx)

	embedding, err := s.Embed(ctx, text)
	if err != nil {
		return domain.RoutingRecommendation{}, err
	}

	similarCases, err := s.FindSimilar(ctx, embedding, TopK)
	if err != nil {
		return domain.RoutingRecommendation{}, err
	}

	prompt := BuildPrompt(tx, similarCases)
	rawResponse, err := s.Chat(ctx, prompt)
	if err != nil {
		return domain.RoutingRecommendation{}, err
	}

	parsed, err := ParseRecommendation(rawResponse)
	if err != nil {
		return domain.RoutingRecommendation{}, err
	}

	return domain.RoutingRecommendation{
		Primary:      parsed.Primary,
		Confidence:   parsed.Confidence,
		Reasoning:    parsed.Reasoning,
		Fallback:     parsed.Fallback,
		SimilarCases: similarCases,
	}, nil
}
