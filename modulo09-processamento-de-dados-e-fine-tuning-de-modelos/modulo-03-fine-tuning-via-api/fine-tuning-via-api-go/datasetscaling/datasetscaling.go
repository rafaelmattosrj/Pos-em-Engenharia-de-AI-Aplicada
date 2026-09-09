// Package datasetscaling gera o dataset em escala de treino real (305 brutos
// -> 300 apos dedup -> 200 balanceados, 120 Amplitude Auto + 80 Amplitude
// Saude Empresarial) -- porte de m3-dataset-scaling-tool.js / DatasetScaling.java
// (Modulo 3.2). Reusa o pipeline formal do Modulo 2.2 via package minhash
// (mesma nota de adaptacao do package minhash: duplicado, nao referenciado,
// por nao haver modulo compartilhado entre projetos Maven/Go independentes
// neste repo).
package datasetscaling

import (
	"fmt"
	"strconv"
	"strings"

	"fine-tuning-via-api/minhash"
)

// ExemploDataset e' um exemplo no schema canonico (instrucao/entrada/saida +
// metadata de proveniencia).
type ExemploDataset struct {
	Instrucao string
	Entrada   string
	Saida     map[string]any
	Caso      string
	Fonte     string
	ID        string
}

// paraDedup usa so `Entrada` (nao Instrucao+Entrada) -- o mesmo criterio do
// dataset-cleaning-balancing-tool.js (Modulo 2.2), que compara so o
// documento de entrada dentro de um mesmo caso (a instrucao e' identica pra
// todo exemplo do caso, entao concatenar so infla a similaridade sem
// discriminar nada). A concatenacao instrucao+entrada e' uma correcao
// especifica do extra Dolly (package dolly), nao deste pipeline.
func (e ExemploDataset) paraDedup() minhash.Exemplo {
	return minhash.Exemplo{ID: e.ID, Caso: e.Caso, Fonte: e.Fonte, TextoParaDedup: e.Entrada}
}

var nomes = []string{
	"Marcos Vinicius Andrade Pereira", "Fernanda Costa Ribeiro", "Joaquim Pedro Salgado",
	"Beatriz Nogueira Lima", "Rafael Augusto Teixeira", "Camila dos Santos Farias",
	"Eduardo Henrique Barros", "Larissa Martins Cardoso", "Thiago Moreira Duarte",
	"Patricia Alves Monteiro", "Bruno Cesar Figueiredo", "Juliana Rocha Pimentel",
	"Gustavo Henrique Vasconcelos", "Renata Souza Albuquerque", "Diego Fernandes Castro",
	"Mariana Lopes Guimaraes", "Vinicius Almeida Correia", "Sabrina Ferreira Nunes",
	"Leonardo Batista Cavalcanti", "Priscila Andrade Melo", "Rodrigo Tavares Siqueira",
	"Amanda Cristina Peixoto", "Felipe Augusto Barbosa", "Carolina Machado Freitas",
	"Anderson Luiz Ramalho", "Vanessa Regina Coutinho", "Fabio Junior Aragao",
	"Debora Cristina Vieira", "Marcelo Souza Bittencourt", "Tatiane Pereira Godoy",
	"Alexandre Costa Miranda", "Cristiane Lopes Assuncao", "Fernando Braga Quintanilha",
	"Simone Rocha Vilaca", "Rogerio dos Santos Pena", "Michele Aparecida Fonseca",
	"Wagner Luiz Bessa", "Andreia Cristina Prado", "Cesar Augusto Nascimento",
	"Roberta Lima Sarmento", "Paulo Ricardo Andrade",
}

var placas = []string{
	"QJK-4F82", "RTL-9921", "MNB-3310", "PLW-7765", "ZXC-2298", "BVN-6641",
	"TYU-1183", "GHJ-5529", "FDS-8842", "LKM-3376", "OIU-9954", "CVB-1120",
	"ASD-6673", "WER-4481", "XSW-2290", "POI-7738", "HGF-3391", "MJU-6624",
	"NBV-1187", "KLO-5540", "ERT-8873", "YUI-2216", "CDE-9950", "VBN-4483",
	"AZS-6617", "QWE-1150", "DFG-7784", "RTY-3318", "FGH-8852", "TGB-2286",
	"YHN-5520", "UJM-9954", "IKM-4488", "OLP-1122", "WSX-6656", "EDC-1190",
	"RFV-5524", "TGB-9958", "YHN-3392", "UJM-7726", "ZAQ-1128", "XSW-6652", "CDE-3396",
}

