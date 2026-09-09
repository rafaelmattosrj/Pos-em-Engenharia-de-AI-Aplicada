// Package cleaning porta dataset-cleaning-balancing-tool.js: geracao do
// dataset simulado, deduplicacao MinHash+LSH, balanceamento por temperatura,
// e diversidade (entropia de Shannon).
package cleaning

import (
	"strconv"
	"strings"
)

type Metadata struct {
	Caso  string
	Fonte string
	ID    string
}

type Exemplo struct {
	Instrucao string
	Entrada   string
	Saida     map[string]any
	Metadata  Metadata
}

func (e Exemplo) ComMetadata(m Metadata) Exemplo {
	return Exemplo{Instrucao: e.Instrucao, Entrada: e.Entrada, Saida: e.Saida, Metadata: m}
}

type templateAuto func(nome, placa, data, valor string) string
type templateSaude func(nome, procedimento, data, valor string) string

var templatesAuto = map[string]templateAuto{
	"Oficina Estrela": func(nome, placa, data, valor string) string {
		return "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA CNPJ 12.345.678/0001-90 Rua das Turbinas 450 Distrito Industrial " +
			"Segurado: " + nome + " Placa do veiculo: " + placa + " Data do sinistro: " + data +
			" Descricao do servico: reparo de lataria e pintura no para-choque dianteiro Valor total do reparo: R$ " + valor
	},
	"Auto Center Silva": func(nome, placa, data, valor string) string {
		return "AUTO CENTER SILVA - FUNILARIA E PINTURA - CNPJ 98.765.432/0001-11 Av. dos Mecanicos 220 " +
			"Cliente/Segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data +
			" Servico executado: troca de para-lama e revisao de suspensao dianteira Valor: R$ " + valor
	},
	"Funilaria Rio Bonito": func(nome, placa, data, valor string) string {
		return "FUNILARIA RIO BONITO ME CNPJ 45.111.222/0001-33 Rua Rio Bonito 88 " +
			"Nome do segurado: " + nome + " Placa do veiculo: " + placa + " Data: " + data +
			" Orcamento: substituicao de parachoque traseiro e polimento Valor total: R$ " + valor
	},
	"Oficina Nova Alianca": func(nome, placa, data, valor string) string {
		return "OFICINA NOVA ALIANCA LTDA CNPJ 22.333.444/0001-55 Estrada Velha 1200 " +
			"Segurado: " + nome + " Placa do carro: " + placa + " Data do orcamento: " + data +
			" Descricao: reparo de amassado na porta dianteira Valor cobrado: R$ " + valor
	},
}

var templatesSaude = map[string]templateSaude{
	"Clinica Vitalis": func(nome, procedimento, data, valor string) string {
		return "CLINICA VITALIS SAUDE OCUPACIONAL CNPJ 33.222.111/0001-44 Av. Paulista 900 " +
			"Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data +
			" Valor cobrado: R$ " + valor
	},
	"Hospital Santa Clara": func(nome, procedimento, data, valor string) string {
		return "HOSPITAL SANTA CLARA CNPJ 66.555.444/0001-22 Rua das Acacias 310 " +
			"Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data +
			" Valor total: R$ " + valor
	},
	"Centro Medico Bem Estar": func(nome, procedimento, data, valor string) string {
		return "CENTRO MEDICO BEM ESTAR CNPJ 77.888.999/0001-66 Rua da Saude 45 " +
			"Nome do beneficiario: " + nome + " Procedimento: " + procedimento + " Data da consulta: " + data +
			" Valor cobrado: R$ " + valor
	},
}

var nomes = []string{
	"Marcos Vinicius Andrade Pereira", "Fernanda Costa Ribeiro", "Joaquim Pedro Salgado",
	"Beatriz Nogueira Lima", "Rafael Augusto Teixeira", "Camila dos Santos Farias",
	"Eduardo Henrique Barros", "Larissa Martins Cardoso", "Thiago Moreira Duarte",
	"Patricia Alves Monteiro", "Bruno Cesar Figueiredo", "Juliana Rocha Pimentel",
	"Gustavo Henrique Vasconcelos", "Renata Souza Albuquerque", "Diego Fernandes Castro",
	"Mariana Lopes Guimaraes",
}
var placas = []string{
	"QJK-4F82", "RTL-9921", "MNB-3310", "PLW-7765", "ZXC-2298", "BVN-6641",
	"TYU-1183", "GHJ-5529", "FDS-8842", "LKM-3376", "OIU-9954", "CVB-1120",
	"ASD-6673", "WER-4481", "XSW-2290", "POI-7738",
}
var procedimentos = []string{
	"consulta cardiologica", "exame de sangue completo", "fisioterapia ortopedica",
	"consulta ortopedica", "exame de imagem (ressonancia)", "consulta psiquiatrica",
	"sessao de fonoaudiologia", "exame oftalmologico", "consulta dermatologica",
	"exame de densitometria ossea", "consulta nutricional", "sessao de acupuntura",
}
var valores = []string{
	"3.210,50", "1.870,00", "5.640,00", "2.430,75", "890,00", "4.120,30",
	"1.250,00", "3.980,60", "2.760,00", "6.310,90", "1.540,00", "2.990,25",
	"3.450,00", "1.780,50", "4.560,00", "2.220,80",
}
var datas = []string{
	"12/03/2026", "02/04/2026", "18/05/2026", "25/03/2026", "09/04/2026", "30/04/2026",
	"14/03/2026", "21/05/2026", "05/04/2026", "11/05/2026", "28/03/2026", "16/04/2026",
}

