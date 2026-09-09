package finance

import (
	"math"
	"math/rand"
	"testing"

	"decision-framework-tool/config"
)

func TestAmostragemTriangularConvergeParaMediaTeorica(t *testing.T) {
	seed := int64(42)
	rng := func() float64 {
		seed = (seed*1103515245 + 12345) % 2147483648
		return float64(seed) / 2147483648.0
	}
	soma := 0.0
	n := 20000
	for i := 0; i < n; i++ {
		soma += AmostrarTriangular(0.01, 0.03, 0.05, rng)
	}
	media := soma / float64(n)
	mediaTeorica := (0.01 + 0.03 + 0.05) / 3
	if math.Abs(media-mediaTeorica) > 0.002 {
		t.Fatalf("media=%v mediaTeorica=%v", media, mediaTeorica)
	}
}

func TestCasoAutoTemProbabilidadeAltaDeNpvPositivo(t *testing.T) {
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	auto, err := cfg.Caso("amplitude-auto")
	if err != nil {
		t.Fatal(err)
	}
	mc := SimularMonteCarlo(auto.Financeiro, 2000, rand.Float64)
	if mc.ProbabilidadePositivo <= 0.9 {
		t.Fatalf("esperado probabilidadePositivo > 0.9, obtido %v", mc.ProbabilidadePositivo)
	}
}
