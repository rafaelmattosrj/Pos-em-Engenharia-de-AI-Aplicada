package reactagent

import (
	"context"
	"fmt"
)

// ImprimirTrilha imprime a trilha de desenvolvimento de uma execucao do
// AgenteICF — mesmo formato do original.
func ImprimirTrilha(trilha []PassoTrilha) {
	fmt.Println("  [Trilha de desenvolvimento — não é a trilha de auditoria do Módulo 5.2]")
	for _, passo := range trilha {
		if passo.Acao == "chamou_ferramenta" {
			fmt.Printf("    volta %d (%dms): chamou %s(%v) -> %+v\n",
				passo.Volta, passo.DuracaoMs, passo.Ferramenta, passo.Argumentos, *passo.Observacao)
		} else {
			fmt.Printf("    volta %d (%dms): decidiu que já tinha o suficiente, respondeu\n",
				passo.Volta, passo.DuracaoMs)
		}
	}
}

// SimularInteracao simula tres interacoes reais entre a Mariana (usuaria) e
// o agente, cobrindo os tres comportamentos que a Missao Pratica #02 pede
// pra demonstrar no proprio prototipo: convergencia normal, ferramenta sem
// resultado, e nao-convergencia com escalonamento.
//
// Porte 1:1 de simularInteracao em react-agent-prototype.js / .py.
func SimularInteracao(ctx context.Context, client ChatClient) error {
	// ---------- Cenário 1: convergência normal, a cláusula existe ----------
	fmt.Println("===== Cenário 1: convergência normal (cláusula encontrada) =====")
	protocolo1 := "Estudo fase II, público-alvo entre 12 e 17 anos, terapia oncológica experimental."
	fmt.Println("[Mariana submete o protocolo ao Gateway]")
	fmt.Println("  ", protocolo1)
	fmt.Println()

	resultado1, err := AgenteICF(ctx, client, protocolo1)
	if err != nil {
		return err
	}
	fmt.Printf("[Agente ICF] Rascunho gerado após %d volta(s) de loop, rodando localmente:\n", resultado1.Iteracoes)
	fmt.Println("  ", resultado1.Rascunho)
	ImprimirTrilha(resultado1.Trilha)
	fmt.Println("[Sistema] Rascunho aguardando revisão do Approval Gate antes de virar versão oficial.")
	fmt.Println()

	// ---------- Cenário 2: ferramenta não encontra cláusula ----------
	fmt.Println("===== Cenário 2: ferramenta não encontra cláusula (população adulta) =====")
	protocolo2 := "Estudo fase III, população adulta (18-65 anos), diabetes tipo 2, sem envolvimento de " +
		"menores de idade."
	fmt.Println("[Mariana submete o protocolo ao Gateway]")
	fmt.Println("  ", protocolo2)
	fmt.Println()

	resultado2, err := AgenteICF(ctx, client, protocolo2)
	if err != nil {
		return err
	}
	ImprimirTrilha(resultado2.Trilha)
	if resultado2.EscalarParaAprovacaoHumana {
		fmt.Println("[Orquestrador] Loop não convergiu —", resultado2.Motivo)
		fmt.Println("[Sistema] Encaminhando ao Approval Gate para intervenção manual.")
	} else {
		fmt.Printf("[Agente ICF] Respondeu após %d volta(s), sem achar cláusula específica.\n", resultado2.Iteracoes)
		fmt.Println("[Atenção] Repare na trilha: mesmo sem achar a cláusula, o modelo escreveu um rascunho " +
			"genérico em vez de escalar — desviando da instrução do system prompt. É exatamente esse tipo " +
			"de desvio que a trilha existe para flagrar, e por isso o Approval Gate revisa antes de " +
			"qualquer coisa virar oficial.")
	}
	fmt.Println()

	// ---------- Cenário 3: não convergência, escalonamento ----------
	// maxIteracoes forçado a 1 só aqui: testado com o modelo real, mesmo sem achar a
	// cláusula, tende a responder algo em vez de insistir por várias voltas (ver o
	// desvio do Cenário 2) — comportamento de LLM não é 100% previsível. Forçar o
	// limiar garante que este cenário sempre demonstre o mecanismo de escalonamento
	// de forma confiável. Em produção, o limiar calibrado continua sendo
	// MaxIteracoesPadrao = 4, não 1.
	fmt.Println("===== Cenário 3: não convergência, escalonamento (limiar forçado a 1 volta pra demo confiável) =====")
	protocolo3 := "Estudo fase III, população adulta, diabetes tipo 2, sem envolvimento de menores de idade."
	fmt.Println("[Mariana submete o protocolo ao Gateway]")
	fmt.Println("  ", protocolo3)
	fmt.Println()

	resultado3, err := AgenteICFComLimite(ctx, client, protocolo3, 1)
	if err != nil {
		return err
	}
	ImprimirTrilha(resultado3.Trilha)
	if resultado3.EscalarParaAprovacaoHumana {
		fmt.Println("[Orquestrador] Loop não convergiu —", resultado3.Motivo)
		fmt.Println("[Sistema] Encaminhando ao Approval Gate para intervenção manual — nunca falha silenciosamente.")
	} else {
		fmt.Println("[Atenção] O modelo convergiu na única volta permitida antes mesmo de precisar escalar.")
	}

	return nil
}
