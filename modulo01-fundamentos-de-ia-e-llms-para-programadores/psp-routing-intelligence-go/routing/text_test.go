package routing

import (
	"strings"
	"testing"

	"psp-routing-intelligence/domain"
)

func TestSerializeTransaction_ComBandeiraIncluiClausulaDeBandeira(t *testing.T) {
	tx := domain.Transaction{
		Amount: 1500.00, Method: domain.CreditCard, Brand: "VISA",
		UserRegion: "SP", Hour: 20, MerchantCategory: "STREAMING",
	}

	got := SerializeTransaction(tx)
	want := "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA, regiao SP, hora 20h, categoria STREAMING."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSerializeTransaction_SemBandeiraOmiteClausula(t *testing.T) {
	tx := domain.Transaction{
		Amount: 59.90, Method: domain.Pix, Brand: "",
		UserRegion: "BA", Hour: 10, MerchantCategory: "STREAMING",
	}

	got := SerializeTransaction(tx)
	want := "Pagamento via PIX, valor R$59.90, regiao BA, hora 10h, categoria STREAMING."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if strings.Contains(got, "bandeira") {
		t.Errorf("nao esperava clausula de bandeira em %q", got)
	}
}

func TestSerializeHistorical_UsaOsMesmosCamposDeTransaction(t *testing.T) {
	tx := domain.HistoricalTransaction{
		Amount: 1320.00, Method: domain.CreditCard, Brand: "VISA", Region: "SP",
		Hour: 22, Category: "STREAMING", PSP: domain.Adyen, Status: domain.Success,
	}

	got := SerializeHistorical(tx)
	want := "Pagamento via CREDIT_CARD, valor R$1320.00, bandeira VISA, regiao SP, hora 22h, categoria STREAMING."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
