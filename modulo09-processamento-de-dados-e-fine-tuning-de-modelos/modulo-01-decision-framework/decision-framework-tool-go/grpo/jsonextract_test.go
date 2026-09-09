package grpo

import "testing"

func TestExtraiJsonCercadoDeTextoSolto(t *testing.T) {
	resultado := ExtrairJSON(`texto antes {"a": 1, "b": 2} texto depois`)
	if resultado == nil {
		t.Fatal("esperado json extraido, obtido nil")
	}
	if resultado["a"] != float64(1) || resultado["b"] != float64(2) {
		t.Fatalf("json extraido incorreto: %v", resultado)
	}
}

func TestRetornaVazioQuandoNaoHaJson(t *testing.T) {
	if resultado := ExtrairJSON("isso nao tem json nenhum"); resultado != nil {
		t.Fatalf("esperado nil, obtido %v", resultado)
	}
}

func TestRetornaVazioParaJsonMalformadoSemLancarExcecao(t *testing.T) {
	if resultado := ExtrairJSON(`{"a": invalido}`); resultado != nil {
		t.Fatalf("esperado nil, obtido %v", resultado)
	}
}
