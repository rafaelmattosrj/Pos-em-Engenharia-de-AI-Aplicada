package trialforge

// Porte dos dois mecanismos do Supervisor (JS/Python):
//
// Mecanismo 1 — decide e executa a estrategia de retry (Teorema CAP, Modulo
// 3.4): timeout explicito, retry com limite (ate MaxTentativas),
// idempotencia. NAO e o padrao Saga: o ICF falhou DENTRO da propria
// execucao, nao e um problema descoberto DEPOIS por uma etapa posterior.
//
// Mecanismo 2 — verifica consistencia (Modulo 3.2, paragrafo 88) e compensa
// via Saga de verdade (Modulo 3.4): so o agente que ficou defasado e
// regenerado, o outro nunca e retrabalhado.

// DecidirEstrategiaDeRetry e o porte de decidirEstrategiaDeRetry (JS) /
// decidir_estrategia_de_retry (Python).
func DecidirEstrategiaDeRetry(reacao Reacao) string {
	if reacao.Ok {
		return "nenhum_retry_necessario"
	}
	if reacao.ResultadoCSRPreservado {
		return "retry_icf_apenas" // so refaz quem falhou, preserva o CSR que ja terminou
	}
	return "retry_ambos" // sem visibilidade de quem terminou bem, precisa refazer tudo
}

// ResultadoRetry e o resultado de ExecutarICFComRetry — sucesso com o valor,
// ou falha apos MaxTentativas.
type ResultadoRetry struct {
	Sucesso          bool
	Resultado        ResultadoICF
	Erro             string
	TentativasFeitas int
}

// ExecutarICFComRetry e o porte de executarICFComRetry (JS) /
// executar_icf_com_retry (Python) — retry COM LIMITE de verdade, nao um
// unico retry incondicional. Tenta ate MaxTentativas vezes, cada uma
// protegida pelo mesmo timeout real do despacho original; se a falha e
// transitoria, a causa raiz ja passou e a 2a tentativa resolve. Se a falha e
// persistente, o loop esgota as tentativas e desiste — CAP manda decidir
// entre esperar mais ou seguir em frente sem aquele resultado
// (Disponibilidade sobre Consistencia), nao esperar pra sempre.
func ExecutarICFComRetry(dadoProtocolo DadoProtocolo, estado *EstadoProtocolo, documentos *DocumentosGerados, ordem *OrdemDeExecucao, opcoesFalha OpcoesFluxo) ResultadoRetry {
	persistente := opcoesFalha.PersistirFalhaNoRetry
	var ultimoErro error

	for tentativa := 1; tentativa <= MaxTentativas; tentativa++ {
		opcoesDaTentativa := OpcoesFluxo{}
		if persistente {
			opcoesDaTentativa.ForcarFalhaICF = opcoesFalha.ForcarFalhaICF
			opcoesDaTentativa.ForcarTravamentoICF = opcoesFalha.ForcarTravamentoICF
		}

		resultado, err := comTimeout(TimeoutICFMs, "Agente ICF", func() (ResultadoICF, error) {
			return AgenteICF(dadoProtocolo, estado, documentos, ordem, opcoesDaTentativa)
		})
		if err == nil {
			return ResultadoRetry{Sucesso: true, Resultado: resultado, TentativasFeitas: tentativa}
		}
		ultimoErro = err
	}

	return ResultadoRetry{Sucesso: false, Erro: ultimoErro.Error(), TentativasFeitas: MaxTentativas}
}

// Verificacao e o porte de verificarConsistencia (JS) / verificar_consistencia
// (Python) — Modulo 3.2, paragrafo 88: "o Supervisor verifica se os
// critérios citados em cada um batem entre si."
type Verificacao struct {
	Consistente        bool
	ICFConsistente     bool
	CSRConsistente     bool
	AgentesDivergentes []string
}

// VerificarConsistencia compara, PARA CADA AGENTE, a versao que ele usou
// contra a versao que ja estava vigente quando ELE PROPRIO terminou.
func VerificarConsistencia(icf ResultadoICF, csr ResultadoCSR) Verificacao {
	icfConsistente := icf.VersaoUsada == icf.VersaoAoConcluir
	csrConsistente := csr.VersaoUsada == csr.VersaoAoConcluir

	var agentesDivergentes []string
	if !icfConsistente {
		agentesDivergentes = append(agentesDivergentes, "ICF")
	}
	if !csrConsistente {
		agentesDivergentes = append(agentesDivergentes, "CSR")
	}

	return Verificacao{
		Consistente:        len(agentesDivergentes) == 0,
		ICFConsistente:     icfConsistente,
		CSRConsistente:     csrConsistente,
		AgentesDivergentes: agentesDivergentes,
	}
}

// CompensacaoResultado e o resultado de CompensarDivergencia — so os campos
// dos agentes que foram de fato regenerados.
type CompensacaoResultado struct {
	ResultadoICF *ResultadoICF
	ResultadoCSR *ResultadoCSR
}

// CompensarDivergencia e o porte de compensarDivergencia (JS) /
// compensar_divergencia (Python) — compensacao Saga de verdade: so o agente
// que ficou defasado e regenerado, com a versao CORRIGIDA do protocolo; o
// outro agente, que ja estava consistente, nunca e retrabalhado. As
// chamadas sao sequenciais (nao concorrentes), igual ao original.
func CompensarDivergencia(agentesDivergentes []string, estado *EstadoProtocolo, estudo string, documentos *DocumentosGerados, ordem *OrdemDeExecucao) CompensacaoResultado {
	dadoAtual := DadoProtocolo{Versao: estado.Versao(), Criterios: estado.Criterios(), Estudo: estudo}
	var resultado CompensacaoResultado

	if contains(agentesDivergentes, "ICF") {
		r, _ := AgenteICF(dadoAtual, estado, documentos, ordem, OpcoesFluxo{})
		resultado.ResultadoICF = &r
	}
	if contains(agentesDivergentes, "CSR") {
		r, _ := AgenteCSR(dadoAtual, estado, ordem)
		resultado.ResultadoCSR = &r
	}

	return resultado
}

func contains(itens []string, alvo string) bool {
	for _, item := range itens {
		if item == alvo {
			return true
		}
	}
	return false
}
