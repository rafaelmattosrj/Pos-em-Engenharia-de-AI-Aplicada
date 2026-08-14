package trialforge

import (
	"errors"
	"fmt"
	"time"
)

// Tempos de simulacao de trabalho, proporcionais aos timeouts reais
// definidos no Slide 2 (30s/45s/60s) mas escalados / 100 pra manter a demo
// rapida — o que importa aqui e a proporcao entre os tres, nao o valor
// absoluto em segundos.
const (
	TempoProtocoloMs = 300
	TempoICFMs       = 450
	TempoCSRMs       = 600

	// TempoEmendaMs: a emenda do comite de etica dispara 500ms depois do
	// protocolo:pronto — depois que o ICF ja terminou (450ms), mas antes do
	// CSR terminar (600ms). Nao e uma corrida: sao atrasos fixos e
	// diferentes, a ordem de conclusao e deterministica.
	TempoEmendaMs = 500

	margemDeSeguranca = 1.5

	// MaxTentativas: "ate tres tentativas" (Slide 2) — mesmo limite aplicado
	// de forma consistente na politica de retry do ICF.
	MaxTentativas = 3
)

// Timeout explicito de verdade (Slide 2: 30s/45s/60s / 100), com margem de
// seguranca sobre o tempo normal de trabalho — nao e o mesmo valor do tempo
// de trabalho, senao toda execucao normal flertaria com o proprio timeout.
var (
	TimeoutProtocoloMs = int(TempoProtocoloMs * margemDeSeguranca)
	TimeoutICFMs       = int(TempoICFMs * margemDeSeguranca)
	TimeoutCSRMs       = int(TempoCSRMs * margemDeSeguranca)

	// TempoICFTravadoMs simula um agente travado: demora bem mais que o
	// proprio timeout, entao nunca chega a responder a tempo — bem diferente
	// de "retornou um erro".
	TempoICFTravadoMs = TimeoutICFMs * 3
)

// CriteriosIniciais e o ponto de partida do Protocolo, versao 1.
var CriteriosIniciais = map[string]int{"idadeMinima": 13}

// AgenteProtocolo e o porte de agenteProtocolo (JS) / agente_protocolo
// (Python) — Slide 2: A Arquitetura Completa.
func AgenteProtocolo(barramento *Barramento, estudo string, estado *EstadoProtocolo, ordem *OrdemDeExecucao, opcoes OpcoesFluxo) (DadoProtocolo, error) {
	time.Sleep(time.Duration(TempoProtocoloMs) * time.Millisecond)
	if opcoes.ForcarFalhaProtocolo {
		return DadoProtocolo{}, errors.New("Agente Protocolo: falha simulada ao gerar critérios.")
	}
	ordem.Registrar("protocolo") // registrado ANTES do emit: prova causal de que Sequential foi respeitado

	// Evento rico: copia do estado atual, nao uma referencia — revisoes
	// futuras no EstadoProtocolo nao mudam retroativamente o que ja foi
	// publicado aqui.
	dado := DadoProtocolo{Versao: estado.Versao(), Criterios: estado.Criterios(), Estudo: estudo}
	barramento.Emit("protocolo:pronto", dado) // nao-bloqueante: publica e segue em frente

	if opcoes.ComEmendaEtica {
		// Simula uma emenda chegando de forma assincrona, independente do
		// fluxo principal — exatamente como um comite de etica revisaria um
		// criterio dias ou semanas depois, sem nenhuma relacao com o timing de
		// ICF/CSR.
		go func() {
			time.Sleep(time.Duration(TempoEmendaMs) * time.Millisecond)
			estado.Revisar(map[string]int{"idadeMinima": 12}, "Comitê de ética corrigiu a idade mínima de 13 para 12 anos.")
		}()
	}

	return dado, nil
}

