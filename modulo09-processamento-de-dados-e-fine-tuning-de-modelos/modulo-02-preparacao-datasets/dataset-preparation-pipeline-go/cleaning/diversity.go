package cleaning

import "math"

func DistribuicaoDe(contagens map[string]int) map[string]float64 {
	total := 0
	for _, n := range contagens {
		total += n
	}
	dist := map[string]float64{}
	for fonte, n := range contagens {
		if total == 0 {
			dist[fonte] = 0
		} else {
			dist[fonte] = float64(n) / float64(total)
		}
	}
	return dist
}

// EntropiaShannon em nats (log natural). H=0 significa uma unica fonte dominando tudo.
func EntropiaShannon(distribuicao map[string]float64) float64 {
	h := 0.0
	for _, p := range distribuicao {
		if p > 0 {
			h += p * math.Log(p)
		}
	}
	return -h
}

// NumeroEfetivoFontes: Hill number de ordem 1 = exp(H).
func NumeroEfetivoFontes(distribuicao map[string]float64) float64 {
	return math.Exp(EntropiaShannon(distribuicao))
}
