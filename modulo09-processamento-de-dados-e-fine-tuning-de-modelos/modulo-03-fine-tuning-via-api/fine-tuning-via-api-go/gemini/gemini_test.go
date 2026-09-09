package gemini

import (
	"encoding/json"
	"strings"
	"testing"
)

func exemploAuto() map[string]any {
	return map[string]any{"segurado": "Camila Costa Ribeiro", "placa": "AZS-6617", "valor": 1780.5}
}

func TestConversaoGeraDoisTurnos(t *testing.T) {
	c, err := Converter("instrucao", "entrada", exemploAuto())
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Contents) != 2 {
		t.Fatalf("esperado 2 turnos, obtido %d", len(c.Contents))
	}
	if c.Contents[0].Role != "user" || c.Contents[1].Role != "model" {
		t.Fatalf("roles incorretos: %+v", c.Contents)
	}
}

func TestTurnoUsuarioConcatenaInstrucaoEEntrada(t *testing.T) {
	c, _ := Converter("Extraia segurado", "Segurado: Camila", exemploAuto())
	if !strings.Contains(c.Contents[0].Text, "Extraia segurado") || !strings.Contains(c.Contents[0].Text, "Segurado: Camila") {
		t.Fatalf("texto do usuario incompleto: %s", c.Contents[0].Text)
	}
}

func TestTurnoModeloEhJsonExatoDaSaida(t *testing.T) {
	c, _ := Converter("x", "y", exemploAuto())
	var saida map[string]any
	if err := json.Unmarshal([]byte(c.Contents[1].Text), &saida); err != nil {
		t.Fatal(err)
	}
	if saida["segurado"] != "Camila Costa Ribeiro" || saida["placa"] != "AZS-6617" {
		t.Fatalf("saida decodificada incorreta: %+v", saida)
	}
}

func TestRejeitaExemploSemInstrucao(t *testing.T) {
	if _, err := Converter("", "y", exemploAuto()); err == nil {
		t.Fatal("esperava erro por instrucao ausente")
	}
}

func TestRejeitaExemploSemEntrada(t *testing.T) {
	if _, err := Converter("x", "", exemploAuto()); err == nil {
		t.Fatal("esperava erro por entrada ausente")
	}
}

func TestRejeitaExemploSemSaida(t *testing.T) {
	if _, err := Converter("x", "y", nil); err == nil {
		t.Fatal("esperava erro por saida ausente")
	}
}

func TestConverterTextoNaoSerializaComoJson(t *testing.T) {
	c, err := ConverterTexto("Pergunta", "Contexto", "Resposta em texto puro")
	if err != nil {
		t.Fatal(err)
	}
	if c.Contents[1].Text != "Resposta em texto puro" {
		t.Fatalf("texto inesperado: %s", c.Contents[1].Text)
	}
	if strings.HasPrefix(c.Contents[1].Text, "\"") {
		t.Fatal("nao deveria vir entre aspas de JSON")
	}
}