var procedimentos = []string{
	"consulta cardiologica", "exame de sangue completo", "fisioterapia ortopedica",
	"consulta ortopedica", "exame de imagem (ressonancia)", "consulta psiquiatrica",
	"sessao de fonoaudiologia", "exame oftalmologico", "consulta dermatologica",
	"exame de densitometria ossea", "consulta nutricional", "sessao de acupuntura",
	"consulta ginecologica", "exame de urina completo", "sessao de terapia ocupacional",
	"consulta endocrinologica", "exame de eletrocardiograma", "consulta neurologica",
	"sessao de pilates terapeutico", "exame de audiometria", "consulta pediatrica",
	"exame de mamografia", "sessao de psicoterapia", "consulta geriatrica",
	"consulta de clinica geral", "exame de colonoscopia", "sessao de fonoterapia",
	"consulta urologica", "exame de tomografia",
}

var valores = []string{
	"3.210,50", "1.870,00", "5.640,00", "2.430,75", "890,00", "4.120,30",
	"1.250,00", "3.980,60", "2.760,00", "6.310,90", "1.540,00", "2.990,25",
	"3.450,00", "1.780,50", "4.560,00", "2.220,80", "5.120,00", "1.630,40",
	"3.870,00", "2.045,90", "4.780,60", "1.395,00", "6.020,50", "2.510,30",
	"3.660,00", "1.925,80", "4.310,00", "2.870,60", "5.480,00", "1.485,70",
	"3.120,00", "2.640,90", "4.950,00", "1.780,00", "3.390,60", "2.210,00",
	"5.870,00", "1.660,40", "4.120,00", "2.980,50", "3.780,90", "2.340,00", "4.910,60",
	"1.590,00", "3.260,40", "2.150,80", "4.430,00",
}

var datas = []string{
	"12/03/2026", "02/04/2026", "18/05/2026", "25/03/2026", "09/04/2026", "30/04/2026",
	"14/03/2026", "21/05/2026", "05/04/2026", "11/05/2026", "28/03/2026", "16/04/2026",
	"03/06/2026", "19/06/2026", "07/06/2026", "24/06/2026", "01/07/2026", "15/07/2026",
	"22/07/2026", "29/07/2026", "06/02/2026", "13/02/2026", "20/02/2026", "27/02/2026",
	"04/02/2026", "10/06/2026", "17/03/2026", "26/04/2026", "02/05/2026", "08/06/2026",
	"23/06/2026",
}

type templateFn func(nome, campo2, data, valor string) string

var templatesAuto = map[string]templateFn{
	"Oficina Estrela": func(nome, placa, data, valor string) string {
		return "ATIVA ORCAMENTOS AUTOMOTIVOS OFICINA ESTRELA LTDA CNPJ 12.345.678/0001-90 Rua das Turbinas 450 Distrito Industrial " +
			"Segurado: " + nome + " Placa do veiculo: " + placa + " Data do sinistro: " + data + " " +
			"Descricao do servico: reparo de lataria e pintura no para-choque dianteiro Valor total do reparo: R$ " + valor
	},
	"Auto Center Silva": func(nome, placa, data, valor string) string {
		return "AUTO CENTER SILVA - FUNILARIA E PINTURA - CNPJ 98.765.432/0001-11 Av. dos Mecanicos 220 " +
			"Cliente/Segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data + " " +
			"Servico executado: troca de para-lama e revisao de suspensao dianteira Valor: R$ " + valor
	},
	"Funilaria Rio Bonito": func(nome, placa, data, valor string) string {
		return "FUNILARIA RIO BONITO ME CNPJ 45.111.222/0001-33 Rua Rio Bonito 88 " +
			"Nome do segurado: " + nome + " Placa do veiculo: " + placa + " Data: " + data + " " +
			"Orcamento: substituicao de parachoque traseiro e polimento Valor total: R$ " + valor
	},
	"Oficina Nova Aliança": func(nome, placa, data, valor string) string {
		return "OFICINA NOVA ALIANCA LTDA CNPJ 22.333.444/0001-55 Estrada Velha 1200 " +
			"Segurado: " + nome + " Placa do carro: " + placa + " Data do orcamento: " + data + " " +
			"Descricao: reparo de amassado na porta dianteira Valor cobrado: R$ " + valor
	},
	"Mecânica Horizonte": func(nome, placa, data, valor string) string {
		return "MECANICA HORIZONTE LTDA CNPJ 51.222.888/0001-19 Av. do Horizonte 640 " +
			"Segurado: " + nome + " Placa do veiculo: " + placa + " Data do servico: " + data + " " +
			"Descricao: alinhamento e balanceamento apos colisao lateral Valor total: R$ " + valor
	},
	"Auto Reparos União": func(nome, placa, data, valor string) string {
		return "AUTO REPAROS UNIAO ME CNPJ 63.444.777/0001-28 Rua da Uniao 305 " +
			"Nome do segurado: " + nome + " Placa: " + placa + " Data do atendimento: " + data + " " +
			"Servico: troca de para-brisa trincado Valor cobrado: R$ " + valor
	},
}

