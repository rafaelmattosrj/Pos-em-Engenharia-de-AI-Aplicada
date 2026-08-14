package tiering

import "math"

// SimilaridadeCosseno é lógica pura, testável sem rede.
func SimilaridadeCosseno(a, b []float64) float64 {
	var produto, normaA, normaB float64
	for i := range a {
		produto += a[i] * b[i]
		normaA += a[i] * a[i]
		normaB += b[i] * b[i]
	}
	return produto / (math.Sqrt(normaA) * math.Sqrt(normaB))
}
