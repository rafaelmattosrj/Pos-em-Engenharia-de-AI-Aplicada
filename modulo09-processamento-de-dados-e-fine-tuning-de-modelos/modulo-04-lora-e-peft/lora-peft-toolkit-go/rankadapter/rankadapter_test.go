package rankadapter

import "testing"

const saidaSemAdaptador = `==========
<|channel>thought
Here's a thinking process to extract the requested information:

1.  **Analyze the Request:** The user wants to extract three specific pieces of information from the provided text (an auto repair quote/budget):
==========
Prompt: 155 tokens, 211.031 tokens-per-sec
Generation: 80 tokens, 50.460 tokens-per-sec
Peak memory: 9.447 GB`

const saidaRank4 = `==========
{"segurado":"Ricardo Alves Monteiro","placa":"JBR-9021","valor":2820}
==========
Prompt: 155 tokens, 218.733 tokens-per-sec
Generation: 27 tokens, 41.540 tokens-per-sec
Peak memory: 9.447 GB`

func TestSemAdaptadorBateNoLimiteNaoChegaAJson(t *testing.T) {
	r, err := ParsearSaida(saidaSemAdaptador)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if r.Tokens != 80 {
		t.Fatalf("tokens=%d, esperado 80", r.Tokens)
	}
	if r.JSON != nil {
		t.Fatal("esperava json nil")
	}
}

func TestRank4JsonExatoBatendoComGabarito(t *testing.T) {
	r, err := ParsearSaida(saidaRank4)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if r.JSON["segurado"] != "Ricardo Alves Monteiro" || r.JSON["placa"] != "JBR-9021" {
		t.Fatalf("json=%v", r.JSON)
	}
}

func TestBaterComGabaritoConfirmaERejeita(t *testing.T) {
	r, _ := ParsearSaida(saidaRank4)
	if !BaterComGabarito(r.JSON) {
		t.Fatal("esperava bater com o gabarito")
	}
	if BaterComGabarito(map[string]any{"segurado": "outro"}) {
		t.Fatal("não deveria bater")
	}
}

func TestAdaptersRetornaOsTresRanks(t *testing.T) {
	a := Adapters("/base")
	if a["rank 4"] != "/base/mlx-adapters-rank4" || a["rank 8"] != "/base/mlx-adapters" || a["rank 16"] != "/base/mlx-adapters-rank16" {
		t.Fatalf("adapters=%v", a)
	}
}

func TestMontaArgumentosSemAdapterPath(t *testing.T) {
	args := MontarArgumentos("")
	for _, arg := range args {
		if arg == "--adapter-path" {
			t.Fatal("não deveria conter --adapter-path")
		}
	}
}