var templatesSaude = map[string]templateFn{
	"Clínica Vitalis": func(nome, procedimento, data, valor string) string {
		return "CLINICA VITALIS SAUDE OCUPACIONAL CNPJ 33.222.111/0001-44 Av. Paulista 900 " +
			"Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data + " " +
			"Valor cobrado: R$ " + valor
	},
	"Hospital Santa Clara": func(nome, procedimento, data, valor string) string {
		return "HOSPITAL SANTA CLARA CNPJ 66.555.444/0001-22 Rua das Acacias 310 " +
			"Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data + " " +
			"Valor total: R$ " + valor
	},
	"Centro Médico Bem Estar": func(nome, procedimento, data, valor string) string {
		return "CENTRO MEDICO BEM ESTAR CNPJ 77.888.999/0001-66 Rua da Saude 45 " +
			"Nome do beneficiario: " + nome + " Procedimento: " + procedimento + " Data da consulta: " + data + " " +
			"Valor cobrado: R$ " + valor
	},
	"Clínica São Rafael": func(nome, procedimento, data, valor string) string {
		return "CLINICA SAO RAFAEL CNPJ 84.111.222/0001-37 Rua Sao Rafael 512 " +
			"Paciente/Beneficiario: " + nome + " Procedimento: " + procedimento + " Data do atendimento: " + data + " " +
			"Valor total: R$ " + valor
	},
	"Instituto Saúde Plena": func(nome, procedimento, data, valor string) string {
		return "INSTITUTO SAUDE PLENA LTDA CNPJ 91.333.555/0001-08 Av. da Saude Plena 78 " +
			"Beneficiario: " + nome + " Procedimento realizado: " + procedimento + " Data: " + data + " " +
			"Valor cobrado: R$ " + valor
	},
}

// prenomes tem uma entrada por nome (len(nomes)); sobrenomes tem só 37 --
// mesma assimetria do original em Java (SOBRENOMES so' preenche os 37
// primeiros nomes mesmo len(nomes)=41), preservada aqui por fidelidade de
// comportamento observável (os índices usados nunca ultrapassam 37).
var prenomes = func() []string {
	p := make([]string, len(nomes))
	for i, n := range nomes {
		p[i] = strings.SplitN(n, " ", 2)[0]
	}
	return p
}()

var sobrenomes = func() []string {
	s := make([]string, 37)
	for i := 0; i < 37; i++ {
		partes := strings.Split(nomes[i], " ")
		s[i] = strings.Join(partes[1:], " ")
	}
	return s
}()

type fonteQtd struct {
	fonte string
	n     int
}

var fontesAuto = []fonteQtd{
	{"Oficina Estrela", 60}, {"Auto Center Silva", 40}, {"Funilaria Rio Bonito", 30},
	{"Oficina Nova Aliança", 25}, {"Mecânica Horizonte", 15}, {"Auto Reparos União", 10},
}

var fontesSaude = []fonteQtd{
	{"Clínica Vitalis", 50}, {"Hospital Santa Clara", 30}, {"Centro Médico Bem Estar", 20},
	{"Clínica São Rafael", 12}, {"Instituto Saúde Plena", 8},
}

func parseValorBr(valor string) float64 {
	semSeparadores := strings.ReplaceAll(strings.ReplaceAll(valor, ".", ""), ",", ".")
	v, _ := strconv.ParseFloat(semSeparadores, 64)
	return v
}

