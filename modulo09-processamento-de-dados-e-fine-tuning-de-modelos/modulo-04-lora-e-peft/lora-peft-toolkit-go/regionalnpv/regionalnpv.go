// Package regionalnpv porta regional-lora-vs-cloud-npv.js (Modulo 4.1).
// Reabre CalcularNPV, do decision-framework-tool do Modulo 1.3 (aqui
// reimplementado -- este repositorio nao compartilha modulo Go entre pastas
// de disciplinas diferentes, entao a formula e replicada aqui e deve ser
// mantida em sincronia manualmente com o original), para responder: quando o
// caso ja foi aprovado mas o volume e pequeno demais pro custo fixo de GPU
// alugada, o que muda se o treino for local via LoRA?
package regionalnpv

import "math"

const (
	VolumeRegionalMensal = 400.0
	CrescimentoMensal    = 0.03
	CustoStatusQuo       = 0.045
	CustoFineTuned       = 0.016
	// [Estimativa de mercado, ago/2026] Numero ilustrativo herdado do case
	// Amplitude Seguros do Modulo 1.3, calibrado pra sustentar a historia
	// financeira do case, usando como referencia frouxa a faixa de GPU cloud
	// do cheatsheet do Modulo 1.3 (~US$0,40-4,00/hora). Confira precos atuais
	// antes de usar isto pra uma decisao real (vast.ai/pricing, runpod.io/pricing).
	CustoJobGerenciado    = 2400.0
	CustoLoraLocal        = 0.0
	HorizonteMeses        = 24
	TaxaDescontoMensal    = 0.01
	NumParceriasRegionais = 5
)

type Params struct {
	VolumeInicialMensal      float64
	CrescimentoMensal        float64
	CustoPorChamadaStatusQuo float64
	CustoPorChamadaFineTuned float64
	CustoTreinamento         float64
	HorizonteMeses           int
	TaxaDescontoMensal       float64
	AtrasoMeses              int
}

type Fluxo struct {
	Mes                int
	Volume             float64
	EconomiaDescontada float64
	NpvAcumulado       float64
}

type Resultado struct {
	Npv          float64
	MesBreakeven *int
	Fluxos       []Fluxo
}

func CalcularNPV(p Params) Resultado {
	npv := -p.CustoTreinamento
	volume := p.VolumeInicialMensal
	var fluxos []Fluxo
	var mesBreakeven *int

	for mes := 1; mes <= p.HorizonteMeses; mes++ {
		volume *= 1 + p.CrescimentoMensal
		economiaDescontada := 0.0
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
		fluxos = append(fluxos, Fluxo{Mes: mes, Volume: volume, EconomiaDescontada: economiaDescontada, NpvAcumulado: npv})
	}

	return Resultado{Npv: math.Round(npv*100) / 100, MesBreakeven: mesBreakeven, Fluxos: fluxos}
}

func ParamsBase(custoTreinamento float64) Params {
	return Params{
		VolumeInicialMensal:      VolumeRegionalMensal,
		CrescimentoMensal:        CrescimentoMensal,
		CustoPorChamadaStatusQuo: CustoStatusQuo,
		CustoPorChamadaFineTuned: CustoFineTuned,
		CustoTreinamento:         custoTreinamento,
		HorizonteMeses:           HorizonteMeses,
		TaxaDescontoMensal:       TaxaDescontoMensal,
	}
}
