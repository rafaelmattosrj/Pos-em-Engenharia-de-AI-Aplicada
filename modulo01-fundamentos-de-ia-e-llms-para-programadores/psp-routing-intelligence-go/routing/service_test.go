package routing

import (
	"context"
	"errors"
	"testing"

	"psp-routing-intelligence/domain"
)

func TestService_Recommend_CombinaRecomendacaoComOsCasosSimilares(t *testing.T) {
	tx := domain.Transaction{
		Amount: 1500.00, Method: domain.CreditCard, Brand: "VISA",
		UserRegion: "SP", Hour: 20, MerchantCategory: "STREAMING",
	}
	similarCases := []domain.SimilarCase{
		{PSP: domain.Adyen, Status: domain.Success, Amount: 1320.00, Method: domain.CreditCard, Similarity: 0.94},
	}

	var promptSeenByChat string
	service := &Service{
		Embed: func(ctx context.Context, text string) ([]float64, error) {
			if text != SerializeTransaction(tx) {
				t.Errorf("texto inesperado passado para Embed: %q", text)
			}
			return []float64{0.1, 0.2}, nil
		},
		FindSimilar: func(ctx context.Context, embedding []float64, topK int) ([]domain.SimilarCase, error) {
			if topK != TopK {
				t.Errorf("esperava topK=%d, obteve %d", TopK, topK)
			}
			return similarCases, nil
		},
		Chat: func(ctx context.Context, prompt string) (string, error) {
			promptSeenByChat = prompt
			return `{"primary":"ADYEN","confidence":0.87,"reasoning":"Alta aprovacao no Adyen.","fallback":["BRASPAG","BRADESCO"]}`, nil
		},
	}

	got, err := service.Recommend(context.Background(), tx)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if got.Primary != domain.Adyen {
		t.Errorf("esperava primary=ADYEN, obteve %s", got.Primary)
	}
	if got.Confidence != 0.87 {
		t.Errorf("esperava confidence=0.87, obteve %v", got.Confidence)
	}
	if len(got.Fallback) != 2 {
		t.Errorf("esperava 2 PSPs de fallback, obteve %d", len(got.Fallback))
	}
	if len(got.SimilarCases) != 1 || got.SimilarCases[0].PSP != domain.Adyen {
		t.Errorf("similarCases nao propagados corretamente: %v", got.SimilarCases)
	}
	if promptSeenByChat == "" {
		t.Error("esperava que o prompt fosse construido e passado ao Chat")
	}
}

func TestService_Recommend_FuncionaSemCasosSimilares(t *testing.T) {
	tx := domain.Transaction{Amount: 9.90, Method: domain.Wallet, Brand: "MERCADOPAGO", UserRegion: "RJ", Hour: 15, MerchantCategory: "NEWS"}

	service := &Service{
		Embed: func(ctx context.Context, text string) ([]float64, error) { return []float64{0.1}, nil },
		FindSimilar: func(ctx context.Context, embedding []float64, topK int) ([]domain.SimilarCase, error) {
			return nil, nil
		},
		Chat: func(ctx context.Context, prompt string) (string, error) {
			return `{"primary":"MERCADOPAGO","confidence":0.6,"reasoning":"Unico PSP compativel."}`, nil
		},
	}

	got, err := service.Recommend(context.Background(), tx)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got.SimilarCases) != 0 {
		t.Errorf("esperava similarCases vazio, obteve %v", got.SimilarCases)
	}
}

func TestService_Recommend_PropagaErroDeEmbedding(t *testing.T) {
	boom := errors.New("ollama indisponivel")
	service := &Service{
		Embed: func(ctx context.Context, text string) ([]float64, error) { return nil, boom },
	}

	_, err := service.Recommend(context.Background(), domain.Transaction{})
	if !errors.Is(err, boom) {
		t.Fatalf("esperava erro propagado de Embed, obteve %v", err)
	}
}

func TestService_Recommend_PropagaErroDeParsingDoLlm(t *testing.T) {
	service := &Service{
		Embed: func(ctx context.Context, text string) ([]float64, error) { return []float64{0.1}, nil },
		FindSimilar: func(ctx context.Context, embedding []float64, topK int) ([]domain.SimilarCase, error) {
			return nil, nil
		},
		Chat: func(ctx context.Context, prompt string) (string, error) { return "isso nao e json", nil },
	}

	_, err := service.Recommend(context.Background(), domain.Transaction{})

	var invalidErr *InvalidLLMResponseError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *InvalidLLMResponseError, obteve %v (%T)", err, err)
	}
}