// GerarExemplo gera um exemplo determinístico de dataset (mesmo índice ->
// mesmo exemplo sempre) -- equivalente a DatasetScaling.gerarExemplo.
func GerarExemplo(caso, fonte string, indice int) ExemploDataset {
	nome := prenomes[indice%len(prenomes)] + " " + sobrenomes[(indice*7+3)%len(sobrenomes)]
	data := datas[(indice*7+3)%len(datas)]
	valor := valores[(indice*11+5)%len(valores)]

	var entrada string
	saida := map[string]any{}
	var instrucao string

	if caso == "amplitude-auto" {
		placa := placas[(indice*13+2)%len(placas)]
		entrada = templatesAuto[fonte](nome, placa, data, valor)
		saida["segurado"] = nome
		saida["placa"] = placa
		saida["valor"] = parseValorBr(valor)
		instrucao = "Extraia segurado, placa e valor do orçamento de oficina abaixo."
	} else {
		procedimento := procedimentos[(indice*5+1)%len(procedimentos)]
		entrada = templatesSaude[fonte](nome, procedimento, data, valor)
		saida["beneficiario"] = nome
		saida["procedimento"] = procedimento
		saida["valor"] = parseValorBr(valor)
		instrucao = "Extraia beneficiário, procedimento e valor do recibo médico abaixo."
	}

	return ExemploDataset{
		Instrucao: instrucao, Entrada: entrada, Saida: saida,
		Caso: caso, Fonte: fonte, ID: fmt.Sprintf("%s-%s-%d", caso, fonte, indice),
	}
}

func buscarPorID(exemplos []ExemploDataset, id string) ExemploDataset {
	for _, e := range exemplos {
		if e.ID == id {
			return e
		}
	}
	panic("id nao encontrado: " + id)
}

func reidentificar(original ExemploDataset, caso, fonte, novoID string) ExemploDataset {
	original.Caso = caso
	original.Fonte = fonte
	original.ID = novoID
	return original
}

// GerarDatasetBruto gera os 305 exemplos brutos (183 amplitude-auto + 122
// amplitude-saude-empresarial) -- equivalente a DatasetScaling.gerarDatasetBruto.
func GerarDatasetBruto() []ExemploDataset {
	var exemplos []ExemploDataset
	for _, fn := range fontesAuto {
		for i := 0; i < fn.n; i++ {
			exemplos = append(exemplos, GerarExemplo("amplitude-auto", fn.fonte, i))
		}
	}
	for _, fn := range fontesSaude {
		for i := 0; i < fn.n; i++ {
			exemplos = append(exemplos, GerarExemplo("amplitude-saude-empresarial", fn.fonte, i))
		}
	}

	exemplos = append(exemplos, reidentificar(buscarPorID(exemplos, "amplitude-auto-Oficina Estrela-0"),
		"amplitude-auto", "Oficina Estrela", "amplitude-auto-Oficina Estrela-0-reenviado"))
	exemplos = append(exemplos, reidentificar(buscarPorID(exemplos, "amplitude-auto-Auto Center Silva-0"),
		"amplitude-auto", "Auto Center Silva", "amplitude-auto-Auto Center Silva-0-reenviado"))

	base1 := buscarPorID(exemplos, "amplitude-auto-Oficina Estrela-5")
	exemplos = append(exemplos, ExemploDataset{
		Instrucao: base1.Instrucao,
		Entrada:   strings.ReplaceAll(base1.Entrada, "Placa do veiculo:", "P1aca do veicu1o:"),
		Saida:     base1.Saida, Caso: "amplitude-auto", Fonte: "Oficina Estrela",
		ID: "amplitude-auto-Oficina Estrela-5-ruido-ocr",
	})

	exemplos = append(exemplos, reidentificar(buscarPorID(exemplos, "amplitude-saude-empresarial-Clínica Vitalis-0"),
		"amplitude-saude-empresarial", "Clínica Vitalis", "amplitude-saude-empresarial-Clínica Vitalis-0-reenviado"))

	base2 := buscarPorID(exemplos, "amplitude-saude-empresarial-Clínica Vitalis-10")
	exemplos = append(exemplos, ExemploDataset{
		Instrucao: base2.Instrucao,
		Entrada:   strings.ReplaceAll(base2.Entrada, "Valor cobrado:", "Va1or cobrad0:"),
		Saida:     base2.Saida, Caso: "amplitude-saude-empresarial", Fonte: "Clínica Vitalis",
		ID: "amplitude-saude-empresarial-Clínica Vitalis-10-ruido-ocr",
	})

	return exemplos
}

