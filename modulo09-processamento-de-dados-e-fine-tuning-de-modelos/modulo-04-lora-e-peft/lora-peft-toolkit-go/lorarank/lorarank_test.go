package lorarank

import (
	"math"
	"testing"
)

func TestRank8ReduzValLossEmCerca81PorCento(t *testing.T) {
	r := CalcularReducaoValLoss(ExecucoesReais[1])
	if math.Abs(r-81.17) >= 0.5 {
		t.Fatalf("reducao=%v, esperado ~81.17", r)
	}
}

func TestRank4Para8DobraParametros(t *testing.T) {
	c := CompararExecucoesSucessivas(ExecucoesReais)
	if c[0].RazaoParametros != 2.0 {
		t.Fatalf("razao=%v, esperado 2.0", c[0].RazaoParametros)
	}
}

func TestRank8Para16DobraParametros(t *testing.T) {
	c := CompararExecucoesSucessivas(ExecucoesReais)
	if c[1].RazaoParametros != 2.0 {
		t.Fatalf("razao=%v, esperado 2.0", c[1].RazaoParametros)
	}
}

func TestMaisRankSempreMelhoraValLoss(t *testing.T) {
	for _, c := range CompararExecucoesSucessivas(ExecucoesReais) {
		if c.MelhoriaValLoss <= 0 {
			t.Fatalf("rank %d->%d: melhoria=%v, esperado > 0", c.DeRank, c.ParaRank, c.MelhoriaValLoss)
		}
	}
}

func TestCustoMemoriaExtraPequeno(t *testing.T) {
	for _, c := range CompararExecucoesSucessivas(ExecucoesReais) {
		if c.CustoMemoriaExtraGB >= 0.2 {
			t.Fatalf("custo=%vGB, esperado < 0.2", c.CustoMemoriaExtraGB)
		}
	}
}

func TestRecomendaRank16ComMargem10PorCento(t *testing.T) {
	r := RecomendarRankMinimo(ExecucoesReais, 0.10)
	if r.Rank != 16 {
		t.Fatalf("rank=%d, esperado 16", r.Rank)
	}
}

func TestRecomendaRank8ComMargem60PorCento(t *testing.T) {
	r := RecomendarRankMinimo(ExecucoesReais, 0.60)
	if r.Rank != 8 {
		t.Fatalf("rank=%d, esperado 8", r.Rank)
	}
}

func TestQuatroBitReduzDiscoEmCerca65PorCento(t *testing.T) {
	r := CompararQuantizacao(Bf16, QuatroBit)
	if math.Abs(r.ReducaoDiscoPct-65.0) >= 1.0 {
		t.Fatalf("reducaoDiscoPct=%v, esperado ~65.0", r.ReducaoDiscoPct)
	}
}

func TestQuatroBitReduzMemTreinoEmCerca61PorCento(t *testing.T) {
	r := CompararQuantizacao(Bf16, QuatroBit)
	if math.Abs(r.ReducaoMemTreinoPct-61.3) >= 1.0 {
		t.Fatalf("reducaoMemTreinoPct=%v, esperado ~61.3", r.ReducaoMemTreinoPct)
	}
}

func TestCustoValLossQuantizacaoPequeno(t *testing.T) {
	r := CompararQuantizacao(Bf16, QuatroBit)
	if r.CustoValLossPct >= 10 {
		t.Fatalf("custoValLossPct=%v, esperado < 10", r.CustoValLossPct)
	}
}

func TestDoraUsaCerca7Ponto5PorCentoMaisParametros(t *testing.T) {
	r := CompararTipoAdaptacao(Lora, Dora)
	if math.Abs(r.RazaoParametros-1.075) >= 0.02 {
		t.Fatalf("razaoParametros=%v, esperado ~1.075", r.RazaoParametros)
	}
}

func TestDoraCustaMemoriaExtraEntre0Ponto2E0Ponto35(t *testing.T) {
	r := CompararTipoAdaptacao(Lora, Dora)
	if r.CustoMemoriaExtraGB <= 0.2 || r.CustoMemoriaExtraGB >= 0.35 {
		t.Fatalf("custoMemoriaExtraGB=%v, esperado entre 0.2 e 0.35", r.CustoMemoriaExtraGB)
	}
}

func TestDoraEmpataComLoraEmValLoss(t *testing.T) {
	r := CompararTipoAdaptacao(Lora, Dora)
	if r.DiferencaValLoss != 0 {
		t.Fatalf("diferencaValLoss=%v, esperado 0", r.DiferencaValLoss)
	}
}
