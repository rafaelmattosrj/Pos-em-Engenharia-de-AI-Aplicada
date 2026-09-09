package cleaning

import (
	"math"
	"sort"
)

const AlphaTemperatura = 0.3

func ContarPorFonte(exemplos []Exemplo, caso string) map[string]int {
	contagem := map[string]int{}
	for _, e := range exemplos {
		if e.Metadata.Caso == caso {
			contagem[e.Metadata.Fonte]++
		}
	}
	return contagem
}

// PesosAmostragemPorTemperatura: peso de amostragem por fonte, p_i
// proporcional a n_i^alpha (Raffel et al. 2020; Xue et al. 2021).
func PesosAmostragemPorTemperatura(contagens map[string]int, alpha float64) map[string]float64 {
	pesosBrutos := map[string]float64{}
	soma := 0.0
	for fonte, n := range contagens {
		p := math.Pow(float64(n), alpha)
		pesosBrutos[fonte] = p
		soma += p
	}
	pesos := map[string]float64{}
	for fonte, p := range pesosBrutos {
		pesos[fonte] = p / soma
	}
	return pesos
}

// AlocarMaiorResto: metodo do maior resto (Hamilton/Hare-Niemeyer), converte
// pesos continuos em contagem inteira que soma exatamente ao alvo.
func AlocarMaiorResto(pesos map[string]float64, alvo int) map[string]int {
	fontes := make([]string, 0, len(pesos))
	for f := range pesos {
		fontes = append(fontes, f)
	}
	sort.Strings(fontes) // ordem deterministica antes do sort por resto

	quotas := make([]float64, len(fontes))
	base := make([]int, len(fontes))
	alocadoBase := 0
	for i, f := range fontes {
		quotas[i] = pesos[f] * float64(alvo)
		base[i] = int(math.Floor(quotas[i]))
		alocadoBase += base[i]
	}

	type resto struct {
		fonte string
		resto float64
	}
	restos := make([]resto, len(fontes))
	for i, f := range fontes {
		restos[i] = resto{fonte: f, resto: quotas[i] - float64(base[i])}
	}
	sort.SliceStable(restos, func(i, j int) bool { return restos[i].resto > restos[j].resto })

	faltam := alvo - alocadoBase
	resultado := map[string]int{}
	for i, f := range fontes {
		resultado[f] = base[i]
	}
	for i := 0; i < faltam; i++ {
		resultado[restos[i].fonte]++
	}
	return resultado
}

// AlocarComCapacidade: alocacao capacitada - aplica o peso por temperatura,
// mas nunca aloca mais do que a fonte realmente tem disponivel. Fontes que
// estourariam a capacidade sao fixadas no maximo disponivel, e o alvo
// restante e redistribuido por temperatura entre as fontes que sobraram -
// processo iterativo ate estabilizar.
func AlocarComCapacidade(contagens map[string]int, alpha float64, alvoTotal int) map[string]int {
	fontesAtivas := make([]string, 0, len(contagens))
	for f := range contagens {
		fontesAtivas = append(fontesAtivas, f)
	}
	alvoRestante := alvoTotal
	resultado := map[string]int{}
	maxIteracoes := len(contagens) + 1

	for iter := 0; iter < maxIteracoes && len(fontesAtivas) > 0; iter++ {
		contagensAtivas := map[string]int{}
		for _, f := range fontesAtivas {
			contagensAtivas[f] = contagens[f]
		}
		pesos := PesosAmostragemPorTemperatura(contagensAtivas, alpha)
		tentativa := AlocarMaiorResto(pesos, alvoRestante)

		var excedentes []string
		for _, f := range fontesAtivas {
			if tentativa[f] > contagens[f] {
				excedentes = append(excedentes, f)
			}
		}
		if len(excedentes) == 0 {
			for _, f := range fontesAtivas {
				resultado[f] = tentativa[f]
			}
			break
		}
		excedenteSet := map[string]struct{}{}
		for _, f := range excedentes {
			resultado[f] = contagens[f]
			alvoRestante -= contagens[f]
			excedenteSet[f] = struct{}{}
		}
		var restantes []string
		for _, f := range fontesAtivas {
			if _, ok := excedenteSet[f]; !ok {
				restantes = append(restantes, f)
			}
		}
		fontesAtivas = restantes
	}
	return resultado
}

type ResultadoBalanceamento struct {
	Exemplos  []Exemplo
	Alocacao  map[string]int
	Contagens map[string]int
}

func BalancearPorTemperatura(exemplos []Exemplo, caso string, alpha float64, alvoTotal int) ResultadoBalanceamento {
	contagens := ContarPorFonte(exemplos, caso)
	alocacao := AlocarComCapacidade(contagens, alpha, alvoTotal)

	var doCaso, outros []Exemplo
	for _, e := range exemplos {
		if e.Metadata.Caso == caso {
			doCaso = append(doCaso, e)
		} else {
			outros = append(outros, e)
		}
	}

	contadorUsado := map[string]int{}
	var selecionados []Exemplo
	for _, e := range doCaso {
		fonte := e.Metadata.Fonte
		if contadorUsado[fonte] < alocacao[fonte] {
			selecionados = append(selecionados, e)
			contadorUsado[fonte]++
		}
	}

	resultado := append(append([]Exemplo{}, outros...), selecionados...)
	return ResultadoBalanceamento{Exemplos: resultado, Alocacao: alocacao, Contagens: contagens}
}
