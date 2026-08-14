package trialforge

import (
	"fmt"
	"strings"
)

// ResultadoFluxo e o porte do objeto de retorno de rodarFluxoTrialForge /
// rodar_fluxo_trialforge (JS/Python) — tudo que a demonstracao e os testes
// precisam inspecionar.
type ResultadoFluxo struct {
	DadoProtocolo     DadoProtocolo
	Reacao            Reacao
	DecisaoSupervisor string
	Documentos        *DocumentosGerados
	TentativasICF     int
	TentativasRetry   int
	OrdemDeExecucao   []string
	EstadoProtocolo   *EstadoProtocolo
	Verificacao       *Verificacao
}

// RodarFluxoTrialForge e o porte de rodarFluxoTrialForge (JS) /
// rodar_fluxo_trialforge (Python) — o fluxo completo: Sequential (Protocolo)
// -> Parallel (ICF+CSR reagindo ao evento "protocolo:pronto") -> Supervisor
// (retry CAP + consistencia/compensacao Saga).
//
// Escopo por fluxo, nao pacote: cada chamada cria seu proprio "banco de
// documentos", seu proprio registro de ordem de execucao e seu proprio
// estado de protocolo — nao compartilhados entre estudos.
func RodarFluxoTrialForge(estudo string, opcoes OpcoesFluxo) (ResultadoFluxo, error) {
	documentos := NovoDocumentosGerados()
	ordem := &OrdemDeExecucao{}
	estado := NovoEstadoProtocolo(CriteriosIniciais)
	barramento := NovoBarramento()

	// Inscreve ANTES do emit.
	reacaoCh := InscreverReacaoParalela(barramento, opcoes.Estrategia, opcoes, documentos, ordem, estado)

	dadoProtocolo, err := comTimeout(TimeoutProtocoloMs, "Agente Protocolo", func() (DadoProtocolo, error) {
		return AgenteProtocolo(barramento, estudo, estado, ordem, opcoes)
	}) // Sequential
	if err != nil {
		return ResultadoFluxo{}, err
	}

	reacao := <-reacaoCh // Parallel ja rodou concorrente enquanto isso resolvia
	decisao := DecidirEstrategiaDeRetry(reacao)
	tentativasRetry := 0

	if decisao == "retry_icf_apenas" {
		// O retry de verdade, COM LIMITE (CAP): reaproveita a MESMA chave de
		// idempotencia em cada tentativa, provando que nenhuma delas duplica o
		// documento do ICF.
		resultadoRetry := ExecutarICFComRetry(dadoProtocolo, estado, documentos, ordem, opcoes)
		tentativasRetry = resultadoRetry.TentativasFeitas
		if resultadoRetry.Sucesso {
			r := resultadoRetry.Resultado
			reacao.ResultadoICF = &r
			reacao.Ok = true
			decisao = "retry_icf_executado_com_sucesso"
		} else {
			// CAP: o limite de tentativas esgotou — segue em frente sem esse
			// resultado (Disponibilidade), em vez de esperar pra sempre (Consistencia).
			reacao.Ok = false
			reacao.ErroICF = resultadoRetry.Erro
			decisao = "retry_esgotado_seguindo_sem_icf"
		}
	}

	tentativasICF := 0
	if registro, ok := documentos.Get(fmt.Sprintf("%d:icf", dadoProtocolo.Versao)); ok {
		tentativasICF = registro.Tentativas
	}

	// Mecanismo 2: so faz sentido verificar consistencia quando os dois
	// resultados de conteudo existem de verdade (depois do mecanismo 1 ja ter
	// resolvido qualquer falha tecnica).
	var verificacao *Verificacao
	if reacao.Ok && reacao.ResultadoICF != nil && reacao.ResultadoCSR != nil {
		v := VerificarConsistencia(*reacao.ResultadoICF, *reacao.ResultadoCSR)
		verificacao = &v
		if !v.Consistente {
			compensacao := CompensarDivergencia(v.AgentesDivergentes, estado, estudo, documentos, ordem)
			if compensacao.ResultadoICF != nil {
				reacao.ResultadoICF = compensacao.ResultadoICF
			}
			if compensacao.ResultadoCSR != nil {
				reacao.ResultadoCSR = compensacao.ResultadoCSR
			}
			decisao = "saga_compensacao_" + strings.ToLower(strings.Join(v.AgentesDivergentes, "_e_"))
		}
	}

	return ResultadoFluxo{
		DadoProtocolo:     dadoProtocolo,
		Reacao:            reacao,
		DecisaoSupervisor: decisao,
		Documentos:        documentos,
		TentativasICF:     tentativasICF,
		TentativasRetry:   tentativasRetry,
		OrdemDeExecucao:   ordem.Itens(),
		EstadoProtocolo:   estado,
		Verificacao:       verificacao,
	}, nil
}
