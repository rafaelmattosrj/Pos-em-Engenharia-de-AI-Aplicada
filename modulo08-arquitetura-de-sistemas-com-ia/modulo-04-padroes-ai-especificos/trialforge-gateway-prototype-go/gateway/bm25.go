package gateway

import "math"

// EstatisticasBM25 guarda o que é pré-computado uma vez por corpus: tokens de
// cada documento, N, comprimento médio (avgdl) e document frequency por termo.
type EstatisticasBM25 struct {
	TokensPorDoc [][]string
	N            int
	Avgdl        float64
	DF           map[string]int
}

// ConstruirEstatisticasBM25 pré-computa as estatísticas do corpus — feito uma
// vez na indexação, nunca a cada busca.
func ConstruirEstatisticasBM25(documentos []string) EstatisticasBM25 {
	tokensPorDoc := make([][]string, len(documentos))
	somaComprimentos := 0
	for i, doc := range documentos {
		tokensPorDoc[i] = Tokenizar(doc)
		somaComprimentos += len(tokensPorDoc[i])
	}
	n := len(documentos)
	avgdl := float64(somaComprimentos) / float64(n)

	df := make(map[string]int)
	for _, tokens := range tokensPorDoc {
		vistos := make(map[string]bool)
		for _, termo := range tokens {
			if vistos[termo] {
				continue
			}
			vistos[termo] = true
			df[termo]++
		}
	}

	return EstatisticasBM25{TokensPorDoc: tokensPorDoc, N: n, Avgdl: avgdl, DF: df}
}

// ScoreBM25 calcula o score BM25 clássico (k1=1.5, b=0.75 por padrão) de um
// documento (já tokenizado) contra os tokens da query.
func ScoreBM25(queryTokens, docTokens []string, estatisticas EstatisticasBM25, k1, b float64) float64 {
	score := 0.0
	for _, termo := range queryTokens {
		freqNoDoc := 0
		for _, t := range docTokens {
			if t == termo {
				freqNoDoc++
			}
		}
		if freqNoDoc == 0 {
			continue
		}
		docFreq := estatisticas.DF[termo]
		// +1 evita idf negativo com poucos docs
		idf := math.Log((float64(estatisticas.N)-float64(docFreq)+0.5)/(float64(docFreq)+0.5) + 1)
		numerador := float64(freqNoDoc) * (k1 + 1)
		denominador := float64(freqNoDoc) + k1*(1-b+(b*float64(len(docTokens)))/estatisticas.Avgdl)
		score += idf * (numerador / denominador)
	}
	return score
}

// ScoreBM25Padrao chama ScoreBM25 com os defaults k1=1.5, b=0.75 usados nos
// dois originais.
func ScoreBM25Padrao(queryTokens, docTokens []string, estatisticas EstatisticasBM25) float64 {
	return ScoreBM25(queryTokens, docTokens, estatisticas, 1.5, 0.75)
}
