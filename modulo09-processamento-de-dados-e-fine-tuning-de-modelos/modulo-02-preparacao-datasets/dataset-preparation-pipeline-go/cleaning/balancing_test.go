package cleaning

import (
	"math"
	"testing"
)

func TestAlpha1SemSuavizacaoReproduzDistribuicaoProporcional(t *testing.T) {
	contagens := map[string]int{"A": 8, "B": 2}
	pesos := PesosAmostragemPorTemperatura(contagens, 1)
	if math.Abs(pesos["A"]-0.8) > 1e-9 || math.Abs(pesos["B"]-0.2) > 1e-9 {
		t.Fatalf("esperado A=0.8 B=0.2, obtive %+v", pesos)
	}
}

func TestAlphaMenorSuavizaEmDirecaoAUniforme(t *testing.T) {
	contagens := map[string]int{"A": 8, "B": 2}
	baixo := PesosAmostragemPorTemperatura(contagens, 0.3)
	alto := PesosAmostragemPorTemperatura(contagens, 1)
	if !(baixo["A"] < alto["A"]) || !(baixo["B"] > alto["B"]) {
		t.Fatalf("esperado suavizacao em direcao a uniforme, obtive baixo=%+v alto=%+v", baixo, alto)
	}
}

func TestAlocacaoSemRestricaoSomaExatamenteAoAlvo(t *testing.T) {
	pesos := map[string]float64{"A": 0.5, "B": 0.3, "C": 0.2}
	aloc := AlocarMaiorResto(pesos, 17)
	soma := 0
	for _, v := range aloc {
		soma += v
	}
	if soma != 17 {
		t.Fatalf("esperado soma 17, obtive %d", soma)
	}
}

func TestAlocacaoCapacitadaNuncaExcedeCapacidade(t *testing.T) {
	contagens := map[string]int{"Oficina Estrela": 14, "Auto Center Silva": 5, "Funilaria Rio Bonito": 4, "Oficina Nova Alianca": 3}
	for _, alvo := range []int{26, 22, 20, 18} {
		aloc := AlocarComCapacidade(contagens, AlphaTemperatura, alvo)
		for f, n := range aloc {
			if n > contagens[f] {
				t.Errorf("alvo=%d: %s alocado %d > capacidade %d", alvo, f, n, contagens[f])
			}
		}
	}
}

func TestAlocacaoCapacitadaSomaExatamenteAoAlvo(t *testing.T) {
	contagens := map[string]int{"Oficina Estrela": 14, "Auto Center Silva": 5, "Funilaria Rio Bonito": 4, "Oficina Nova Alianca": 3}
	for _, alvo := range []int{26, 22, 20, 18} {
		aloc := AlocarComCapacidade(contagens, AlphaTemperatura, alvo)
		soma := 0
		for _, n := range aloc {
			soma += n
		}
		if soma != alvo {
			t.Errorf("alvo=%d: soma obtida %d", alvo, soma)
		}
	}
}
