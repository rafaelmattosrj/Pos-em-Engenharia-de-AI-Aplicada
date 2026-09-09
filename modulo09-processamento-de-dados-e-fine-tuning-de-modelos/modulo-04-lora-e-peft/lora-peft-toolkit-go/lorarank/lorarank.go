// Package lorarank porta lora-rank-tradeoff-tool.js (Modulo 4.3). Dados reais
// de tres treinos LoRA (rank 4/8/16) mais comparacoes de quantizacao (bf16 vs.
// 4-bit / QLoRA) e tipo de adaptacao (LoRA vs. DoRA).
package lorarank

import "math"

type Execucao struct {
	Rank                 int
	ParametrosTreinaveis float64
	PercentualModelo     float64
	ValLossInicial       float64
	ValLossFinal         float64
	PicoMemGB            float64
	ItPorSegundoFinal    float64
	TamanhoAdapterMB     int
}

var ExecucoesReais = []Execucao{
	{4, 3.408e6, 0.074, 4.752, 1.246, 10.787, 7.401, 13},
	{8, 6.816e6, 0.147, 4.752, 0.895, 10.833, 7.295, 27},
	{16, 13.631e6, 0.295, 4.752, 0.725, 10.930, 7.253, 52},
}

type ConfigQuantizacao struct {
	TamanhoModeloDiscoGB float64
	PicoMemTreinoGB      float64
	ValLossFinal         float64
	TamanhoAdapterMB     int
}

var Bf16 = ConfigQuantizacao{10.241, 10.833, 0.895, 27}
var QuatroBit = ConfigQuantizacao{3.583, 4.193, 0.932, 27}

type ConfigAdaptacao struct {
	ParametrosTreinaveis float64
	PicoMemGB            float64
	ValLossFinal         float64
	TamanhoAdapterMB     int
}

var Lora = ConfigAdaptacao{6.816e6, 10.833, 0.895, 27}
var Dora = ConfigAdaptacao{7.328e6, 11.099, 0.895, 28}

func round(v float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(v*fator) / fator
}

func CalcularReducaoValLoss(e Execucao) float64 {
	return round(((e.ValLossInicial-e.ValLossFinal)/e.ValLossInicial)*100, 2)
}

type ComparacaoSucessiva struct {
	DeRank              int
	ParaRank            int
	RazaoParametros     float64
	MelhoriaValLoss     float64
	MelhoriaPercentual  float64
	CustoMemoriaExtraGB float64
}

func CompararExecucoesSucessivas(execucoes []Execucao) []ComparacaoSucessiva {
	comparacoes := make([]ComparacaoSucessiva, 0, len(execucoes)-1)
	for i := 1; i < len(execucoes); i++ {
		anterior := execucoes[i-1]
		atual := execucoes[i]
		razaoParametros := round(atual.ParametrosTreinaveis/anterior.ParametrosTreinaveis, 2)
		melhoriaValLoss := round(anterior.ValLossFinal-atual.ValLossFinal, 3)
		melhoriaPercentual := round((melhoriaValLoss/anterior.ValLossFinal)*100, 2)
		custoMemoriaExtraGB := round(atual.PicoMemGB-anterior.PicoMemGB, 3)
		comparacoes = append(comparacoes, ComparacaoSucessiva{
			DeRank: anterior.Rank, ParaRank: atual.Rank, RazaoParametros: razaoParametros,
			MelhoriaValLoss: melhoriaValLoss, MelhoriaPercentual: melhoriaPercentual,
			CustoMemoriaExtraGB: custoMemoriaExtraGB,
		})
	}
	return comparacoes
}

type ResultadoQuantizacao struct {
	ReducaoDiscoPct     float64
	ReducaoMemTreinoPct float64
	CustoValLoss        float64
	CustoValLossPct     float64
}

func CompararQuantizacao(bf16, quatroBit ConfigQuantizacao) ResultadoQuantizacao {
	reducaoDiscoPct := round((1-quatroBit.TamanhoModeloDiscoGB/bf16.TamanhoModeloDiscoGB)*100, 1)
	reducaoMemTreinoPct := round((1-quatroBit.PicoMemTreinoGB/bf16.PicoMemTreinoGB)*100, 1)
	custoValLoss := round(quatroBit.ValLossFinal-bf16.ValLossFinal, 3)
	custoValLossPct := round((custoValLoss/bf16.ValLossFinal)*100, 1)
	return ResultadoQuantizacao{reducaoDiscoPct, reducaoMemTreinoPct, custoValLoss, custoValLossPct}
}

type ResultadoAdaptacao struct {
	RazaoParametros     float64
	CustoMemoriaExtraGB float64
	DiferencaValLoss    float64
}

func CompararTipoAdaptacao(lora, dora ConfigAdaptacao) ResultadoAdaptacao {
	razaoParametros := round(dora.ParametrosTreinaveis/lora.ParametrosTreinaveis, 3)
	custoMemoriaExtraGB := round(dora.PicoMemGB-lora.PicoMemGB, 3)
	diferencaValLoss := round(dora.ValLossFinal-lora.ValLossFinal, 3)
	return ResultadoAdaptacao{razaoParametros, custoMemoriaExtraGB, diferencaValLoss}
}

func RecomendarRankMinimo(execucoes []Execucao, margemAceitavel float64) Execucao {
	melhorValLoss := math.Inf(1)
	for _, e := range execucoes {
		if e.ValLossFinal < melhorValLoss {
			melhorValLoss = e.ValLossFinal
		}
	}
	limiar := melhorValLoss * (1 + margemAceitavel)

	var recomendado Execucao
	primeiro := true
	for _, e := range execucoes {
		if e.ValLossFinal <= limiar {
			if primeiro || e.Rank < recomendado.Rank {
				recomendado = e
				primeiro = false
			}
		}
	}
	return recomendado
}
