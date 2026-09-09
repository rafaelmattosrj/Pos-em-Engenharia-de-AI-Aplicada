package extraction

import (
	"strings"
	"testing"
)

func TestParseiaValorFormatoBrasileiroComMilhar(t *testing.T) {
	if v := ParsearValorBRL("3.210,50"); v != 3210.5 {
		t.Fatalf("esperado 3210.5, obtive %v", v)
	}
}

func TestParseiaValorComPrefixoRS(t *testing.T) {
	if v := ParsearValorBRL("R$ 145,90"); v != 145.9 {
		t.Fatalf("esperado 145.9, obtive %v", v)
	}
}

func TestParseiaOrcamentoAutoTolerandoRotuloVariavel(t *testing.T) {
	textoOcr := "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA\nSegurado: Marcos Vinicius Andrade Pereira\n" +
		"Placa do veiculo: QJK-4F82\nValor total do reparo: R$ 3.210,50"
	campos := ParsearOrcamentoAuto(textoOcr)
	if campos.Segurado != "Marcos Vinicius Andrade Pereira" || campos.Placa != "QJK-4F82" || campos.Valor == nil || *campos.Valor != 3210.50 {
		t.Fatalf("obtive %+v", campos)
	}
}

func TestParseiaReciboSaude(t *testing.T) {
	textoOcr := "CLINICA VITALIS SAUDE OCUPACIONAL\nBeneficiario: Carlos Eduardo Martins\n" +
		"Procedimento: Consulta Cardiologica\nValor cobrado: R$ 380,00"
	campos := ParsearReciboSaude(textoOcr)
	if campos.Beneficiario != "Carlos Eduardo Martins" || campos.Procedimento != "Consulta Cardiologica" || campos.Valor == nil || *campos.Valor != 380.00 {
		t.Fatalf("obtive %+v", campos)
	}
}

func TestExemploSemCampoObrigatorioEInvalido(t *testing.T) {
	saida := map[string]any{"segurado": "Fulano", "placa": nil, "valor": 100.0}
	v := ValidarExemplo("amplitude-auto", saida)
	if v.Valido {
		t.Fatal("esperado invalido")
	}
	achou := false
	for _, e := range v.Erros {
		if strings.Contains(e, "placa") {
			achou = true
		}
	}
	if !achou {
		t.Fatalf("esperado erro mencionando placa, obtive %v", v.Erros)
	}
}

func TestExemploComValorZeroEInvalido(t *testing.T) {
	saida := map[string]any{"segurado": "Fulano", "placa": "ABC-1234", "valor": 0.0}
	v := ValidarExemplo("amplitude-auto", saida)
	if v.Valido {
		t.Fatal("esperado invalido para valor zero")
	}
}
