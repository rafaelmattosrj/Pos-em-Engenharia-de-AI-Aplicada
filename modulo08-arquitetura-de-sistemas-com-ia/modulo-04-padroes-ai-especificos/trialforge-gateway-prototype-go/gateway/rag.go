package gateway

import (
	"context"
	"fmt"
)

// Embedder é o único contrato de rede que o RAG precisa: transformar um texto
// num vetor de embedding. Implementado por ollama.Client — a busca em si
// (BuscarClausulaHibrida) nunca faz rede, só compara vetores já prontos.
type Embedder interface {
	Embedar(ctx context.Context, texto string) ([]float64, error)
}

// IndicePreparado é o resultado da indexação (Módulo 4.1) de UM domínio:
// embeddings de cada cláusula (tema e texto) e as estatísticas BM25 do
// corpus — tudo calculado uma vez, na inicialização.
type IndicePreparado struct {
	Clausulas       []Clausula
	Estatisticas    EstatisticasBM25
	EmbeddingsTema  [][]float64
	EmbeddingsTexto [][]float64
}

// PrepararIndices pré-computa embeddings de CADA cláusula (tema e texto) e as
// estatísticas BM25 do corpus, um índice por domínio — tudo calculado uma vez
// só. BuscarClausulaHibrida nunca chama Embedar() pro lado do corpus; só lê o
// que já está pronto aqui.
func PrepararIndices(ctx context.Context, embedder Embedder, logf func(format string, args ...interface{})) (map[string]IndicePreparado, error) {
	totalClausulas := 0
	for _, nome := range NomesIndices {
		totalClausulas += len(Indices[nome])
	}
	if logf != nil {
		logf("[Indexação] Pré-computando embeddings e estatísticas BM25 de %d cláusulas em %d índices — uma vez, não a cada busca...", totalClausulas, len(NomesIndices))
	}

	preparados := make(map[string]IndicePreparado, len(NomesIndices))
	for _, nomeIndice := range NomesIndices {
		clausulas := Indices[nomeIndice]
		textos := make([]string, len(clausulas))
		for i, c := range clausulas {
			textos[i] = c.Texto
		}
		estatisticas := ConstruirEstatisticasBM25(textos)

		embeddingsTema := make([][]float64, len(clausulas))
		embeddingsTexto := make([][]float64, len(clausulas))
		for i, c := range clausulas {
			embTema, err := embedder.Embedar(ctx, c.Tema)
			if err != nil {
				return nil, fmt.Errorf("indexação de %q (tema): %w", nomeIndice, err)
			}
			embeddingsTema[i] = embTema

			embTexto, err := embedder.Embedar(ctx, c.Texto)
			if err != nil {
				return nil, fmt.Errorf("indexação de %q (texto): %w", nomeIndice, err)
			}
			embeddingsTexto[i] = embTexto
		}

		preparados[nomeIndice] = IndicePreparado{
			Clausulas:       clausulas,
			Estatisticas:    estatisticas,
			EmbeddingsTema:  embeddingsTema,
			EmbeddingsTexto: embeddingsTexto,
		}
	}

	if logf != nil {
		logf("[Indexação] Concluída.\n")
	}
	return preparados, nil
}

// ResultadoBusca é o resultado de uma busca híbrida dentro de um índice.
type ResultadoBusca struct {
	Indice              string
	Clausula            Clausula
	SimilaridadeCosseno float64
	ScoreBM25           float64
	ScoreRRF            float64
	IteracoesUsadas     int
	EsgotouLimite       bool
}

