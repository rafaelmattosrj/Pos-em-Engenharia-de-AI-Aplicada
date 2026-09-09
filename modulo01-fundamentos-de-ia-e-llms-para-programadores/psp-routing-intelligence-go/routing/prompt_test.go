package routing

import (
	"strings"
	"testing"

	"psp-routing-intelligence/domain"
)

func TestBuildPrompt_IncluiTransacaoECasosSimilares(t *testing.T) {
	tx := domain.Transaction{
		Amount: 1500.00, Method: domain.CreditCard, Brand: "VISA",
		UserRegion: "SP", Hour: 20, MerchantCategory: "STREAMING",
	}
	similarCases := []domain.SimilarCase{
		{PSP: domain.Adyen, Status: domain.Success, Amount: 1320.00, Method: domain.CreditCard, Similarity: 0.94},
		{PSP: domain.Bradesco, Status: domain.Failed, Amount: 1450.00, Method: domain.CreditCard, Similarity: 0.89},
	}

	prompt := BuildPrompt(tx, similarCases)

	for _, want := range []string{
		"Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA",
		"PSP=ADYEN", "status=SUCCESS",
		"PSP=BRADESCO", "status=FAILED",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt deveria conter %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "{transaction}") || strings.Contains(prompt, "{similarCases}") {
		t.Errorf("placeholders nao substituidos:\n%s", prompt)
	}
}

func TestBuildPrompt_MantemInstrucoesFixasDoTemplate(t *testing.T) {
	tx := domain.Transaction{Amount: 59.90, Method: domain.Pix, UserRegion: "BA", Hour: 10, MerchantCategory: "STREAMING"}

	prompt := BuildPrompt(tx, nil)

	if !strings.Contains(prompt, "ADYEN, BRASPAG, BRADESCO, SANTANDER, NUPAY, PICPAY, MERCADOPAGO") {
		t.Error("prompt deveria listar os PSPs disponiveis")
	}
	if !strings.Contains(prompt, "Responda EXCLUSIVAMENTE em JSON válido") {
		t.Error("prompt deveria manter a instrucao de formato de saida")
	}
	if !strings.Contains(prompt, "(nenhum caso historico similar encontrado)") {
		t.Error("prompt deveria indicar ausencia de casos similares")
	}
}
