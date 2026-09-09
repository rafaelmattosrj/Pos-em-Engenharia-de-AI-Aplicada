package pii

import (
	"strings"
	"testing"
)

func TestCpfDeTesteConhecidoEValido(t *testing.T) {
	if !ValidarCPF("111.444.777-35") {
		t.Fatal("esperado CPF valido")
	}
}

func TestTrocarUltimoDigitoInvalidaCpf(t *testing.T) {
	if ValidarCPF("111.444.777-34") {
		t.Fatal("esperado CPF invalido")
	}
}

func TestSequenciaDeDigitosRepetidosNuncaEValida(t *testing.T) {
	if ValidarCPF("000.000.000-00") {
		t.Fatal("esperado invalido para digitos repetidos")
	}
}

func TestCpfSemMascaraValidaIgual(t *testing.T) {
	if !ValidarCPF("11144477735") {
		t.Fatal("esperado valido sem mascara")
	}
}

func TestDigitosAleatoriosSaoRejeitados(t *testing.T) {
	if ValidarCPF("123.456.789-00") {
		t.Fatal("esperado rejeitado")
	}
}

func TestExtraiNomeDepoisDeSegurado(t *testing.T) {
	encontrados := DetectarNomesAncorados("Segurado: Marcos Vinicius Andrade Pereira\nPlaca: ABC-1234")
	if len(encontrados) != 1 || encontrados[0].Nome != "Marcos Vinicius Andrade Pereira" {
		t.Fatalf("obtive %+v", encontrados)
	}
}

func TestExtraiNomeDepoisDeBeneficiarioComAcento(t *testing.T) {
	encontrados := DetectarNomesAncorados("Beneficiário: Carlos Eduardo Martins")
	if len(encontrados) == 0 || encontrados[0].Nome != "Carlos Eduardo Martins" {
		t.Fatalf("obtive %+v", encontrados)
	}
}

func TestExtraiNomeDepoisDeBeneficiarioSemAcento(t *testing.T) {
	encontrados := DetectarNomesAncorados("Beneficiario: Carlos Eduardo Martins")
	if len(encontrados) == 0 || encontrados[0].Nome != "Carlos Eduardo Martins" {
		t.Fatalf("obtive %+v", encontrados)
	}
}

func TestTextoSemRotuloNaoGeraFalsoPositivo(t *testing.T) {
	encontrados := DetectarNomesAncorados("OFICINA ESTRELA - ORCAMENTO N. 4471")
	if len(encontrados) != 0 {
		t.Fatalf("esperado nenhum, obtive %+v", encontrados)
	}
}

func TestDocumentoRealTemNomeECpfRedigidos(t *testing.T) {
	doc := "OFICINA ESTRELA - ORCAMENTO N. 4471\nSegurado: Marcos Vinicius Andrade Pereira\n" +
		"CPF: 111.444.777-35\nPlaca do veiculo: QJK-4F82\n\nValor total do reparo: R$ 3.210,50"
	resultado := VarrerPII(doc)
	if len(resultado.NomesEncontrados) != 1 {
		t.Fatalf("esperado 1 nome, obtive %d", len(resultado.NomesEncontrados))
	}
	cpfsValidos := 0
	for _, c := range resultado.CpfsEncontrados {
		if c.Valido {
			cpfsValidos++
		}
	}
	if cpfsValidos != 1 {
		t.Fatalf("esperado 1 CPF valido, obtive %d", cpfsValidos)
	}
	if !strings.Contains(resultado.TextoRedigido, "[NOME_REDIGIDO]") || !strings.Contains(resultado.TextoRedigido, "[CPF_REDIGIDO]") {
		t.Fatalf("esperado redacao de nome e CPF, obtive: %s", resultado.TextoRedigido)
	}
	if !strings.Contains(resultado.TextoRedigido, "QJK-4F82") || !strings.Contains(resultado.TextoRedigido, "3.210,50") {
		t.Fatalf("placa e valor deveriam sobreviver a redacao: %s", resultado.TextoRedigido)
	}
}

func TestCpfComDigitoInvalidoNaoERedigido(t *testing.T) {
	doc := "Protocolo interno: 123.456.789-00\nSegurado: Ana Paula Ribeiro"
	resultado := VarrerPII(doc)
	for _, c := range resultado.CpfsEncontrados {
		if c.Valido {
			t.Fatal("nenhum CPF deveria ser valido")
		}
	}
	if strings.Contains(resultado.TextoRedigido, "[CPF_REDIGIDO]") {
		t.Fatal("nao deveria conter marcador de CPF redigido")
	}
}
