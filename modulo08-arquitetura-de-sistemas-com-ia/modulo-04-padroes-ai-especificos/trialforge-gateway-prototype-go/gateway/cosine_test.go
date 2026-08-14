package gateway

import (
	"math"
	"testing"
)

func TestSimilaridadeCosseno(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	ortogonal := []float64{0, 1, 0}
	oposto := []float64{-1, 0, 0}

	casos := []struct {
		nome     string
		a, b     []float64
		esperado float64
	}{
		{"vetores idênticos", a, b, 1},
		{"vetores ortogonais", a, ortogonal, 0},
		{"vetores opostos", a, oposto, -1},
	}

	for _, c := range casos {
		got := SimilaridadeCosseno(c.a, c.b)
		if math.Abs(got-c.esperado) > 1e-9 {
			t.Errorf("%s: SimilaridadeCosseno = %v, esperado %v", c.nome, got, c.esperado)
		}
	}
}
