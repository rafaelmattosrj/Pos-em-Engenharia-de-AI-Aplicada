package ahp

import (
	"math"
	"testing"

	"decision-framework-tool/config"
)

func carregarMatriz(t *testing.T) [][]float64 {
	t.Helper()
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Ahp.Matriz
}

func TestPesosSomam1(t *testing.T) {
	pesos := DerivarPesos(carregarMatriz(t))
	soma := 0.0
	for _, p := range pesos {
		soma += p
	}
	if math.Abs(soma-1.0) > 1e-9 {
		t.Fatalf("esperado soma=1.0, obtido %v", soma)
	}
}

func TestPergunta3RecebeOMaiorPeso(t *testing.T) {
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	pesos := DerivarPesos(cfg.Ahp.Matriz)
	indiceP3 := -1
	for i, chave := range cfg.Ahp.OrdemPerguntas {
		if chave == "p3" {
			indiceP3 = i
		}
	}
	if indiceP3 == -1 {
		t.Fatal("p3 nao encontrado em ordemPerguntas")
	}
	maiorPeso := 0.0
	for _, p := range pesos {
		if p > maiorPeso {
			maiorPeso = p
		}
	}
	if pesos[indiceP3] != maiorPeso {
		t.Fatalf("esperado p3 com o maior peso: pesos=%v p3=%v maior=%v", pesos, pesos[indiceP3], maiorPeso)
	}
}

func TestMatrizEConsistente(t *testing.T) {
	matriz := carregarMatriz(t)
	pesos := DerivarPesos(matriz)
	consistencia := CalcularConsistencia(matriz, pesos)
	if !consistencia.Consistente {
		t.Fatalf("esperado CR < 0.10, obtido CR=%v", consistencia.CR)
	}
}

func TestComiteDeUmAvaliadorReproduzOsPesosOriginais(t *testing.T) {
	matrizSolo := [][]float64{{1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1}}
	agregada := AgregarMatrizesComite([][][]float64{matrizSolo})
	pesosSolo := DerivarPesos(agregada)
	pesosOriginais := DerivarPesos(matrizSolo)
	for i := range pesosSolo {
		if math.Abs(pesosSolo[i]-pesosOriginais[i]) > 1e-9 {
			t.Fatalf("peso %d divergente: solo=%v original=%v", i, pesosSolo[i], pesosOriginais[i])
		}
	}
}

func TestMatrizAgregadaEReciprocamenteValida(t *testing.T) {
	m1 := [][]float64{{1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1}}
	m2 := [][]float64{{1, 2, 1.0 / 5, 1.0 / 3}, {0.5, 1, 1.0 / 5, 1.0 / 3}, {5, 5, 1, 3}, {3, 3, 1.0 / 3, 1}}
	agregada := AgregarMatrizesComite([][][]float64{m1, m2})
	for i := range agregada {
		for j := range agregada {
			if math.Abs(agregada[i][j]*agregada[j][i]-1.0) > 1e-9 {
				t.Fatalf("par (%d,%d) nao reciproco: %v * %v", i, j, agregada[i][j], agregada[j][i])
			}
		}
	}
}