// BuscarClausulaHibrida (Hybrid Search dentro de UM índice, já roteado pelo
// Multi-Index): combina léxico (BM25) com denso (embedding), fundidos por
// RANK via Reciprocal Rank Fusion — nunca somando os dois scores brutos
// direto, escalas incompatíveis. campoEmbedding escolhe, entre os embeddings
// JÁ pré-computados, comparar contra o tema (mais preciso) ou o texto inteiro
// da cláusula (mais abrangente) — o que o Agentic RAG varia entre tentativas.
func BuscarClausulaHibrida(preparados map[string]IndicePreparado, pergunta string, perguntaEmbedding []float64, nomeIndice, campoEmbedding string) ResultadoBusca {
	indice := preparados[nomeIndice]
	embeddingsCorpus := indice.EmbeddingsTema
	if campoEmbedding == "texto" {
		embeddingsCorpus = indice.EmbeddingsTexto
	}
	queryTokens := Tokenizar(pergunta)

	similaridades := make([]float64, len(embeddingsCorpus))
	for i, emb := range embeddingsCorpus {
		similaridades[i] = SimilaridadeCosseno(perguntaEmbedding, emb)
	}
	rankingDenso := OrdenarPorScore(similaridades)

	scoresBM25 := make([]float64, len(indice.Clausulas))
	for i := range indice.Clausulas {
		scoresBM25[i] = ScoreBM25Padrao(queryTokens, indice.Estatisticas.TokensPorDoc[i], indice.Estatisticas)
	}
	rankingEsparso := OrdenarPorScore(scoresBM25)

	fusao := FusaoReciprocalRank(rankingDenso, rankingEsparso, 60)
	melhor := fusao[0]

	return ResultadoBusca{
		Indice:              nomeIndice,
		Clausula:            indice.Clausulas[melhor.Idx],
		SimilaridadeCosseno: similaridades[melhor.Idx],
		ScoreBM25:           scoresBM25[melhor.Idx],
		ScoreRRF:            melhor.Score,
	}
}

// BuscarEmTodosIndices é o último recurso do Agentic RAG: cruza TODOS os
// índices, não só o roteado inicialmente.
func BuscarEmTodosIndices(preparados map[string]IndicePreparado, pergunta string, perguntaEmbedding []float64) ResultadoBusca {
	var melhor *ResultadoBusca
	for _, nomeIndice := range NomesIndices {
		resultado := BuscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndice, "texto")
		if melhor == nil || resultado.SimilaridadeCosseno > melhor.SimilaridadeCosseno {
			r := resultado
			melhor = &r
		}
	}
	return *melhor
}

// MaxIteracoesAgentic é o mesmo limite do retry com limite do Módulo 3.5:
// insistir além disso não ajuda mais.
const MaxIteracoesAgentic = 3

// BuscarClausulaAgentica (Agentic RAG, Módulo 4.1): busca → avalia confiança →
// busca de novo com estratégia mais ampla, até MaxIteracoesAgentic. Se nunca
// atingir o limiar, devolve o melhor achado e deixa o Confidence Threshold
// (Módulo 4.4) escalar pro Approval Gate — o RAG não decide sozinho quando
// desistir, ele só para de insistir e passa a decisão adiante.
func BuscarClausulaAgentica(preparados map[string]IndicePreparado, pergunta string, perguntaEmbedding []float64, nomeIndiceInicial string, limiarConfianca float64, logf func(format string, args ...interface{})) ResultadoBusca {
	var melhorResultado *ResultadoBusca

	for iteracao := 1; iteracao <= MaxIteracoesAgentic; iteracao++ {
		var resultado ResultadoBusca
		var estrategia string
		switch iteracao {
		case 1:
			estrategia = fmt.Sprintf("índice %q, comparando com o tema da cláusula", nomeIndiceInicial)
			resultado = BuscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndiceInicial, "tema")
		case 2:
			estrategia = fmt.Sprintf("índice %q, ampliando pro texto completo da cláusula", nomeIndiceInicial)
			resultado = BuscarClausulaHibrida(preparados, pergunta, perguntaEmbedding, nomeIndiceInicial, "texto")
		default:
			estrategia = "todos os índices, cruzando domínios (último recurso)"
			resultado = BuscarEmTodosIndices(preparados, pergunta, perguntaEmbedding)
		}

		if logf != nil {
			logf("  [Agentic RAG] Iteração %d/%d — %s — confiança %.3f", iteracao, MaxIteracoesAgentic, estrategia, resultado.SimilaridadeCosseno)
		}
		if melhorResultado == nil || resultado.SimilaridadeCosseno > melhorResultado.SimilaridadeCosseno {
			r := resultado
			melhorResultado = &r
		}

		if resultado.SimilaridadeCosseno >= limiarConfianca {
			final := *melhorResultado
			final.IteracoesUsadas = iteracao
			final.EsgotouLimite = false
			return final
		}
	}

	if logf != nil {
		logf("  [Agentic RAG] %d iterações esgotadas sem atingir o limiar — segue com o melhor resultado (confiança %.3f) e deixa o Confidence Threshold decidir.", MaxIteracoesAgentic, melhorResultado.SimilaridadeCosseno)
	}
	final := *melhorResultado
	final.IteracoesUsadas = MaxIteracoesAgentic
	final.EsgotouLimite = true
	return final
}
