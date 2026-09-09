package routing

import (
	"context"
	"errors"
	"testing"

	"psp-routing-intelligence/domain"
)

func TestSeedService_Seed_CarregaAs50TransacoesDoJsonEmbutido(t *testing.T) {
	var cleared bool
	var stored int

	service := &SeedService{
		Clear: func(ctx context.Context) error { cleared = true; return nil },
		Embed: func(ctx context.Context, text string) ([]float64, error) { return []float64{0.1}, nil },
		Store: func(ctx context.Context, tx domain.HistoricalTransaction, text string, embedding []float64) error {
			stored++
			return nil
		},
	}

	n, err := service.Seed(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if n != 50 {
		t.Errorf("esperava 50 transacoes carregadas, obteve %d", n)
	}
	if !cleared {
		t.Error("esperava que Clear fosse chamado antes de popular o seed")
	}
	if stored != 50 {
		t.Errorf("esperava 50 chamadas de Store, obteve %d", stored)
	}
}

func TestSeedService_Seed_PropagaErroDeClear(t *testing.T) {
	boom := errors.New("neo4j indisponivel")
	service := &SeedService{
		Clear: func(ctx context.Context) error { return boom },
	}

	_, err := service.Seed(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("esperava erro propagado de Clear, obteve %v", err)
	}
}

func TestSeedService_Seed_PropagaErroDeEmbedding(t *testing.T) {
	boom := errors.New("ollama indisponivel")
	service := &SeedService{
		Clear: func(ctx context.Context) error { return nil },
		Embed: func(ctx context.Context, text string) ([]float64, error) { return nil, boom },
	}

	_, err := service.Seed(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("esperava erro propagado de Embed, obteve %v", err)
	}
}
