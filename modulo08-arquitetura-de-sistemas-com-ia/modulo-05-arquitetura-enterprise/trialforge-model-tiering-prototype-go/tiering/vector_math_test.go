package tiering

import (
	"math"
	"testing"
)

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestSimilaridadeCosseno_VetoresIdenticos(t *testing.T) {
	if got := SimilaridadeCosseno([]float64{1, 0, 0}, []float64{1, 0, 0}); !closeEnough(got, 1.0) {
		t.Errorf("esperava 1.0, obteve %v", got)
	}
}

func TestSimilaridadeCosseno_VetoresOrtogonais(t *testing.T) {
	if got := SimilaridadeCosseno([]float64{1, 0, 0}, []float64{0, 1, 0}); !closeEnough(got, 0.0) {
		t.Errorf("esperava 0.0, obteve %v", got)
	}
}
