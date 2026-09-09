package finance

import (
	"math"
	"testing"
)

func TestSemCrescimentoBateComFormulaFechadaDeAnuidade(t *testing.T) {
	params := NpvParams{
		VolumeInicialMensal: 1000, CrescimentoMensal: 0,
		CustoPorChamadaStatusQuo: 0.05, CustoPorChamadaFineTuned: 0.02,
		CustoTreinamento: 1000, HorizonteMeses: 12, TaxaDescontoMensal: 0.01,
	}
	resultado := CalcularNPV(params)

	economiaPorChamada := 0.05 - 0.02
	fluxoMensal := 1000 * economiaPorChamada
	pvAnuidade := fluxoMensal * (1 - math.Pow(1.01, -12)) / 0.01
	npvEsperado := -1000 + pvAnuidade

	if math.Abs(resultado.Npv-npvEsperado) > 0.5 {
		t.Fatalf("esperado npv~=%v, obtido %v", npvEsperado, resultado.Npv)
	}
}

func TestComAtrasoNaoGeraEconomiaDuranteEspera(t *testing.T) {
	params := NpvParams{
		VolumeInicialMensal: 1000, CrescimentoMensal: 0,
		CustoPorChamadaStatusQuo: 0.05, CustoPorChamadaFineTuned: 0.02,
		CustoTreinamento: 0, HorizonteMeses: 3, TaxaDescontoMensal: 0, AtrasoMeses: 3,
	}
	resultado := CalcularNPV(params)
	if resultado.Npv != 0.0 {
		t.Fatalf("esperado npv=0, obtido %v", resultado.Npv)
	}
}
