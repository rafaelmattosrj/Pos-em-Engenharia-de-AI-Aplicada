package minhash

import (
	"math"
	"testing"
)

func TestAlphaUmReproduzDistribuicaoProporcional(t *testing.T) {
	pesos := PesosAmostragemPorTemperatura(map[string]int{"A": 8, "B": 2}, 1)
	if math.Abs(pesos["A"]-0.8) > 1e-9 || math.Abs(pesos["B"]-0.2) > 1e-9 {
		t.Fatalf("pesos inesperados: %+v", pesos)
	}
}

func TestAlphaMenorSuavizaDistribuicao(t *testing.T) {
	contagens := map[string]int{"A": 8, "B": 2}
	baixo := PesosAmostragemPorTemperatura(contagens, 0.3)
	alto := PesosAmostragemPorTemperatura(contagens, 1)
	if baixo["A"] >= alto["A"] || baixo["B"] <= alto["B"] {
		t.Fatalf("suavizacao nao ocorreu: baixo=%+v alto=%+v", baixo, alto)
	}
}

func TestAlocacaoSemRestricaoSomaExatamenteAoAlvo(t *testing.T) {
	aloc := AlocarMaiorResto(map[string]float64{"A": 0.5, "B": 0.3, "C": 0.2}, 17)
	soma := 0
	for _, v := range aloc {
		soma += v
	}
	if soma != 17 {
		t.Fatalf("soma inesperada: %d", soma)
	}
}

func TestAlocacaoCapacitadaNuncaExcedeCapacidade(t *testing.T) {
	contagens := map[string]int{"Oficina Estrela": 14, "Auto Center Silva": 5, "Funilaria Rio Bonito": 4, "Oficina Nova Aliança": 3}
	for _, alvo := range []int{26, 22, 20, 18} {
		aloc := AlocarComCapacidade(contagens, AlphaTemperatura, alvo)
		for f, n := range aloc {
			if n > contagens[f] {
				t.Fatalf("%s: alocado %d > capacidade %d", f, n, contagens[f])
			}
		}
	}
}

func TestAlocacaoCapacitadaSomaExatamenteAoAlvo(t *testing.T) {
	contagens := map[string]int{"Oficina Estrela": 14, "Auto Center Silva": 5, "Funilaria Rio Bonito": 4, "Oficina Nova Aliança": 3}
	for _, alvo := range []int{26, 22, 20, 18} {
		aloc := AlocarComCapacidade(contagens, AlphaTemperatura, alvo)
		soma := 0
		for _, n := range aloc {
			soma += n
		}
		if soma != alvo {
			t.Fatalf("alvo=%d soma=%d", alvo, soma)
		}
	}
}

func TestDistribuicaoUniformeTemNumeroEfetivoIgualAN(t *testing.T) {
	dist := map[string]float64{"A": 0.25, "B": 0.25, "C": 0.25, "D": 0.25}
	if got := NumeroEfetivoFontes(dist); math.Abs(got-4) > 1e-9 {
		t.Fatalf("esperado 4, obtido %v", got)
	}
}

func TestUmaUnicaFonteTemEntropiaZero(t *testing.T) {
	dist := map[string]float64{"A": 1.0, "B": 0, "C": 0}
	if math.Abs(EntropiaShannon(dist)) > 1e-9 || math.Abs(NumeroEfetivoFontes(dist)-1) > 1e-9 {
		t.Fatalf("entropia/N-efetivo inesperados: H=%v N=%v", EntropiaShannon(dist), NumeroEfetivoFontes(dist))
	}
}

func TestAssinaturaMinHashDeTextoIdenticoDaSimilaridadeUm(t *testing.T) {
	coef := GerarCoeficientesHash(MinHashK, MinHashSemente)
	sig := AssinaturaMinHash(Shingles("texto de teste com varias palavras diferentes aqui", 5), coef)
	if SimilaridadeMinHashEstimada(sig, sig) != 1 {
		t.Fatal("similaridade consigo mesmo deveria ser 1.0")
	}
}
