package cleaning

import (
	"math"
	"testing"
)

func TestDistribuicaoUniformeTemNumeroEfetivoIgualAN(t *testing.T) {
	dist := map[string]float64{"A": 0.25, "B": 0.25, "C": 0.25, "D": 0.25}
	if math.Abs(NumeroEfetivoFontes(dist)-4.0) > 1e-9 {
		t.Fatalf("esperado 4.0, obtive %v", NumeroEfetivoFontes(dist))
	}
}

func TestUmaUnicaFonteTemEntropiaZeroENumeroEfetivo1(t *testing.T) {
	dist := map[string]float64{"A": 1.0, "B": 0.0, "C": 0.0}
	if math.Abs(EntropiaShannon(dist)) > 1e-9 {
		t.Fatalf("esperado entropia 0, obtive %v", EntropiaShannon(dist))
	}
	if math.Abs(NumeroEfetivoFontes(dist)-1.0) > 1e-9 {
		t.Fatalf("esperado numero efetivo 1, obtive %v", NumeroEfetivoFontes(dist))
	}
}

func TestSuavizarComAlphaMenorAumentaEntropia(t *testing.T) {
	contagens := map[string]int{"A": 14, "B": 5, "C": 4, "D": 3}
	hAlto := EntropiaShannon(PesosAmostragemPorTemperatura(contagens, 1.0))
	hBaixo := EntropiaShannon(PesosAmostragemPorTemperatura(contagens, 0.3))
	if !(hBaixo > hAlto) {
		t.Fatalf("esperado hBaixo > hAlto, obtive hBaixo=%v hAlto=%v", hBaixo, hAlto)
	}
}
