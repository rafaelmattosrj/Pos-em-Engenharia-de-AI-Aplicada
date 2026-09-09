package finance

import (
	"math"
	"math/rand"
	"testing"

	"decision-framework-tool/config"
	"decision-framework-tool/framework"
)

func carregarSaude(t *testing.T) (*config.AmplitudeConfig, config.Caso) {
	t.Helper()
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	saude, err := cfg.Caso("amplitude-saude-empresarial")
	if err != nil {
		t.Fatal(err)
	}
	return cfg, saude
}

func TestEsperarTemValorPositivoQuandoCustoDeErroEAlto(t *testing.T) {
	cfg, saude := carregarSaude(t)
	opcao := PrecificarOpcaoDeEsperar(saude.Financeiro, saude.Scores["p3"], cfg.LimiarVerde, 2000, rand.Float64)
	if opcao.MesesParaEsperar <= 0 {
		t.Fatalf("esperado mesesParaEsperar > 0, obtido %d", opcao.MesesParaEsperar)
	}
	if opcao.Recomendacao != framework.Esperar {
		t.Fatalf("esperado recomendacao Esperar, obtido %v", opcao.Recomendacao)
	}
	if opcao.ValorDeEsperar <= 0 {
		t.Fatalf("esperado valorDeEsperar > 0, obtido %v", opcao.ValorDeEsperar)
	}
}

func TestVolatilidadePositivaEFatoresUDConsistentes(t *testing.T) {
	cfg, saude := carregarSaude(t)
	opcao := PrecificarOpcaoDeEsperar(saude.Financeiro, saude.Scores["p3"], cfg.LimiarVerde, 2000, rand.Float64)
	if opcao.Sigma <= 0 {
		t.Fatalf("esperado sigma > 0, obtido %v", opcao.Sigma)
	}
	if math.Abs(opcao.U*opcao.D-1.0) > 1e-3 {
		t.Fatalf("esperado u*d~=1, obtido %v", opcao.U*opcao.D)
	}
	if opcao.ProbabilidadeRiscoNeutra < 0 || opcao.ProbabilidadeRiscoNeutra > 1 {
		t.Fatalf("esperado probabilidadeRiscoNeutra em [0,1], obtido %v", opcao.ProbabilidadeRiscoNeutra)
	}
}

func TestEsperarVemComReavaliacaoAgendada(t *testing.T) {
	cfg, saude := carregarSaude(t)
	opcao := PrecificarOpcaoDeEsperar(saude.Financeiro, saude.Scores["p3"], cfg.LimiarVerde, 2000, rand.Float64)
	if opcao.ReavaliacaoAgendadaEm != "Módulo 3.2" {
		t.Fatalf("esperado 'Módulo 3.2', obtido %s", opcao.ReavaliacaoAgendadaEm)
	}
}
