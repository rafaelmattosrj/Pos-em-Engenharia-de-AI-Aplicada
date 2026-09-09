package extraction

import "testing"

func TestNormalizarRemoveAcentoCaixaEEspacosExtras(t *testing.T) {
	if v := Normalizar("  José  DA Silva "); v != "jose da silva" {
		t.Fatalf("esperado 'jose da silva', obtive %q", v)
	}
}

func TestCompararComEsperadoBateParaValoresIguais(t *testing.T) {
	extraido := map[string]any{"segurado": "Marcos Vinicius Andrade Pereira", "placa": "QJK-4F82", "valor": 3210.50}
	esperado := map[string]any{"segurado": "Marcos Vinicius Andrade Pereira", "placa": "QJK-4F82", "valor": 3210.50}
	r := CompararComEsperado(extraido, esperado)
	if r.Acertos != 3 || r.Total != 3 {
		t.Fatalf("esperado 3/3, obtive %d/%d", r.Acertos, r.Total)
	}
}

func TestCompararComEsperadoToleraAcentoECaixa(t *testing.T) {
	extraido := map[string]any{"beneficiario": "carlos eduardo martins"}
	esperado := map[string]any{"beneficiario": "Carlos Eduardo Martins"}
	r := CompararComEsperado(extraido, esperado)
	if r.Acertos != 1 {
		t.Fatalf("esperado 1 acerto, obtive %d", r.Acertos)
	}
}

func TestCompararComEsperadoDetectaDivergenciaDeValor(t *testing.T) {
	extraido := map[string]any{"valor": 100.0}
	esperado := map[string]any{"valor": 200.0}
	r := CompararComEsperado(extraido, esperado)
	if r.Acertos != 0 {
		t.Fatalf("esperado 0 acertos, obtive %d", r.Acertos)
	}
}
