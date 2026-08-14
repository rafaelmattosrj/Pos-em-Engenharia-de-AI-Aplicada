package gateway

import "math"

// SimilaridadeCosseno é a base do Semantic Cache e do score de confiança da
// busca de cláusula do RAG — usa embeddings reais, não heurística de palavra-chave.
func SimilaridadeCosseno(a, b []float64) float64 {
	produto := 0.0
	normaA := 0.0
	normaB := 0.0
	for i := range a {
		produto += a[i] * b[i]
		normaA += a[i] * a[i]
		normaB += b[i] * b[i]
	}
	return produto / (math.Sqrt(normaA) * math.Sqrt(normaB))
}
