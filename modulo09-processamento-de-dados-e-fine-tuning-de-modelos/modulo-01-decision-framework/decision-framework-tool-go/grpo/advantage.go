package grpo

import "math"

// epsAdvantage guarda contra divisão por zero quando o grupo empata (desvio padrão = 0).
const epsAdvantage = 1e-4

// VantagemResultado é o resultado de VantagemRelativaAoGrupo.
type VantagemResultado struct {
	Vantagens    []float64
	Media, Desvio float64
}

// VantagemRelativaAoGrupo calcula A_i = (r_i - média(r)) / desvio_padrao(r) --
// o coração do GRPO: o próprio grupo vira a linha de base, sem precisar de um
// modelo de valor (critic) separado como no PPO clássico.
func VantagemRelativaAoGrupo(recompensas []float64) VantagemResultado {
	media := 0.0
	for _, r := range recompensas {
		media += r
	}
	media /= float64(len(recompensas))

	variancia := 0.0
	for _, r := range recompensas {
		variancia += math.Pow(r-media, 2)
	}
	variancia /= float64(len(recompensas))
	desvio := math.Sqrt(variancia)

	desvioEfetivo := desvio
	if desvioEfetivo == 0 {
		desvioEfetivo = epsAdvantage
	}

	vantagens := make([]float64, len(recompensas))
	for i, r := range recompensas {
		vantagens[i] = (r - media) / desvioEfetivo
	}

	return VantagemResultado{Vantagens: vantagens, Media: media, Desvio: desvio}
}
