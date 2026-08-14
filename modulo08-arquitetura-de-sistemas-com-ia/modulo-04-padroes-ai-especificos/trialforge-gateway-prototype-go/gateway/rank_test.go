package gateway

import "testing"

func TestOrdenarPorScore(t *testing.T) {
	scores := []float64{0.2, 0.9, 0.5}
	got := OrdenarPorScore(scores)
	esperado := []int{1, 2, 0}
	for i := range esperado {
		if got[i] != esperado[i] {
			t.Fatalf("OrdenarPorScore(%v) = %v, esperado %v", scores, got, esperado)
		}
	}
}

// Mesmo caso dos testes puros: ranking1 e ranking2 concordam que o doc 0 é o
// melhor — sem empate, resultado não depende de estabilidade de sort.
func TestFusaoReciprocalRank(t *testing.T) {
	fusao := FusaoReciprocalRank([]int{0, 1, 2}, []int{0, 2, 1}, 60)
	if fusao[0].Idx != 0 {
		t.Errorf("esperado doc 0 no topo da fusão, obtido doc %d", fusao[0].Idx)
	}
}

func TestFusaoReciprocalRankTodosOsDocumentos(t *testing.T) {
	fusao := FusaoReciprocalRank([]int{0, 1}, []int{2, 3}, 60)
	if len(fusao) != 4 {
		t.Fatalf("esperado 4 documentos na fusão, obtido %d", len(fusao))
	}
}
