package routing

import (
	"errors"
	"testing"

	"psp-routing-intelligence/domain"
)

func TestParseRecommendation_JsonValido(t *testing.T) {
	raw := `{
		"primary": "ADYEN",
		"confidence": 0.87,
		"reasoning": "Transacoes VISA em SP tem alta aprovacao no Adyen.",
		"fallback": ["BRASPAG", "BRADESCO"]
	}`

	got, err := ParseRecommendation(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.Primary != domain.Adyen {
		t.Errorf("esperava ADYEN, obteve %s", got.Primary)
	}
	if got.Confidence != 0.87 {
		t.Errorf("esperava 0.87, obteve %v", got.Confidence)
	}
	if len(got.Fallback) != 2 || got.Fallback[0] != domain.Braspag || got.Fallback[1] != domain.Bradesco {
		t.Errorf("fallback inesperado: %v", got.Fallback)
	}
}

func TestParseRecommendation_RemoveCercaMarkdown(t *testing.T) {
	raw := "```json\n{\"primary\": \"PICPAY\", \"confidence\": 0.95, \"reasoning\": \"Carteira PicPay.\", \"fallback\": []}\n```"

	got, err := ParseRecommendation(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.Primary != domain.Picpay {
		t.Errorf("esperava PICPAY, obteve %s", got.Primary)
	}
	if len(got.Fallback) != 0 {
		t.Errorf("esperava fallback vazio, obteve %v", got.Fallback)
	}
}

func TestParseRecommendation_SemFallbackRetornaListaVazia(t *testing.T) {
	raw := `{"primary": "BRASPAG", "confidence": 0.7, "reasoning": "PIX."}`

	got, err := ParseRecommendation(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got.Fallback) != 0 {
		t.Errorf("esperava fallback vazio, obteve %v", got.Fallback)
	}
}

func TestParseRecommendation_JsonInvalidoRetornaInvalidLLMResponseError(t *testing.T) {
	_, err := ParseRecommendation("isso nao e json")

	var invalidErr *InvalidLLMResponseError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *InvalidLLMResponseError, obteve %v (%T)", err, err)
	}
}

func TestParseRecommendation_SemCamposObrigatoriosRetornaErro(t *testing.T) {
	_, err := ParseRecommendation(`{"reasoning": "faltam campos"}`)

	var invalidErr *InvalidLLMResponseError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *InvalidLLMResponseError, obteve %v (%T)", err, err)
	}
}

func TestParseRecommendation_PspDesconhecidoRetornaErro(t *testing.T) {
	raw := `{"primary": "PAYPAL", "confidence": 0.5, "reasoning": "PSP fora do dominio."}`

	_, err := ParseRecommendation(raw)

	var invalidErr *InvalidLLMResponseError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *InvalidLLMResponseError, obteve %v (%T)", err, err)
	}
}

func TestParseRecommendation_PspDesconhecidoNoFallbackRetornaErro(t *testing.T) {
	raw := `{"primary": "ADYEN", "confidence": 0.5, "reasoning": "ok", "fallback": ["PAYPAL"]}`

	_, err := ParseRecommendation(raw)

	var invalidErr *InvalidLLMResponseError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperava *InvalidLLMResponseError, obteve %v (%T)", err, err)
	}
}
