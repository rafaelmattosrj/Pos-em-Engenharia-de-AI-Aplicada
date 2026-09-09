package routing

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"psp-routing-intelligence/domain"
)

// seedData e data/transactions-seed.json embutido no binario — as mesmas 50
// transacoes ficticias usadas na versao Java (copiadas 1:1 para manter os
// dois portes consistentes).
//
//go:embed seed_transactions.json
var seedData []byte

// SeedService carrega as transacoes historicas e popula o Neo4j —
// equivalente a com.psprouting.application.SeedService, executado por
// POST /api/routing/seed.
type SeedService struct {
	// Clear remove as transacoes historicas armazenadas anteriormente.
	Clear func(ctx context.Context) error
	// Embed gera o vetor de embedding de um texto.
	Embed func(ctx context.Context, text string) ([]float64, error)
	// Store persiste uma transacao historica com seu texto e embedding.
	Store func(ctx context.Context, tx domain.HistoricalTransaction, text string, embedding []float64) error
}

// Seed limpa os dados anteriores e (re)popula o Neo4j com as transacoes de
// data/transactions-seed.json, retornando quantas foram carregadas.
func (s *SeedService) Seed(ctx context.Context) (int, error) {
	var transactions []domain.HistoricalTransaction
	if err := json.Unmarshal(seedData, &transactions); err != nil {
		return 0, fmt.Errorf("carregando transactions-seed.json: %w", err)
	}

	if err := s.Clear(ctx); err != nil {
		return 0, fmt.Errorf("limpando dados anteriores no Neo4j: %w", err)
	}

	for _, tx := range transactions {
		text := SerializeHistorical(tx)
		embedding, err := s.Embed(ctx, text)
		if err != nil {
			return 0, fmt.Errorf("gerando embedding do seed: %w", err)
		}
		if err := s.Store(ctx, tx, text, embedding); err != nil {
			return 0, fmt.Errorf("armazenando transacao do seed: %w", err)
		}
	}

	return len(transactions), nil
}
