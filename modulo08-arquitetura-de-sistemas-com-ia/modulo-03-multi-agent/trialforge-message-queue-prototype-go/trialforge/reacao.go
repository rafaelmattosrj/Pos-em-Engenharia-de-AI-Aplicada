package trialforge

// Reacao e o resultado da reacao Parallel (ICF + CSR) ao evento
// "protocolo:pronto".
type Reacao struct {
	Ok                     bool
	ResultadoICF           *ResultadoICF
	ResultadoCSR           *ResultadoCSR
	ErroICF                string
	ResultadoCSRPreservado bool
}

// InscreverReacaoParalela e o porte de inscreverReacaoParalela (JS) /
// inscrever_reacao_paralela (Python) — registra o ouvinte ANTES do emit (o
// chamador deve chamar isto antes de disparar o Agente Protocolo) e devolve
// um canal que recebe a Reacao assim que ICF e CSR terminarem (ou
// estourarem o timeout).
//
// ICF e CSR sao iniciados CONCORRENTEMENTE (via iniciarComTimeout, que nao
// bloqueia) antes de qualquer branch de estrategia — e isso que garante que
// a estrategia EstrategiaAll (o bug) e a EstrategiaAllSettled (a correcao)
// enxerguem exatamente a mesma corrida, so decidindo diferente o que fazer
// com uma falha parcial.
func InscreverReacaoParalela(barramento *Barramento, estrategia Estrategia, opcoesFalha OpcoesFluxo, documentos *DocumentosGerados, ordem *OrdemDeExecucao, estado *EstadoProtocolo) <-chan Reacao {
	saida := make(chan Reacao, 1)

	barramento.Once("protocolo:pronto", func(dadoProtocolo DadoProtocolo) {
		icfCh := iniciarComTimeout(TimeoutICFMs, "Agente ICF", func() (ResultadoICF, error) {
			return AgenteICF(dadoProtocolo, estado, documentos, ordem, opcoesFalha)
		})
		csrCh := iniciarComTimeout(TimeoutCSRMs, "Agente CSR", func() (ResultadoCSR, error) {
			return AgenteCSR(dadoProtocolo, estado, ordem)
		})

		if estrategia == EstrategiaAll {
			// BUG (TP, paragrafos 68-71): equivalente a Promise.all — a primeira
			// falha derruba o lote inteiro, mesmo que o outro agente tenha
			// terminado bem.
			icfR := <-icfCh
			if icfR.erro != nil {
				saida <- Reacao{Ok: false, ErroICF: icfR.erro.Error(), ResultadoCSRPreservado: false}
				return
			}
			csrR := <-csrCh
			if csrR.erro != nil {
				saida <- Reacao{Ok: false, ErroICF: csrR.erro.Error(), ResultadoCSRPreservado: false}
				return
			}
			icf, csr := icfR.valor, csrR.valor
			saida <- Reacao{Ok: true, ResultadoICF: &icf, ResultadoCSR: &csr, ResultadoCSRPreservado: true}
			return
		}

		// Correcao: equivalente a Promise.allSettled — cada resultado e
		// preservado independente do outro ter falhado.
		icfR := <-icfCh
		csrR := <-csrCh
		reacao := Reacao{ResultadoCSRPreservado: csrR.erro == nil}
		if icfR.erro == nil {
			icf := icfR.valor
			reacao.ResultadoICF = &icf
		} else {
			reacao.ErroICF = icfR.erro.Error()
		}
		if csrR.erro == nil {
			csr := csrR.valor
			reacao.ResultadoCSR = &csr
		}
		reacao.Ok = icfR.erro == nil && csrR.erro == nil
		saida <- reacao
	})

	return saida
}
