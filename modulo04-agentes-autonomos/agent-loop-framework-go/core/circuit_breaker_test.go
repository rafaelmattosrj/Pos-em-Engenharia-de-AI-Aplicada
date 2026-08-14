package core

import "testing"

func TestCircuitBreaker_BreaksAfter3ConsecutiveInvalid(t *testing.T) {
	cb := NewCircuitBreaker()

	cb.RecordInvalid()
	cb.RecordInvalid()
	if cb.ShouldBreak() {
		t.Fatal("NAO deveria abrir antes de atingir o threshold")
	}

	cb.RecordInvalid()
	if !cb.ShouldBreak() {
		t.Fatal("DEVE abrir apos 3 respostas invalidas consecutivas")
	}
	if cb.InvalidCount() != 3 {
		t.Errorf("esperava invalidCount=3, obteve %d", cb.InvalidCount())
	}

	cb.RecordValid()
	if cb.ShouldBreak() {
		t.Error("deve ser resetado apos resposta valida")
	}
	if cb.InvalidCount() != 0 {
		t.Errorf("esperava invalidCount=0 apos reset, obteve %d", cb.InvalidCount())
	}
}
