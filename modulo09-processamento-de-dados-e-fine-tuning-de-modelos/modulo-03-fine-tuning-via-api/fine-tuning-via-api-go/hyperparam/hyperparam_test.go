package hyperparam

import (
	"strings"
	"testing"
)

func TestAceitaConfigRealUsadaNoJob(t *testing.T) {
	if err := Validar(Hiperparametros{EpochCount: 3, LearningRateMultiplier: 5.0}); err != nil {
		t.Fatal(err)
	}
}

func TestRejeitaEpochCountZero(t *testing.T) {
	err := Validar(Hiperparametros{EpochCount: 0, LearningRateMultiplier: 5.0})
	if err == nil || !strings.Contains(err.Error(), "epochCount deve ser inteiro entre 1 e 20") {
		t.Fatalf("esperava erro de epochCount, obtido: %v", err)
	}
}

func TestRejeitaEpocasNegativas(t *testing.T) {
	if err := Validar(Hiperparametros{EpochCount: -3, LearningRateMultiplier: 1}); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestRejeitaLearningRateForaDaFaixa(t *testing.T) {
	err := Validar(Hiperparametros{EpochCount: 3, LearningRateMultiplier: 50})
	if err == nil || !strings.Contains(err.Error(), "learningRateMultiplier") {
		t.Fatalf("esperava erro de learningRateMultiplier, obtido: %v", err)
	}
}

func TestNenhumaDivergenciaQuandoTudoBate(t *testing.T) {
	d := CompararHiperparametros(map[string]any{"epochCount": 3}, map[string]any{"epochCount": 3})
	if len(d) != 0 {
		t.Fatalf("esperava zero divergencias, obtido %v", d)
	}
}

func TestDetectaCasoRealEpochCountAusente(t *testing.T) {
	d := CompararHiperparametros(map[string]any{"epochCount": 0}, map[string]any{"epochCount": nil})
	if len(d) != 1 || !strings.Contains(d[0], "default silencioso") {
		t.Fatalf("resultado inesperado: %v", d)
	}
}

func TestDetectaDivergenciaDeValorNumerico(t *testing.T) {
	d := CompararHiperparametros(map[string]any{"epochCount": 3}, map[string]any{"epochCount": 5})
	if len(d) != 1 {
		t.Fatalf("esperava 1 divergencia, obtido %v", d)
	}
}

func TestNaoApontaDivergenciaQuandoTipoDiferenteMesmoValor(t *testing.T) {
	d := CompararHiperparametros(
		map[string]any{"epochCount": 3, "learningRateMultiplier": 5},
		map[string]any{"epochCount": "3", "learningRateMultiplier": "5"})
	if len(d) != 0 {
		t.Fatalf("esperava zero divergencias (tipos diferentes, valor igual), obtido %v", d)
	}
}
