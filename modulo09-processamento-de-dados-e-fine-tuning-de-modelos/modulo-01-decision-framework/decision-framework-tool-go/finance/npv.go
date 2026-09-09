// Package finance implementa NPV/DCF, Monte Carlo, Real Options (árvore
// binomial CRR) e análise de sensibilidade -- equivalente às partes 3-6 de
// decision-framework-tool.js.
package finance

import (
	"math"

	"decision-framework-tool/config"
)

// NpvParams são os parâmetros determinísticos de um cenário de NPV.
type NpvParams struct {
	VolumeInicialMensal      float64
	CrescimentoMensal        float64
	CustoPorChamadaStatusQuo float64
	CustoPorChamadaFineTuned float64
	CustoTreinamento         float64
	HorizonteMeses           int
	TaxaDescontoMensal       float64
	AtrasoMeses              int
}

// FluxoMensal é um ponto da série temporal de NPV.
type FluxoMensal struct {
	Mes               int
	Volume            float64
	EconomiaDescontada float64
	NpvAcumulado      float64
}

// NpvResultado é o resultado de CalcularNPV.
type NpvResultado struct {
	Npv          float64
	MesBreakeven *int
	Fluxos       []FluxoMensal
}

// ParamsDeterministicos converte o bloco `financeiro` num cenário determinístico
// usando o valor "moda" de cada distribuição triangular.
func ParamsDeterministicos(f *config.Financeiro) NpvParams {
	return NpvParams{
		VolumeInicialMensal:      f.VolumeInicialMensal,
		CrescimentoMensal:        f.CrescimentoMensal.Moda,
		CustoPorChamadaStatusQuo: f.CustoPorChamadaStatusQuo.Moda,
		CustoPorChamadaFineTuned: f.CustoPorChamadaFineTuned.Moda,
		CustoTreinamento:         f.CustoTreinamento,
		HorizonteMeses:           f.HorizonteMeses,
		TaxaDescontoMensal:       f.TaxaDescontoMensal,
	}
}

// CalcularNPV calcula o NPV de migrar pra um modelo fine-tunado: economia
// mensal (volume x diferença de custo por chamada), descontada mês a mês,
// menos o custo de treinar. Volume cresce geometricamente pela taxa informada.
func CalcularNPV(p NpvParams) NpvResultado {
	npv := -p.CustoTreinamento
	volume := p.VolumeInicialMensal
	fluxos := make([]FluxoMensal, 0, p.HorizonteMeses)
	var mesBreakeven *int

	for mes := 1; mes <= p.HorizonteMeses; mes++ {
		volume *= 1 + p.CrescimentoMensal
		economiaDescontada := 0.0
		// durante o atraso (esperando dado), não há economia: ainda se paga o status quo
		if mes > p.AtrasoMeses {
			economiaMes := volume * (p.CustoPorChamadaStatusQuo - p.CustoPorChamadaFineTuned)
			fatorDesconto := math.Pow(1+p.TaxaDescontoMensal, float64(mes))
			economiaDescontada = economiaMes / fatorDesconto
			npv += economiaDescontada
		}
		if mesBreakeven == nil && npv > 0 {
			m := mes
			mesBreakeven = &m
		}
		fluxos = append(fluxos, FluxoMensal{Mes: mes, Volume: volume, EconomiaDescontada: economiaDescontada, NpvAcumulado: npv})
	}

	return NpvResultado{Npv: roundMoney(npv), MesBreakeven: mesBreakeven, Fluxos: fluxos}
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundRatio(v float64) float64 {
	return math.Round(v*10000) / 10000
}
