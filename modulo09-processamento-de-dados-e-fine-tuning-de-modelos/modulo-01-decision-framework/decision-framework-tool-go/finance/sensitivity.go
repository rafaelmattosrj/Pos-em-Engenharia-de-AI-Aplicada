package finance

import (
	"sort"

	"decision-framework-tool/config"
)

// SensibilidadeResultado é o impacto no NPV de variar um parâmetro em +/- percentual.
type SensibilidadeResultado struct {
	Parametro          string
	NpvBaixo, NpvAlto  float64
	Amplitude          float64
}

// Analisar varia cada parâmetro financeiro em +/- percentualVariacao (padrão
// 20%), mantendo os demais no valor mais provável, e mede o quanto o NPV
// oscila -- ranqueado do maior impacto pro menor (estilo tornado chart).
func Analisar(f *config.Financeiro, percentualVariacao float64) []SensibilidadeResultado {
	base := ParamsDeterministicos(f)

	resultado := []SensibilidadeResultado{
		variar(base, "crescimentoMensal", base.CrescimentoMensal, percentualVariacao),
		variar(base, "custoPorChamadaStatusQuo", base.CustoPorChamadaStatusQuo, percentualVariacao),
		variar(base, "custoPorChamadaFineTuned", base.CustoPorChamadaFineTuned, percentualVariacao),
		variar(base, "custoTreinamento", base.CustoTreinamento, percentualVariacao),
	}

	sort.Slice(resultado, func(i, j int) bool { return resultado[i].Amplitude > resultado[j].Amplitude })
	return resultado
}

func variar(base NpvParams, parametro string, valorBase, percentual float64) SensibilidadeResultado {
	baixoParams := comValor(base, parametro, valorBase*(1-percentual))
	altoParams := comValor(base, parametro, valorBase*(1+percentual))

	npvBaixo := CalcularNPV(baixoParams).Npv
	npvAlto := CalcularNPV(altoParams).Npv

	amplitude := npvAlto - npvBaixo
	if amplitude < 0 {
		amplitude = -amplitude
	}

	return SensibilidadeResultado{Parametro: parametro, NpvBaixo: npvBaixo, NpvAlto: npvAlto, Amplitude: roundMoney(amplitude)}
}

func comValor(base NpvParams, parametro string, valor float64) NpvParams {
	p := base
	switch parametro {
	case "crescimentoMensal":
		p.CrescimentoMensal = valor
	case "custoPorChamadaStatusQuo":
		p.CustoPorChamadaStatusQuo = valor
	case "custoPorChamadaFineTuned":
		p.CustoPorChamadaFineTuned = valor
	case "custoTreinamento":
		p.CustoTreinamento = valor
	}
	return p
}
