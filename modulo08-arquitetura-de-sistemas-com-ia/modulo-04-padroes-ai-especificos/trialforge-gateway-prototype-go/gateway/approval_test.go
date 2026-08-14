package gateway

import (
	"bytes"
	"strings"
	"testing"
)

func TestApprovalGateAprova(t *testing.T) {
	entrada := strings.NewReader("s\n")
	var saida bytes.Buffer
	gate := NewApprovalGate(entrada, &saida)

	aprovado, err := gate.PedirAprovacaoHumana("rascunho de teste")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !aprovado {
		t.Errorf("esperado aprovado=true")
	}
	if !strings.Contains(saida.String(), "rascunho de teste") {
		t.Errorf("esperado o rascunho ser exibido no prompt")
	}
}

func TestApprovalGateRejeita(t *testing.T) {
	entrada := strings.NewReader("n\n")
	var saida bytes.Buffer
	gate := NewApprovalGate(entrada, &saida)

	aprovado, err := gate.PedirAprovacaoHumana("rascunho")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if aprovado {
		t.Errorf("esperado aprovado=false")
	}
}

func TestApprovalGateEOFRejeita(t *testing.T) {
	entrada := strings.NewReader("")
	var saida bytes.Buffer
	gate := NewApprovalGate(entrada, &saida)

	aprovado, err := gate.PedirAprovacaoHumana("rascunho")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if aprovado {
		t.Errorf("esperado aprovado=false quando a entrada acaba (EOF)")
	}
}

// Mesma fila de respostas consumida em sequência por chamadas sucessivas —
// replica o cenário de entrada não-interativa (`printf "s\ns\n" | ...`) que
// motivou o bug corrigido nos dois originais (reader único, nunca recriado).
func TestApprovalGateMultiplasChamadasConsomeFila(t *testing.T) {
	entrada := strings.NewReader("s\nn\ns\n")
	var saida bytes.Buffer
	gate := NewApprovalGate(entrada, &saida)

	respostas := make([]bool, 3)
	for i := range respostas {
		aprovado, err := gate.PedirAprovacaoHumana("rascunho")
		if err != nil {
			t.Fatalf("erro inesperado na chamada %d: %v", i, err)
		}
		respostas[i] = aprovado
	}

	esperado := []bool{true, false, true}
	for i, e := range esperado {
		if respostas[i] != e {
			t.Errorf("chamada %d: aprovado = %v, esperado %v", i, respostas[i], e)
		}
	}
}
