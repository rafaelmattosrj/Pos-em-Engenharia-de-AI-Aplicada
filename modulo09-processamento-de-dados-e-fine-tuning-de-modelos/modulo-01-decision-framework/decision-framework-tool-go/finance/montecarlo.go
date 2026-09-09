package finance

import (
	"math"
	"sort"

	"decision-framework-tool/config"
)

// MonteCarloResultado é a distribuição de NPVs simulados.
type MonteCarloResultado struct {
	Media                float64
	P5, P50, P95         float64
	ProbabilidadePositivo float64
	N                    int
}

// AmostrarTriangular amostra de uma distribuição triangular via inversão de CDF.
func AmostrarTriangular(min, moda, max float64, rng func() float64) float64 {
	u := rng()
	f := (moda - min) / (max - min)
	if u < f {
		return min + math.Sqrt(u*(max-min)*(moda-min))
	}
	return max - math.Sqrt((1-u)*(max-min)*(max-moda))
}

// Percentil retorna o valor no percentil p (0-1) de um slice já ordenado.
func Percentil(valoresOrdenados []float64, p float64) float64 {
	indice := int(p * float64(len(valoresOrdenados)))
	if indice >= len(valoresOrdenados) {
		indice = len(valoresOrdenados) - 1
	}
	return valoresOrdenados[indice]
}

// SimularMonteCarlo roda n simulações de NPV amostrando crescimento e custo
// por chamada de distribuições triangulares.
func SimularMonteCarlo(f *config.Financeiro, n int, rng func() float64) MonteCarloResultado {
	resultados := make([]float64, n)
	for i := 0; i < n; i++ {
		crescimento := AmostrarTriangular(f.CrescimentoMensal.Min, f.CrescimentoMensal.Moda, f.CrescimentoMensal.Max, rng)
		custoStatusQuo := AmostrarTriangular(f.CustoPorChamadaStatusQuo.Min, f.CustoPorChamadaStatusQuo.Moda, f.CustoPorChamadaStatusQuo.Max, rng)
		custoFineTuned := AmostrarTriangular(f.CustoPorChamadaFineTuned.Min, f.CustoPorChamadaFineTuned.Moda, f.CustoPorChamadaFineTuned.Max, rng)

		params := NpvParams{
			VolumeInicialMensal:      f.VolumeInicialMensal,
			CrescimentoMensal:        crescimento,
			CustoPorChamadaStatusQuo: custoStatusQuo,
			CustoPorChamadaFineTuned: custoFineTuned,
			CustoTreinamento:         f.CustoTreinamento,
			HorizonteMeses:           f.HorizonteMeses,
			TaxaDescontoMensal:       f.TaxaDescontoMensal,
		}
		resultados[i] = CalcularNPV(params).Npv
	}
	sort.Float64s(resultados)

	soma := 0.0
	positivos := 0
	for _, v := range resultados {
		soma += v
		if v > 0 {
			positivos++
		}
	}
	media := soma / float64(n)

	return MonteCarloResultado{
		Media:                 roundMoney(media),
		P5:                    roundMoney(Percentil(resultados, 0.05)),
		P50:                   roundMoney(Percentil(resultados, 0.50)),
		P95:                   roundMoney(Percentil(resultados, 0.95)),
		ProbabilidadePositivo: roundRatio(float64(positivos) / float64(n)),
		N:                     n,
	}
}
