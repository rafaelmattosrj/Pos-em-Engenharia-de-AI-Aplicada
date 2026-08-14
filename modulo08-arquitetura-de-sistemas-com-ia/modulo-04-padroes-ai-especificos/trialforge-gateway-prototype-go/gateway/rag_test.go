package gateway

import (
	"context"
	"testing"
)

func TestBuscarClausulaHibrida(t *testing.T) {
	preparados := map[string]IndicePreparado{
		"teste": {
			Clausulas: []Clausula{
				{Tema: "tema um", Texto: "texto sobre gatos e felinos", Fonte: "fonte1"},
				{Tema: "tema dois", Texto: "texto sobre carros e motos", Fonte: "fonte2"},
			},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"texto sobre gatos e felinos", "texto sobre carros e motos"}),
			EmbeddingsTema:  [][]float64{{1, 0}, {0, 1}},
			EmbeddingsTexto: [][]float64{{1, 0}, {0, 1}},
		},
	}

	resultado := BuscarClausulaHibrida(preparados, "gatos", []float64{1, 0}, "teste", "tema")

	if resultado.Clausula.Tema != "tema um" {
		t.Errorf("esperado 'tema um', obtido %q", resultado.Clausula.Tema)
	}
	if resultado.SimilaridadeCosseno != 1 {
		t.Errorf("similaridade = %v, esperado 1", resultado.SimilaridadeCosseno)
	}
	if resultado.Indice != "teste" {
		t.Errorf("indice = %q, esperado 'teste'", resultado.Indice)
	}
}

func TestBuscarEmTodosIndices(t *testing.T) {
	preparados := map[string]IndicePreparado{
		"icf": {
			Clausulas:       []Clausula{{Tema: "icf tema", Texto: "icf texto", Fonte: "f1"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"icf texto"}),
			EmbeddingsTema:  [][]float64{{0, 1}},
			EmbeddingsTexto: [][]float64{{0, 1}},
		},
		"protocolo": {
			Clausulas:       []Clausula{{Tema: "protocolo tema", Texto: "protocolo texto", Fonte: "f2"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"protocolo texto"}),
			EmbeddingsTema:  [][]float64{{1, 0}},
			EmbeddingsTexto: [][]float64{{1, 0}},
		},
		"csr": {
			Clausulas:       []Clausula{{Tema: "csr tema", Texto: "csr texto", Fonte: "f3"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"csr texto"}),
			EmbeddingsTema:  [][]float64{{0, -1}},
			EmbeddingsTexto: [][]float64{{0, -1}},
		},
	}

	resultado := BuscarEmTodosIndices(preparados, "pergunta qualquer", []float64{1, 0})

	if resultado.Indice != "protocolo" {
		t.Errorf("esperado a busca cruzada escolher 'protocolo' (maior cosseno), obtido %q", resultado.Indice)
	}
}

// Constrói um cenário onde a 1ª iteração (comparação com o tema) fica abaixo
// do limiar, mas a 2ª (texto completo) atinge — replica o comportamento de
// "amplia a estratégia até achar confiança suficiente" do Agentic RAG.
func TestBuscarClausulaAgenticaConvergeNaSegundaIteracao(t *testing.T) {
	preparados := map[string]IndicePreparado{
		"icf": {
			Clausulas:       []Clausula{{Tema: "tema fraco", Texto: "texto forte", Fonte: "f1"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"texto forte"}),
			EmbeddingsTema:  [][]float64{{0.5, 0.5}},
			EmbeddingsTexto: [][]float64{{1, 0}},
		},
	}

	resultado := BuscarClausulaAgentica(preparados, "pergunta", []float64{1, 0}, "icf", 0.9, nil)

	if resultado.IteracoesUsadas != 2 {
		t.Errorf("IteracoesUsadas = %d, esperado 2", resultado.IteracoesUsadas)
	}
	if resultado.EsgotouLimite {
		t.Errorf("EsgotouLimite = true, esperado false (convergiu na 2ª iteração)")
	}
}

// Nenhuma das 3 estratégias atinge o limiar — Agentic RAG esgota o limite e
// devolve o melhor resultado encontrado, deixando o Confidence Threshold decidir.
func TestBuscarClausulaAgenticaEsgotaLimite(t *testing.T) {
	preparados := map[string]IndicePreparado{
		"icf": {
			Clausulas:       []Clausula{{Tema: "tema", Texto: "texto", Fonte: "f1"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"texto"}),
			EmbeddingsTema:  [][]float64{{0, 1}},
			EmbeddingsTexto: [][]float64{{0, 1}},
		},
		"protocolo": {
			Clausulas:       []Clausula{{Tema: "tema2", Texto: "texto2", Fonte: "f2"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"texto2"}),
			EmbeddingsTema:  [][]float64{{0, 1}},
			EmbeddingsTexto: [][]float64{{0, 1}},
		},
		"csr": {
			Clausulas:       []Clausula{{Tema: "tema3", Texto: "texto3", Fonte: "f3"}},
			Estatisticas:    ConstruirEstatisticasBM25([]string{"texto3"}),
			EmbeddingsTema:  [][]float64{{0, 1}},
			EmbeddingsTexto: [][]float64{{0, 1}},
		},
	}

	resultado := BuscarClausulaAgentica(preparados, "pergunta", []float64{1, 0}, "icf", 0.9, nil)

	if resultado.IteracoesUsadas != MaxIteracoesAgentic {
		t.Errorf("IteracoesUsadas = %d, esperado %d", resultado.IteracoesUsadas, MaxIteracoesAgentic)
	}
	if !resultado.EsgotouLimite {
		t.Errorf("EsgotouLimite = false, esperado true")
	}
}

// fakeEmbedder devolve um vetor determinístico (não-nulo) a partir do
// comprimento do texto — suficiente pra exercitar PrepararIndices sem rede.
type fakeEmbedder struct{ chamadas int }

func (f *fakeEmbedder) Embedar(_ context.Context, texto string) ([]float64, error) {
	f.chamadas++
	return []float64{float64(len(texto)), 1}, nil
}

func TestPrepararIndices(t *testing.T) {
	fake := &fakeEmbedder{}
	preparados, err := PrepararIndices(context.Background(), fake, nil)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(preparados) != len(NomesIndices) {
		t.Fatalf("esperado %d índices preparados, obtido %d", len(NomesIndices), len(preparados))
	}
	for _, nome := range NomesIndices {
		indice, ok := preparados[nome]
		if !ok {
			t.Fatalf("índice %q não foi preparado", nome)
		}
		if len(indice.EmbeddingsTema) != len(indice.Clausulas) || len(indice.EmbeddingsTexto) != len(indice.Clausulas) {
			t.Errorf("índice %q: quantidade de embeddings não bate com quantidade de cláusulas", nome)
		}
	}

	// 2 cláusulas por índice x 3 índices x 2 embeddings (tema + texto) = 12 chamadas
	if fake.chamadas != 12 {
		t.Errorf("esperado 12 chamadas de embedding, obtido %d", fake.chamadas)
	}
}