// RelatorioCaso é a comparação de diversidade de fontes antes/depois do
// balanceamento por temperatura, pra um caso.
type RelatorioCaso struct {
	ContagensAntes  map[string]int
	EntropiaAntes   float64
	NEfetivoAntes   float64
	ContagensDepois map[string]int
	EntropiaDepois  float64
	NEfetivoDepois  float64
}

// ResultadoPipeline é o resultado de LimparEBalancear.
type ResultadoPipeline struct {
	Original            int
	AposDedup           int
	DuplicatasRemovidas int
	Total               int
	ExemplosFinal       []ExemploDataset
	RelatorioPorCaso    map[string]RelatorioCaso
}

func contarPorFonte(exemplos []ExemploDataset, caso string) map[string]int {
	contagem := make(map[string]int)
	for _, e := range exemplos {
		if e.Caso == caso {
			contagem[e.Fonte]++
		}
	}
	return contagem
}

// LimparEBalancear deduplica por caso (mantendo a primeira ocorrência) e
// depois balanceia por temperatura -- equivalente a
// DatasetScaling.limparEBalancear / limparEBalancear() do .js.
func LimparEBalancear(exemplos []ExemploDataset, alvos map[string]int) ResultadoPipeline {
	paraDedup := make([]minhash.Exemplo, len(exemplos))
	for i, e := range exemplos {
		paraDedup[i] = e.paraDedup()
	}

	remover := make(map[int]bool)
	for _, caso := range []string{"amplitude-auto", "amplitude-saude-empresarial"} {
		var indicesDoCaso []int
		for i, e := range exemplos {
			if e.Caso == caso {
				indicesDoCaso = append(indicesDoCaso, i)
			}
		}
		doCaso := make([]minhash.Exemplo, len(indicesDoCaso))
		for i, idx := range indicesDoCaso {
			doCaso[i] = paraDedup[idx]
		}
		dedup := minhash.EncontrarQuaseDuplicatasGenerico(doCaso, 5)
		for _, par := range dedup.ParesDuplicata {
			remover[indicesDoCaso[par.J]] = true
		}
	}

	var aposDedup []ExemploDataset
	for i, e := range exemplos {
		if !remover[i] {
			aposDedup = append(aposDedup, e)
		}
	}

	relatorio := make(map[string]RelatorioCaso)
	atual := aposDedup

	for _, caso := range []string{"amplitude-auto", "amplitude-saude-empresarial"} {
		contagensAntes := contarPorFonte(atual, caso)
		distAntes := minhash.DistribuicaoDe(contagensAntes)
		alvo := alvos[caso]
		alocacao := minhash.AlocarComCapacidade(contagensAntes, minhash.AlphaTemperatura, alvo)

		usados := make(map[string]int)
		var selecionados, outros []ExemploDataset
		for _, e := range atual {
			if e.Caso != caso {
				outros = append(outros, e)
				continue
			}
			usadoAtual := usados[e.Fonte]
			if usadoAtual < alocacao[e.Fonte] {
				selecionados = append(selecionados, e)
				usados[e.Fonte] = usadoAtual + 1
			}
		}
		novo := append(append([]ExemploDataset{}, outros...), selecionados...)
		atual = novo

		contagensDepois := contarPorFonte(atual, caso)
		distDepois := minhash.DistribuicaoDe(contagensDepois)
		relatorio[caso] = RelatorioCaso{
			ContagensAntes: contagensAntes,
			EntropiaAntes:  minhash.EntropiaShannon(distAntes),
			NEfetivoAntes:  minhash.NumeroEfetivoFontes(distAntes),

			ContagensDepois: contagensDepois,
			EntropiaDepois:  minhash.EntropiaShannon(distDepois),
			NEfetivoDepois:  minhash.NumeroEfetivoFontes(distDepois),
		}
	}

	return ResultadoPipeline{
		Original:            len(exemplos),
		AposDedup:           len(aposDedup),
		DuplicatasRemovidas: len(remover),
		Total:               len(atual),
		ExemplosFinal:       atual,
		RelatorioPorCaso:    relatorio,
	}
}
