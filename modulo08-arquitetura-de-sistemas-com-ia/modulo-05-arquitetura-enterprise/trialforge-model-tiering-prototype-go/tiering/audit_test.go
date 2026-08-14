package tiering

import (
	"path/filepath"
	"testing"
)

func TestRegistrar_GravaLinhaJsonComTimestamp(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "audit-trail-tiering.jsonl")
	at := NewAuditTrail(caminho)

	if err := at.Registrar(map[string]any{"estudoId": "estudo-A", "status_final": "aprovado"}); err != nil {
		t.Fatal(err)
	}

	linhas, err := at.LerTodas()
	if err != nil {
		t.Fatal(err)
	}
	if len(linhas) != 1 {
		t.Fatalf("esperava 1 linha, obteve %d", len(linhas))
	}
	if linhas[0]["timestamp"] == nil || linhas[0]["timestamp"] == "" {
		t.Error("esperava timestamp preenchido")
	}
	if linhas[0]["estudoId"] != "estudo-A" {
		t.Errorf("esperava estudoId=estudo-A, obteve %v", linhas[0]["estudoId"])
	}
}

func TestRegistrar_EhAppendOnly(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "audit-trail-tiering.jsonl")
	at := NewAuditTrail(caminho)

	at.Registrar(map[string]any{"n": 1.0})
	at.Registrar(map[string]any{"n": 2.0})
	at.Registrar(map[string]any{"n": 3.0})

	linhas, err := at.LerTodas()
	if err != nil {
		t.Fatal(err)
	}
	if len(linhas) != 3 {
		t.Fatalf("esperava 3 linhas, obteve %d", len(linhas))
	}
	if linhas[0]["n"] != 1.0 || linhas[2]["n"] != 3.0 {
		t.Errorf("ordem inesperada: %v", linhas)
	}
}

func TestLerTodas_ArquivoInexistente_RetornaListaVazia(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "nao-existe.jsonl")
	at := NewAuditTrail(caminho)

	linhas, err := at.LerTodas()
	if err != nil {
		t.Fatal(err)
	}
	if len(linhas) != 0 {
		t.Errorf("esperava lista vazia, obteve %d linhas", len(linhas))
	}
}
