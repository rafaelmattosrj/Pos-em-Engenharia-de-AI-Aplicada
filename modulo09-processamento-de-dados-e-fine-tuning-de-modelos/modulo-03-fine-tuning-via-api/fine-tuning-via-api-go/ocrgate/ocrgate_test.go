package ocrgate

import "testing"

func f(v float64) *float64 { return &v }

func TestQuatroDocumentosReaisPassamNoLimiarPadrao(t *testing.T) {
	exemplos := []Exemplo{
		{ID: "doc-auto-1", ConfiancaOcr: f(0.943)},
		{ID: "doc-auto-2", ConfiancaOcr: f(0.958)},
		{ID: "doc-saude-1", ConfiancaOcr: f(0.957)},
		{ID: "doc-saude-2", ConfiancaOcr: f(0.959)},
	}
	r := Filtrar(exemplos)
	if len(r.AprovadosPorOcr) != 4 || len(r.SinalizadosParaRevisao) != 0 {
		t.Fatalf("resultado inesperado: %+v", r)
	}
}

func TestExemploAbaixoDoLimiarESinalizado(t *testing.T) {
	r := Filtrar([]Exemplo{{ID: "doc-degradado", ConfiancaOcr: f(0.62)}})
	if len(r.SinalizadosParaRevisao) != 1 || len(r.Aprovados) != 0 {
		t.Fatalf("resultado inesperado: %+v", r)
	}
}

func TestExemploSemConfiancaSegueAprovadoSemGate(t *testing.T) {
	r := Filtrar([]Exemplo{{ID: "sintetico"}})
	if len(r.SemConfianca) != 1 || len(r.AprovadosPorOcr) != 0 || len(r.Aprovados) != 1 {
		t.Fatalf("resultado inesperado: %+v", r)
	}
}

func TestConfiancaIgualAoLimiarEhAprovada(t *testing.T) {
	r := Filtrar([]Exemplo{{ID: "x", ConfiancaOcr: f(LimiarPadrao)}})
	if len(r.AprovadosPorOcr) != 1 {
		t.Fatalf("resultado inesperado: %+v", r)
	}
}

func TestLimiarCustomizadoEhRespeitado(t *testing.T) {
	e := Exemplo{ID: "x", ConfiancaOcr: f(0.7)}
	rPadrao := Filtrar([]Exemplo{e})
	rFrouxo := Filtrar([]Exemplo{e}, 0.6)
	if len(rPadrao.SinalizadosParaRevisao) != 1 || len(rFrouxo.AprovadosPorOcr) != 1 {
		t.Fatalf("limiar customizado nao respeitado: %+v %+v", rPadrao, rFrouxo)
	}
}
