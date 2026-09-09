// Package pii porta pii-scrubbing-gate-tool.js: gate de higienizacao de PII
// antes que um documento entre no dataset de fine-tuning. Abordagem hibrida
// (regex + heuristica), a mesma usada em pipelines reais de producao
// (Microsoft Presidio combina regex + NER).
//   - CPF: regex de formato + validacao real do digito verificador (Modulo 11).
//   - Nome: ancora de rotulo (Segurado:/Beneficiario:) mais confiavel.
package pii

import (
	"regexp"
	"strconv"
	"strings"
)

var regexCPF = regexp.MustCompile(`\b\d{3}\.?\d{3}\.?\d{3}-?\d{2}\b`)

func calcularDigitoVerificador(parcial string) int {
	soma := 0
	peso := len(parcial) + 1
	for _, c := range parcial {
		d, _ := strconv.Atoi(string(c))
		soma += d * peso
		peso--
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

var naoDigito = regexp.MustCompile(`\D`)

func ValidarCPF(cpfComOuSemMascara string) bool {
	digitos := naoDigito.ReplaceAllString(cpfComOuSemMascara, "")
	if len(digitos) != 11 {
		return false
	}
	todosIguais := true
	for i := 1; i < len(digitos); i++ {
		if digitos[i] != digitos[0] {
			todosIguais = false
			break
		}
	}
	if todosIguais {
		return false
	}
	d1 := calcularDigitoVerificador(digitos[:9])
	d2 := calcularDigitoVerificador(digitos[:9] + strconv.Itoa(d1))
	return digitos[9] == byte('0'+d1) && digitos[10] == byte('0'+d2)
}

var regexNomeAncorado = regexp.MustCompile(`(?i)(Segurado|Beneficiário|Beneficiario|Nome do segurado)[ \t]*:[ \t]*([A-ZÀ-Ú][\wÀ-ú]*(?:[ \t]+[A-ZÀ-Ú][\wÀ-ú]*){1,4})`)

type NomeEncontrado struct {
	Nome      string
	Confianca string
	Metodo    string
}

func DetectarNomesAncorados(texto string) []NomeEncontrado {
	var encontrados []NomeEncontrado
	for _, m := range regexNomeAncorado.FindAllStringSubmatch(texto, -1) {
		encontrados = append(encontrados, NomeEncontrado{Nome: m[2], Confianca: "alta", Metodo: "ancora_rotulo"})
	}
	return encontrados
}

type CpfEncontrado struct {
	Texto  string
	Valido bool
}

type ResultadoVarredura struct {
	CpfsEncontrados  []CpfEncontrado
	NomesEncontrados []NomeEncontrado
	TextoRedigido    string
	TotalPiiRedigido int
}

func VarrerPII(textoDocumento string) ResultadoVarredura {
	var cpfsEncontrados []CpfEncontrado
	for _, texto := range regexCPF.FindAllString(textoDocumento, -1) {
		cpfsEncontrados = append(cpfsEncontrados, CpfEncontrado{Texto: texto, Valido: ValidarCPF(texto)})
	}

	nomesEncontrados := DetectarNomesAncorados(textoDocumento)

	textoRedigido := textoDocumento
	for _, cpf := range cpfsEncontrados {
		if cpf.Valido {
			textoRedigido = strings.ReplaceAll(textoRedigido, cpf.Texto, "[CPF_REDIGIDO]")
		}
	}
	for _, nome := range nomesEncontrados {
		textoRedigido = strings.ReplaceAll(textoRedigido, nome.Nome, "[NOME_REDIGIDO]")
	}

	cpfsValidos := 0
	for _, cpf := range cpfsEncontrados {
		if cpf.Valido {
			cpfsValidos++
		}
	}

	return ResultadoVarredura{
		CpfsEncontrados: cpfsEncontrados, NomesEncontrados: nomesEncontrados,
		TextoRedigido: textoRedigido, TotalPiiRedigido: cpfsValidos + len(nomesEncontrados),
	}
}
