// Package reavaliacao e' o callback real pros Modulos 1.2/1.3 (Modulo 3.2):
// reabre o mesmo gate de decisao (decisionframework) com o caso Saude
// Empresarial atualizado pros 9 meses que se passaram, usando a mesma taxa
// de crescimento de score ja projetada no caso original -- nao inventa
// numero novo.
package reavaliacao

import (
	"math"

	"fine-tuning-via-api/decisionframework"
)

const MesesDecorridos = 9

func ConstruirCasoNoveMesesDepois(casoOriginal decisionframework.Caso) decisionframework.Caso {
	opcaoReal := casoOriginal.Financeiro.OpcaoReal
	p3Atualizado := math.Round((casoOriginal.Scores["p3"]+opcaoReal.TaxaCrescimentoScorePorMes*MesesDecorridos)*100) / 100

	scoresAtualizados := make(map[string]float64, len(casoOriginal.Scores))
	for k, v := range casoOriginal.Scores {
		scoresAtualizados[k] = v
	}
	scoresAtualizados["p3"] = p3Atualizado

	fatorCrescimentoVolume := math.Pow(1+casoOriginal.Financeiro.CrescimentoMensalModa, MesesDecorridos)
	volumeAtualizado := int(math.Round(float64(casoOriginal.Financeiro.VolumeInicialMensal) * fatorCrescimentoVolume))

	financeiroAtualizado := &decisionframework.Financeiro{
		VolumeInicialMensal:   volumeAtualizado,
		CrescimentoMensalModa: casoOriginal.Financeiro.CrescimentoMensalModa,
		OpcaoReal:             opcaoReal,
	}

	return decisionframework.Caso{ID: casoOriginal.ID, Scores: scoresAtualizados, Financeiro: financeiroAtualizado}
}
