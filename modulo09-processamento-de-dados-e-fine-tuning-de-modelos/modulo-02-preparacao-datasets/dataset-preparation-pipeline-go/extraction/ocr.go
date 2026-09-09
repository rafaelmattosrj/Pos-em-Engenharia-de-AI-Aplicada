// Package extraction porta extraction-to-jsonl-tool.js (OCR + parsing +
// validacao de schema) e extracao-llm-multimodal-tool.js (Vertex AI/Gemini).
//
// OcrTexto/OcrConfiancaMedia chamam o binario `tesseract` real via
// os/exec.Command (equivalente a execFileSync no original). Requer
// Tesseract instalado com o pacote de idioma "por" - nao roda em
// CI/teste automatizado sem essa dependencia externa. As funcoes de
// parsing e validacao de schema sao puras e testadas isoladamente.
package extraction

import (
	"bufio"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// OcrTexto roda o Tesseract em modo texto puro contra a imagem, em portugues.
func OcrTexto(caminhoImagem string) (string, error) {
	out, err := exec.Command("tesseract", caminhoImagem, "stdout", "-l", "por").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// OcrConfiancaMedia roda o Tesseract em modo TSV para extrair a confianca
// media (0.0-1.0) das palavras reconhecidas.
func OcrConfiancaMedia(caminhoImagem string) (float64, error) {
	out, err := exec.Command("tesseract", caminhoImagem, "stdout", "-l", "por", "tsv").Output()
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Scan() // pula cabecalho
	soma, total := 0.0, 0
	for scanner.Scan() {
		campos := strings.Split(scanner.Text(), "\t")
		if len(campos) >= 12 {
			conf, err := strconv.ParseFloat(campos[10], 64)
			if err == nil && conf >= 0 {
				soma += conf / 100
				total++
			}
		}
	}
	if total == 0 {
		return 0, nil
	}
	return soma / float64(total), nil
}

// ParsearValorBRL converte valor no formato brasileiro ("3.210,50") para numero (3210.50).
func ParsearValorBRL(texto string) float64 {
	limpo := regexp.MustCompile(`[^\d.,]`).ReplaceAllString(texto, "")
	limpo = strings.ReplaceAll(limpo, ".", "")
	limpo = strings.ReplaceAll(limpo, ",", ".")
	v, _ := strconv.ParseFloat(limpo, 64)
	return v
}

func extrairCampo(texto string, padroes []*regexp.Regexp) string {
	for _, padrao := range padroes {
		m := padrao.FindStringSubmatch(texto)
		if m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

var (
	regexSegurado   = regexp.MustCompile(`(?i)(?:nome do )?segurado:?\s*(.+)`)
	regexPlaca      = regexp.MustCompile(`(?i)placa(?:\s+do\s+ve[ií]culo)?:?\s*([A-Z0-9\-]+)`)
	regexValorAuto  = regexp.MustCompile(`(?i)(?:valor total do reparo|total):?\s*r\$?\s*([\d.,]+)`)
	regexBenef      = regexp.MustCompile(`(?i)(?:paciente/)?benefici[aá]rio:?\s*(.+)`)
	regexProced     = regexp.MustCompile(`(?i)procedimento(?:\s+realizado)?:?\s*(.+)`)
	regexValorSaude = regexp.MustCompile(`(?i)valor(?:\s+cobrado)?:?\s*r\$?\s*([\d.,]+)`)
)

type CamposAuto struct {
	Segurado string
	Placa    string
	Valor    *float64
}

func ParsearOrcamentoAuto(textoOcr string) CamposAuto {
	segurado := extrairCampo(textoOcr, []*regexp.Regexp{regexSegurado})
	placa := extrairCampo(textoOcr, []*regexp.Regexp{regexPlaca})
	valorTexto := extrairCampo(textoOcr, []*regexp.Regexp{regexValorAuto})
	var valor *float64
	if valorTexto != "" {
		v := ParsearValorBRL(valorTexto)
		valor = &v
	}
	return CamposAuto{Segurado: segurado, Placa: placa, Valor: valor}
}

type CamposSaude struct {
	Beneficiario string
	Procedimento string
	Valor        *float64
}

func ParsearReciboSaude(textoOcr string) CamposSaude {
	beneficiario := extrairCampo(textoOcr, []*regexp.Regexp{regexBenef})
	procedimento := extrairCampo(textoOcr, []*regexp.Regexp{regexProced})
	valorTexto := extrairCampo(textoOcr, []*regexp.Regexp{regexValorSaude})
	var valor *float64
	if valorTexto != "" {
		v := ParsearValorBRL(valorTexto)
		valor = &v
	}
	return CamposSaude{Beneficiario: beneficiario, Procedimento: procedimento, Valor: valor}
}

var camposObrigatorios = map[string][]string{
	"amplitude-auto":              {"segurado", "placa", "valor"},
	"amplitude-saude-empresarial": {"beneficiario", "procedimento", "valor"},
}

type Validacao struct {
	Valido bool
	Erros  []string
}

// ValidarExemplo valida se um exemplo estruturado (mapa campo->valor) tem
// todos os campos obrigatorios do seu caso, sem nulo (string vazia/nil) nem valor invalido.
func ValidarExemplo(caso string, saida map[string]any) Validacao {
	var erros []string
	for _, campo := range camposObrigatorios[caso] {
		v, ok := saida[campo]
		if !ok || v == nil {
			erros = append(erros, "campo obrigatorio ausente: "+campo)
			continue
		}
		if s, isStr := v.(string); isStr && s == "" {
			erros = append(erros, "campo obrigatorio ausente: "+campo)
		}
	}
	if valor, ok := saida["valor"]; ok && valor != nil {
		if v, isFloat := valor.(float64); isFloat {
			if math.IsNaN(v) || v <= 0 {
				erros = append(erros, "valor invalido: "+strconv.FormatFloat(v, 'f', -1, 64))
			}
		}
	}
	return Validacao{Valido: len(erros) == 0, Erros: erros}
}
