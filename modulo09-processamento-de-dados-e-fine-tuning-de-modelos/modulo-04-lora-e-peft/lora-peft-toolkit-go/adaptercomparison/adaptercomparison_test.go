package adaptercomparison

import (
	"path/filepath"
	"testing"
)

const saidaRealSemAdaptador = `==========
<|channel>thought
Here's a thinking process to extract the requested information:

1.  **Analyze the Request:** The user wants to extract three specific pieces of information from the provided text (a medical receipt/summary):
    *   Beneficiário (Beneficiary)
    *   Procedimento (Procedure)
    *   Valor (Value/Amount)

2.  **
==========
Prompt: 119 tokens, 119.302 tokens-per-sec
Generation: 80 tokens, 52.503 tokens-per-sec
Peak memory: 9.406 GB`

const saidaRealComAdaptador = `==========
{"beneficiario":"Felipe Alves Monteiro","procedimento":"consulta de clinica geral","valor":3450}
==========
Prompt: 119 tokens, 161.590 tokens-per-sec
Generation: 28 tokens, 43.129 tokens-per-sec
Peak memory: 9.406 GB`

func testJSONL() string {
	return filepath.Join("..", "..", "mlx-data", "test.jsonl")
}

func TestCarregaExemploRealIndice8(t *testing.T) {
	exemplo, err := CarregarExemploTeste(testJSONL(), IndiceExemplo)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if exemplo.Gabarito["beneficiario"] != "Felipe Alves Monteiro" {
		t.Fatalf("beneficiario=%v", exemplo.Gabarito["beneficiario"])
	}
	if exemplo.Gabarito["procedimento"] != "consulta de clinica geral" {
		t.Fatalf("procedimento=%v", exemplo.Gabarito["procedimento"])
	}
}

func TestMontaArgumentosSemAdapterPath(t *testing.T) {
	args := MontarArgumentosGenerate("x", "", MaxTokens, ModeloBase)
	for _, a := range args {
		if a == "--adapter-path" {
			t.Fatal("não deveria conter --adapter-path")
		}
	}
}

func TestMontaArgumentosComAdapterPath(t *testing.T) {
	args := MontarArgumentosGenerate("x", "/caminho/adapters", MaxTokens, ModeloBase)
	achou := false
	for i, a := range args {
		if a == "--adapter-path" {
			achou = true
			if args[i+1] != "/caminho/adapters" {
				t.Fatalf("valor esperado /caminho/adapters, obtido %v", args[i+1])
			}
		}
	}
	if !achou {
		t.Fatal("esperava --adapter-path nos argumentos")
	}
}

func TestParseiaSaidaSemAdaptador(t *testing.T) {
	r, err := ParsearSaidaGenerate(saidaRealSemAdaptador)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if r.TokensGerados != 80 {
		t.Fatalf("tokens=%d, esperado 80", r.TokensGerados)
	}
	if !r.BateuNoLimiteDeTokens {
		t.Fatal("esperava bateuNoLimiteDeTokens = true")
	}
	if r.JSON != nil {
		t.Fatal("esperava json nil")
	}
}

func TestParseiaSaidaComAdaptador(t *testing.T) {
	r, err := ParsearSaidaGenerate(saidaRealComAdaptador)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if r.TokensGerados != 28 {
		t.Fatalf("tokens=%d, esperado 28", r.TokensGerados)
	}
	if r.BateuNoLimiteDeTokens {
		t.Fatal("esperava bateuNoLimiteDeTokens = false")
	}
	if r.JSON["beneficiario"] != "Felipe Alves Monteiro" {
		t.Fatalf("beneficiario=%v", r.JSON["beneficiario"])
	}
}

func TestCompararComGabaritoBate(t *testing.T) {
	r, _ := ParsearSaidaGenerate(saidaRealComAdaptador)
	gabarito := map[string]any{"beneficiario": "Felipe Alves Monteiro", "procedimento": "consulta de clinica geral", "valor": float64(3450)}
	if !CompararComGabarito(r.JSON, gabarito) {
		t.Fatal("esperava bater com o gabarito")
	}
}

func TestCompararComGabaritoFalhaSemJson(t *testing.T) {
	r, _ := ParsearSaidaGenerate(saidaRealSemAdaptador)
	gabarito := map[string]any{"beneficiario": "Felipe Alves Monteiro"}
	if CompararComGabarito(r.JSON, gabarito) {
		t.Fatal("não deveria bater, json é nil")
	}
}

func TestCarregarExemploIndiceForaDoIntervalo(t *testing.T) {
	_, err := CarregarExemploTeste(testJSONL(), 999)
	if err == nil {
		t.Fatal("esperava erro para índice fora do intervalo")
	}
}
