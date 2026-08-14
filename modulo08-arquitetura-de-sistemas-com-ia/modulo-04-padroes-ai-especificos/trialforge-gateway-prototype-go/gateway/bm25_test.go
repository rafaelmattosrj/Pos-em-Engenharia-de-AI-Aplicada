package gateway

import "testing"

// Mesmo corpus/query dos testes puros dos dois originais: o doc que
// compartilha "idade" e "minima" com a query deve vencer a busca lexical.
func TestScoreBM25(t *testing.T) {
	corpus := []string{
		"idade mínima de doze anos para participar do estudo",
		"consentimento do responsável legal é obrigatório",
		"retirada do participante a qualquer momento sem justificativa",
	}
	estatisticas := ConstruirEstatisticasBM25(corpus)
	query := Tokenizar("qual a idade mínima exigida")

	melhorScore := -1.0
	vencedor := -1
	for idx := range corpus {
		score := ScoreBM25Padrao(query, estatisticas.TokensPorDoc[idx], estatisticas)
		if score > melhorScore {
			melhorScore = score
			vencedor = idx
		}
	}

	if vencedor != 0 {
		t.Errorf("esperado doc 0 vencer a busca lexical, venceu doc %d", vencedor)
	}
}

func TestScoreBM25SemTermoComum(t *testing.T) {
	corpus := []string{"gatos e cachorros", "carros e motos"}
	estatisticas := ConstruirEstatisticasBM25(corpus)
	query := Tokenizar("nenhumtermoemcomum")

	for idx := range corpus {
		score := ScoreBM25Padrao(query, estatisticas.TokensPorDoc[idx], estatisticas)
		if score != 0 {
			t.Errorf("esperado score 0 sem termos em comum, obtido %v", score)
		}
	}
}

func TestConstruirEstatisticasBM25(t *testing.T) {
	corpus := []string{"a b c", "a b"}
	estatisticas := ConstruirEstatisticasBM25(corpus)

	if estatisticas.N != 2 {
		t.Errorf("N = %d, esperado 2", estatisticas.N)
	}
	esperadoAvgdl := 2.5
	if estatisticas.Avgdl != esperadoAvgdl {
		t.Errorf("Avgdl = %v, esperado %v", estatisticas.Avgdl, esperadoAvgdl)
	}
	if estatisticas.DF["a"] != 2 {
		t.Errorf(`DF["a"] = %d, esperado 2`, estatisticas.DF["a"])
	}
	if estatisticas.DF["c"] != 1 {
		t.Errorf(`DF["c"] = %d, esperado 1`, estatisticas.DF["c"])
	}
}