func parseValorBr(valor string) float64 {
	limpo := strings.ReplaceAll(valor, ".", "")
	limpo = strings.ReplaceAll(limpo, ",", ".")
	v, _ := strconv.ParseFloat(limpo, 64)
	return v
}

func GerarExemplo(caso, fonte string, indice int) Exemplo {
	nome := nomes[indice%len(nomes)]
	data := datas[indice%len(datas)]
	valor := valores[indice%len(valores)]

	var entrada, instrucao string
	saida := map[string]any{}

	if caso == "amplitude-auto" {
		placa := placas[indice%len(placas)]
		entrada = templatesAuto[fonte](nome, placa, data, valor)
		saida["segurado"] = nome
		saida["placa"] = placa
		saida["valor"] = parseValorBr(valor)
		instrucao = "Extraia segurado, placa e valor do orcamento de oficina abaixo."
	} else {
		procedimento := procedimentos[indice%len(procedimentos)]
		entrada = templatesSaude[fonte](nome, procedimento, data, valor)
		saida["beneficiario"] = nome
		saida["procedimento"] = procedimento
		saida["valor"] = parseValorBr(valor)
		instrucao = "Extraia beneficiario, procedimento e valor do recibo medico abaixo."
	}

	id := caso + "-" + fonte + "-" + strconv.Itoa(indice)
	return Exemplo{Instrucao: instrucao, Entrada: entrada, Saida: saida, Metadata: Metadata{Caso: caso, Fonte: fonte, ID: id}}
}

// GerarDatasetSimulado gera o mesmo dataset simulado do original, com 2
// quase-duplicatas plantadas por caso (reenvio identico + ruido de OCR).
func GerarDatasetSimulado() []Exemplo {
	var exemplos []Exemplo

	for i := 0; i < 16; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-auto", "Oficina Estrela", i))
	}
	for i := 0; i < 5; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-auto", "Auto Center Silva", i+20))
	}
	for i := 0; i < 4; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-auto", "Funilaria Rio Bonito", i+30))
	}
	for i := 0; i < 3; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-auto", "Oficina Nova Alianca", i+40))
	}

	exemplos[1] = GerarExemplo("amplitude-auto", "Oficina Estrela", 0).
		ComMetadata(Metadata{Caso: "amplitude-auto", Fonte: "Oficina Estrela", ID: "amplitude-auto-Oficina Estrela-0-reenviado"})

	base := GerarExemplo("amplitude-auto", "Oficina Estrela", 2)
	ocrRuido := Exemplo{
		Instrucao: base.Instrucao,
		Entrada:   strings.Replace(base.Entrada, "Placa do veiculo:", "P1aca do veicu1o:", 1),
		Saida:     base.Saida,
		Metadata:  Metadata{Caso: "amplitude-auto", Fonte: "Oficina Estrela", ID: "amplitude-auto-Oficina Estrela-2-ruido-ocr"},
	}
	exemplos[3] = ocrRuido

	for i := 0; i < 12; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-saude-empresarial", "Clinica Vitalis", i))
	}
	for i := 0; i < 4; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-saude-empresarial", "Hospital Santa Clara", i+20))
	}
	for i := 0; i < 3; i++ {
		exemplos = append(exemplos, GerarExemplo("amplitude-saude-empresarial", "Centro Medico Bem Estar", i+30))
	}

	idxVitalis0 := -1
	for i, e := range exemplos {
		if e.Metadata.Fonte == "Clinica Vitalis" && strings.HasSuffix(e.Metadata.ID, "-0") {
			idxVitalis0 = i
			break
		}
	}
	exemplos[idxVitalis0+1] = GerarExemplo("amplitude-saude-empresarial", "Clinica Vitalis", 0).
		ComMetadata(Metadata{Caso: "amplitude-saude-empresarial", Fonte: "Clinica Vitalis", ID: "amplitude-saude-empresarial-Clinica Vitalis-0-reenviado"})

	return exemplos
}
