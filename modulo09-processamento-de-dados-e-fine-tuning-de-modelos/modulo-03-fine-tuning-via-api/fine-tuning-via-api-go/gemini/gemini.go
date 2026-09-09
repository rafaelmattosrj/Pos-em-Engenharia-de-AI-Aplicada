// Package gemini converte o schema canonico (instrucao/entrada/saida/metadata)
// pro formato de fine-tuning supervisionado da Vertex AI (contents/role/parts).
// Porte de dataset-upload-and-tracking-tool.js (Modulo 3.2) e
// finetuning-automation-tool.js (Modulo 3.4).
package gemini

import (
	"encoding/json"
	"errors"
	"strings"
)

type Turno struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type Conversao struct {
	Contents []Turno `json:"contents"`
}

// Converter espera saida como um mapa estruturado (objeto), serializado como
// JSON no turno do modelo -- equivalente ao caso Amplitude (Modulos 3.2/3.4).
func Converter(instrucao, entrada string, saida map[string]any) (Conversao, error) {
	if strings.TrimSpace(instrucao) == "" {
		return Conversao{}, errors.New("exemplo sem instrucao valida")
	}
	if strings.TrimSpace(entrada) == "" {
		return Conversao{}, errors.New("exemplo sem entrada valida")
	}
	if saida == nil {
		return Conversao{}, errors.New("exemplo sem saida valida")
	}
	textoSaida, err := json.Marshal(saida)
	if err != nil {
		return Conversao{}, err
	}
	return Conversao{Contents: []Turno{
		{Role: "user", Text: instrucao + "\n\n" + entrada},
		{Role: "model", Text: string(textoSaida)},
	}}, nil
}

// ConverterTexto e' a variante do extra Dolly-15k: a saida e' texto solto,
// nao objeto estruturado, entao nao passa por json.Marshal (senao a resposta
// esperada sairia com aspas extras).
func ConverterTexto(instrucao, entrada, saidaTexto string) (Conversao, error) {
	if strings.TrimSpace(instrucao) == "" {
		return Conversao{}, errors.New("instrucao obrigatoria")
	}
	if strings.TrimSpace(entrada) == "" {
		return Conversao{}, errors.New("entrada obrigatoria")
	}
	if strings.TrimSpace(saidaTexto) == "" {
		return Conversao{}, errors.New("saida (texto) obrigatoria")
	}
	return Conversao{Contents: []Turno{
		{Role: "user", Text: instrucao + "\n\n" + entrada},
		{Role: "model", Text: saidaTexto},
	}}, nil
}
