package grpo

import (
	"math"
	"testing"
)

func TestGrupoComVarianciaRealGeraVantagensPositivasENegativas(t *testing.T) {
	resultado := VantagemRelativaAoGrupo([]float64{1.0, 0.5, 0.0})
	if math.Abs(resultado.Media-0.5) > 1e-9 {
		t.Fatalf("esperado media=0.5, obtido %v", resultado.Media)
	}
	if resultado.Vantagens[0] <= 0 {
		t.Fatalf("esperado vantagem[0] > 0, obtido %v", resultado.Vantagens[0])
	}
	if resultado.Vantagens[2] >= 0 {
		t.Fatalf("esperado vantagem[2] < 0, obtido %v", resultado.Vantagens[2])
	}
	if math.Abs(resultado.Vantagens[1]) > 1e-9 {
		t.Fatalf("esperado vantagem[1]~=0, obtido %v", resultado.Vantagens[1])
	}
}

func TestGrupoDegeneradoProduzVantagemZeroSemDividirPorZero(t *testing.T) {
	resultado := VantagemRelativaAoGrupo([]float64{0.83, 0.83, 0.83, 0.83})
	if resultado.Desvio != 0.0 {
		t.Fatalf("esperado desvio=0, obtido %v", resultado.Desvio)
	}
	for i, v := range resultado.Vantagens {
		if v != 0.0 {
			t.Fatalf("vantagem[%d] esperado 0.0, obtido %v", i, v)
		}
	}
}

func TestRecompensasIguaisDentroDoGrupoRecebemAMesmaVantagem(t *testing.T) {
	resultado := VantagemRelativaAoGrupo([]float64{1.0, 1.0, 0.0, 0.0})
	if resultado.Vantagens[0] != resultado.Vantagens[1] {
		t.Fatalf("esperado vantagens[0]==vantagens[1]")
	}
	if resultado.Vantagens[2] != resultado.Vantagens[3] {
		t.Fatalf("esperado vantagens[2]==vantagens[3]")
	}
	if !(resultado.Vantagens[0] > resultado.Vantagens[2]) {
		t.Fatalf("esperado vantagens[0] > vantagens[2]")
	}
}