// ResultadoICFBase e o resultado do ICF antes de saber quantas tentativas o
// documento levou — e exatamente o que fica gravado em DocumentoRegistro.
type ResultadoICFBase struct {
	Agente           string
	Secao            string
	CriterioUsado    map[string]int
	VersaoUsada      int
	VersaoAoConcluir int
}

// ResultadoICF e o resultado final do Agente ICF, ja com o contador de
// tentativas do registro idempotente.
type ResultadoICF struct {
	ResultadoICFBase
	Tentativas int
}

// AgenteICF e o porte de agenteICF (JS) / agente_icf (Python).
func AgenteICF(dadoProtocolo DadoProtocolo, estado *EstadoProtocolo, documentos *DocumentosGerados, ordem *OrdemDeExecucao, opcoes OpcoesFluxo) (ResultadoICF, error) {
	ordem.Registrar("icf:inicio") // registrado ANTES do sleep: prova que so comeca depois do protocolo:pronto

	// ForcarTravamentoICF simula o agente travado (demora alem do timeout) —
	// bem diferente de ForcarFalhaICF, que simula um erro retornado dentro do
	// proprio agente.
	tempoMs := TempoICFMs
	if opcoes.ForcarTravamentoICF {
		tempoMs = TempoICFTravadoMs
	}
	time.Sleep(time.Duration(tempoMs) * time.Millisecond)

	chave := fmt.Sprintf("%d:icf", dadoProtocolo.Versao)
	base := ResultadoICFBase{
		Agente: "ICF",
		Secao: fmt.Sprintf(
			"Assentimento gerado com idade mínima de %d anos (a partir da versão %d do protocolo).",
			dadoProtocolo.Criterios["idadeMinima"], dadoProtocolo.Versao,
		),
		CriterioUsado: dadoProtocolo.Criterios,
		VersaoUsada:   dadoProtocolo.Versao,
		// Conferido na hora de concluir, nao na hora de comecar — e isso que
		// permite o Supervisor perceber se uma emenda chegou enquanto o ICF
		// ainda trabalhava.
		VersaoAoConcluir: estado.Versao(),
	}

	// O documento e gravado ANTES da falha simulada: o cenario real que retry
	// precisa tratar nao e "a tarefa nunca rodou", e "a tarefa rodou, mas a
	// confirmacao se perdeu" — daí a necessidade de idempotencia, nao so de
	// tentar de novo.
	registro := documentos.Registrar(chave, base)
	if opcoes.ForcarFalhaICF {
		return ResultadoICF{}, errors.New("Agente ICF: falha simulada na confirmação, depois do documento já ter sido gravado.")
	}
	return ResultadoICF{ResultadoICFBase: base, Tentativas: registro.Tentativas}, nil
}

// ResultadoCSR e o resultado do Agente CSR. Sem registro idempotente: o CSR
// nao grava documento.
type ResultadoCSR struct {
	Agente           string
	Sintese          string
	CriterioUsado    map[string]int
	VersaoUsada      int
	VersaoAoConcluir int
}

// AgenteCSR e o porte de agenteCSR (JS) / agente_csr (Python).
func AgenteCSR(dadoProtocolo DadoProtocolo, estado *EstadoProtocolo, ordem *OrdemDeExecucao) (ResultadoCSR, error) {
	ordem.Registrar("csr:inicio") // registrado ANTES do sleep: prova que so comeca depois do protocolo:pronto
	time.Sleep(time.Duration(TempoCSRMs) * time.Millisecond)

	return ResultadoCSR{
		Agente: "CSR",
		Sintese: fmt.Sprintf(
			"Conformidade ICH E3 verificada com idade mínima de %d anos (a partir da versão %d do protocolo).",
			dadoProtocolo.Criterios["idadeMinima"], dadoProtocolo.Versao,
		),
		CriterioUsado:    dadoProtocolo.Criterios,
		VersaoUsada:      dadoProtocolo.Versao,
		VersaoAoConcluir: estado.Versao(),
	}, nil
}
