package gateway

import "sort"

// scoreIdx associa um índice de documento ao seu score, usado só para ordenar.
type scoreIdx struct {
	idx   int
	score float64
}

// OrdenarPorScore devolve os índices dos documentos ordenados por score
// decrescente — o "ranking" usado tanto pela busca densa (cosseno) quanto
// pela esparsa (BM25) antes da fusão.
func OrdenarPorScore(scores []float64) []int {
	pares := make([]scoreIdx, len(scores))
	for i, s := range scores {
		pares[i] = scoreIdx{idx: i, score: s}
	}
	sort.SliceStable(pares, func(i, j int) bool {
		return pares[i].score > pares[j].score
	})
	resultado := make([]int, len(pares))
	for i, p := range pares {
		resultado[i] = p.idx
	}
	return resultado
}

// RankedScore é uma entrada (índice de documento, score RRF) do resultado da
// fusão, já ordenada do melhor pro pior.
type RankedScore struct {
	Idx   int
	Score float64
}

// FusaoReciprocalRank (Cormack, Clarke & Büttcher, SIGIR 2009): funde dois
// rankings pela POSIÇÃO de cada documento em cada um, não pelo valor do
// score — é assim que se combina BM25 (escala aberta) com cosseno (-1 a 1)
// sem normalizar nada à mão.
func FusaoReciprocalRank(ranking1, ranking2 []int, k int) []RankedScore {
	scores := make(map[int]float64)
	ordemPrimeiraAparicao := make([]int, 0, len(ranking1)+len(ranking2))
	vistos := make(map[int]bool)

	acumular := func(ranking []int) {
		for posicao, idx := range ranking {
			scores[idx] += 1 / float64(k+posicao+1)
			if !vistos[idx] {
				vistos[idx] = true
				ordemPrimeiraAparicao = append(ordemPrimeiraAparicao, idx)
			}
		}
	}
	acumular(ranking1)
	acumular(ranking2)

	resultado := make([]RankedScore, len(ordemPrimeiraAparicao))
	for i, idx := range ordemPrimeiraAparicao {
		resultado[i] = RankedScore{Idx: idx, Score: scores[idx]}
	}
	sort.SliceStable(resultado, func(i, j int) bool {
		return resultado[i].Score > resultado[j].Score
	})
	return resultado
}
