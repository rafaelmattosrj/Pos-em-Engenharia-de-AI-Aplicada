package finance

import (
	"math"
	"sort"

	"decision-framework-tool/config"
	"decision-framework-tool/framework"
)

// RealOptionsResultado é o resultado de PrecificarOpcaoDeEsperar.
type RealOptionsResultado struct {
	MesesParaEsperar          int
	Sigma                     float64
	U, D                      float64
	ProbabilidadeRiscoNeutra  float64
	ValorPresenteFluxosBrutos float64
	ValorExercerAgora         float64
	ValorOpcaoEsperar         float64
	ValorDeEsperar            float64
	Recomendacao              framework.Recomendacao
	ReavaliacaoAgendadaEm     string
}

// CalcularValorPresenteFluxos é o valor presente BRUTO dos fluxos de economia,
// sem subtrair o custo de treino.
func CalcularValorPresenteFluxos(p NpvParams) float64 {
	return CalcularNPV(p).Npv + p.CustoTreinamento
}

// DerivarVolatilidade deriva a volatilidade mensal do valor presente bruto a
// partir do desvio padrão do log-retorno de cada simulação em relação à
// mediana -- usa a mesma simulação de Monte Carlo de SimularMonteCarlo.
func DerivarVolatilidade(f *config.Financeiro, n int, rng func() float64) float64 {
	valores := make([]float64, n)
	for i := 0; i < n; i++ {
		crescimento := AmostrarTriangular(f.CrescimentoMensal.Min, f.CrescimentoMensal.Moda, f.CrescimentoMensal.Max, rng)
		custoStatusQuo := AmostrarTriangular(f.CustoPorChamadaStatusQuo.Min, f.CustoPorChamadaStatusQuo.Moda, f.CustoPorChamadaStatusQuo.Max, rng)
		custoFineTuned := AmostrarTriangular(f.CustoPorChamadaFineTuned.Min, f.CustoPorChamadaFineTuned.Moda, f.CustoPorChamadaFineTuned.Max, rng)
		valores[i] = CalcularValorPresenteFluxos(NpvParams{
			VolumeInicialMensal:      f.VolumeInicialMensal,
			CrescimentoMensal:        crescimento,
			CustoPorChamadaStatusQuo: custoStatusQuo,
			CustoPorChamadaFineTuned: custoFineTuned,
			CustoTreinamento:         f.CustoTreinamento,
			HorizonteMeses:           f.HorizonteMeses,
			TaxaDescontoMensal:       f.TaxaDescontoMensal,
		})
	}
	sort.Float64s(valores)
	mediana := Percentil(valores, 0.5)

	somaLog, qtd := 0.0, 0
	for _, v := range valores {
		if v > 0 && mediana > 0 {
			somaLog += math.Log(v / mediana)
			qtd++
		}
	}
	media := somaLog / float64(qtd)

	variancia := 0.0
	for _, v := range valores {
		if v > 0 && mediana > 0 {
			logRetorno := math.Log(v / mediana)
			variancia += math.Pow(logRetorno-media, 2)
		}
	}
	variancia /= float64(qtd - 1)
	return math.Sqrt(variancia)
}

// PrecificarOpcaoDeEsperar precifica o valor de esperar como opção real
// (árvore binomial CRR, opção americana com única janela de exercício,
// indução retroativa). Só se aplica quando a única reprovação é dado
// insuficiente (FalhaSoDado).
func PrecificarOpcaoDeEsperar(f *config.Financeiro, scoreAtualP3, limiarVerde float64, n int, rng func() float64) RealOptionsResultado {
	opcaoReal := f.OpcaoReal
	mesesParaEsperar := int(math.Ceil((opcaoReal.ScoreAlvo - scoreAtualP3) / opcaoReal.TaxaCrescimentoScorePorMes))
	custoTreinamento := f.CustoTreinamento

	s0 := CalcularValorPresenteFluxos(NpvParams{
		VolumeInicialMensal:      f.VolumeInicialMensal,
		CrescimentoMensal:        f.CrescimentoMensal.Moda,
		CustoPorChamadaStatusQuo: f.CustoPorChamadaStatusQuo.Moda,
		CustoPorChamadaFineTuned: f.CustoPorChamadaFineTuned.Moda,
		CustoTreinamento:         custoTreinamento,
		HorizonteMeses:           f.HorizonteMeses,
		TaxaDescontoMensal:       f.TaxaDescontoMensal,
	})

	sigma := DerivarVolatilidade(f, n, rng)
	deltaT := 1.0
	r := f.TaxaDescontoMensal
	u := math.Exp(sigma * math.Sqrt(deltaT))
	d := 1 / u
	p := (math.Exp(r*deltaT) - d) / (u - d)

	valoresOpcao := make([]float64, mesesParaEsperar+1)
	for j := 0; j <= mesesParaEsperar; j++ {
		sFinal := s0 * math.Pow(u, float64(mesesParaEsperar-j)) * math.Pow(d, float64(j))
		valoresOpcao[j] = math.Max(sFinal-custoTreinamento, 0)
	}
	for passo := mesesParaEsperar; passo > 0; passo-- {
		proximoPasso := make([]float64, passo)
		for j := 0; j < passo; j++ {
			valorEsperado := p*valoresOpcao[j] + (1-p)*valoresOpcao[j+1]
			proximoPasso[j] = valorEsperado / math.Exp(r*deltaT)
		}
		valoresOpcao = proximoPasso
	}
	valorOpcaoEsperar := roundMoney(valoresOpcao[0])

	sAgora := CalcularValorPresenteFluxos(NpvParams{
		VolumeInicialMensal:      f.VolumeInicialMensal,
		CrescimentoMensal:        f.CrescimentoMensal.Moda,
		CustoPorChamadaStatusQuo: f.CustoPorChamadaStatusQuo.Moda,
		CustoPorChamadaFineTuned: f.CustoPorChamadaFineTuned.Moda + opcaoReal.CustoDeErroEsperadoPorChamada,
		CustoTreinamento:         custoTreinamento,
		HorizonteMeses:           f.HorizonteMeses,
		TaxaDescontoMensal:       f.TaxaDescontoMensal,
	})
	valorExercerAgora := roundMoney(math.Max(sAgora-custoTreinamento, 0))

	valorDeEsperar := roundMoney(valorOpcaoEsperar - valorExercerAgora)

	recomendacao := framework.FineTuning
	if valorDeEsperar > 0 {
		recomendacao = framework.Esperar
	}

	return RealOptionsResultado{
		MesesParaEsperar:          mesesParaEsperar,
		Sigma:                     roundRatio(sigma),
		U:                         roundRatio(u),
		D:                         roundRatio(d),
		ProbabilidadeRiscoNeutra:  roundRatio(p),
		ValorPresenteFluxosBrutos: roundMoney(s0),
		ValorExercerAgora:         valorExercerAgora,
		ValorOpcaoEsperar:         valorOpcaoEsperar,
		ValorDeEsperar:            valorDeEsperar,
		Recomendacao:              recomendacao,
		ReavaliacaoAgendadaEm:     "Módulo 3.2",
	}
}
