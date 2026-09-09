// Package fullvslora porta full-vs-lora-tradeoff-tool.js (Modulo 4.4). Dados
// reais de um treino full fine-tuning comparado contra os tres treinos LoRA
// do Modulo 4.3 -- nenhum numero aqui e estimado, capturados em 2026-08-08.
package fullvslora

import "math"

type Configuracao struct {
	Tipo               string
	PercentualModelo   float64
	ValLossFinal       float64
	PicoMemGB          float64
	TamanhoCheckpoinMB int
}

var ConfiguracoesReais = []Configuracao{
	{"LoRA rank 4", 0.074, 1.246, 10.787, 13},
	{"LoRA rank 8", 0.147, 0.895, 10.833, 27},
	{"LoRA rank 16", 0.295, 0.725, 10.930, 52},
	{"Full fine-tuning", 22.567, 0.612, 15.338, 1992},
}

type ResultadoInferencia struct {
	AcertouTodosCampos bool
}

var ExemploFacil = ResultadoInferencia{AcertouTodosCampos: true}
var ExemploComDistratores = ResultadoInferencia{AcertouTodosCampos: true}

func round(v float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(v*fator) / fator
}

func CalcularGanhoValLoss(base, comparado Configuracao) float64 {
	ganho := ((base.ValLossFinal - comparado.ValLossFinal) / base.ValLossFinal) * 100
	return round(ganho, 2)
}

type RazaoCusto struct {
	Parametro  float64
	Memoria    float64
	Checkpoint float64
}

func CalcularRazaoCusto(base, comparado Configuracao) RazaoCusto {
	return RazaoCusto{
		Parametro:  round(comparado.PercentualModelo/base.PercentualModelo, 1),
		Memoria:    round(comparado.PicoMemGB/base.PicoMemGB, 2),
		Checkpoint: round(float64(comparado.TamanhoCheckpoinMB)/float64(base.TamanhoCheckpoinMB), 1),
	}
}

type DecisaoFullFineTuning struct {
	MelhorLora      string
	GanhoPercentual float64
	ValeAPena       bool
}

func acharConfiguracao(configuracoes []Configuracao, tipo string) Configuracao {
	for _, c := range configuracoes {
		if c.Tipo == tipo {
			return c
		}
	}
	panic("configuração não encontrada: " + tipo)
}

func ValeAPenaFullFineTuning(configuracoes []Configuracao, limiarGanhoMinimo float64) DecisaoFullFineTuning {
	var melhorLora Configuracao
	primeiro := true
	for _, c := range configuracoes {
		if len(c.Tipo) >= 4 && c.Tipo[:4] == "LoRA" {
			if primeiro || c.ValLossFinal < melhorLora.ValLossFinal {
				melhorLora = c
				primeiro = false
			}
		}
	}
	full := acharConfiguracao(configuracoes, "Full fine-tuning")
	ganho := CalcularGanhoValLoss(melhorLora, full)
	return DecisaoFullFineTuning{
		MelhorLora:      melhorLora.Tipo,
		GanhoPercentual: ganho,
		ValeAPena:       ganho >= limiarGanhoMinimo,
	}
}
