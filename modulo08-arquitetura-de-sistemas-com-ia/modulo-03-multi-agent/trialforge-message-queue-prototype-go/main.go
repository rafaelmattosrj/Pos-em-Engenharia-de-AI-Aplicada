// Porte de main() (JS/Python) — demonstracao narrada, igual ao Slide 3 +
// paragrafos 68-73 do TP.
//
// A suite de "testes automatizados" que o original roda dentro do main()
// virou testes idiomaticos em Go (trialforge/fluxo_test.go) — rode com
// `go test ./...`. Este main roda so a demonstracao dos 5 cenarios.
//
// Uso: go run .
package main

import (
	"fmt"
	"strings"

	"trialforge-message-queue-prototype/trialforge"
)

func main() {
	fmt.Println("===== TRIALFORGE: FILA DE MENSAGENS (Barramento sobre goroutines/channels) =====")
	fmt.Println()

	fmt.Println("[Cenário 1] Caminho feliz — Protocolo primeiro, depois ICF+CSR em paralelo:")
	feliz, err := trialforge.RodarFluxoTrialForge("Estudo fase II, 12-17 anos", trialforge.OpcoesFluxo{})
	falhar(err)
	fmt.Printf("  Evento protocolo:pronto: %+v\n", feliz.DadoProtocolo)
	fmt.Printf("  Resultado ICF: %+v\n", *feliz.Reacao.ResultadoICF)
	fmt.Printf("  Resultado CSR: %+v\n", *feliz.Reacao.ResultadoCSR)
	fmt.Println("  Decisão do Supervisor:", feliz.DecisaoSupervisor)
	fmt.Println("  Ordem real de execução:", strings.Join(feliz.OrdemDeExecucao, " -> "))
	fmt.Println()

	fmt.Println("[Cenário 2] ICF falha, usando EstrategiaAll (o bug do parágrafo 68-71):")
	bug, err := trialforge.RodarFluxoTrialForge("Estudo fase II, 12-17 anos", trialforge.OpcoesFluxo{
		Estrategia:     trialforge.EstrategiaAll,
		ForcarFalhaICF: true,
	})
	falhar(err)
	fmt.Println("  ok:", bug.Reacao.Ok, "| resultado do CSR preservado?", bug.Reacao.ResultadoCSRPreservado)
	fmt.Println("  -> O CSR terminou bem, mas a EstrategiaAll descartou o lote inteiro: o resultado dele se perdeu.")
	fmt.Println()

	fmt.Println("[Cenário 3] Mesmo cenário, corrigido com EstrategiaAllSettled — CAP + idempotência (parágrafo 72-73):")
	corrigido, err := trialforge.RodarFluxoTrialForge("Estudo fase II, 12-17 anos", trialforge.OpcoesFluxo{
		Estrategia:     trialforge.EstrategiaAllSettled,
		ForcarFalhaICF: true,
	})
	falhar(err)
	fmt.Println("  Resultado do CSR preservado?", corrigido.Reacao.ResultadoCSRPreservado)
	fmt.Println("  Decisão do Supervisor:", corrigido.DecisaoSupervisor)
	fmt.Printf("  Documentos de ICF gravados: %d (%d tentativas registradas nesse único documento)\n",
		corrigido.Documentos.Tamanho(), corrigido.TentativasICF)
	fmt.Println("  -> O Supervisor não só decidiu o que refazer: refez de verdade, e o registro idempotente")
	fmt.Println("     garantiu que a segunda tentativa não gerou um segundo documento de ICF.")
	fmt.Println()

	fmt.Println("[Cenário 4] Emenda do comitê de ética chega no meio do caminho — Saga de verdade:")
	divergencia, err := trialforge.RodarFluxoTrialForge("Estudo com emenda ética", trialforge.OpcoesFluxo{
		ComEmendaEtica: true,
	})
	falhar(err)
	fmt.Printf("  Verificação de consistência: %+v\n", *divergencia.Verificacao)
	fmt.Println("  Decisão do Supervisor:", divergencia.DecisaoSupervisor)
	fmt.Println("  Resultado final do ICF (nunca tocado):", divergencia.Reacao.ResultadoICF.Secao)
	fmt.Println("  Resultado final do CSR (regenerado):", divergencia.Reacao.ResultadoCSR.Sintese)
	fmt.Printf("  Histórico de revisões do protocolo: %+v\n", divergencia.EstadoProtocolo.Historico())
	fmt.Println("  -> O ICF terminou ANTES da emenda chegar — seu resultado já nasceu correto, nunca precisou")
	fmt.Println("     ser refeito. O CSR terminou DEPOIS da emenda, mas com o critério antigo — o Supervisor")
	fmt.Println("     detectou e regenerou só a parte dele: compensação Saga, desfazer só o que precisa.")
	fmt.Println()

	fmt.Println("[Cenário 5] CAP completo — timeout real, retry com limite, e o limite esgotando:")
	travado, err := trialforge.RodarFluxoTrialForge("Estudo com agente travado", trialforge.OpcoesFluxo{
		ForcarTravamentoICF: true,
	})
	falhar(err)
	fmt.Println("  Erro do despacho original:", travado.Reacao.ErroICF)
	fmt.Printf("  Decisão do Supervisor: %s (%d tentativa(s) de retry)\n", travado.DecisaoSupervisor, travado.TentativasRetry)
	fmt.Println("  -> O ICF travou de verdade (corrida real contra o timeout), não retornou um erro —")
	fmt.Println("     e o retry resolveu na primeira tentativa, porque a causa raiz era transitória.")
	fmt.Println()

	esgotado, err := trialforge.RodarFluxoTrialForge("Estudo com falha persistente", trialforge.OpcoesFluxo{
		ForcarTravamentoICF:   true,
		PersistirFalhaNoRetry: true,
	})
	falhar(err)
	fmt.Println("  Agora com falha PERSISTENTE (não transitória):")
	fmt.Printf("  Decisão do Supervisor: %s (%d tentativa(s) de retry)\n", esgotado.DecisaoSupervisor, esgotado.TentativasRetry)
	fmt.Println("  Reação final ok?", esgotado.Reacao.Ok)
	fmt.Printf("  -> %d tentativas esgotadas, o Supervisor desiste do ICF e segue em frente sem esse\n", trialforge.MaxTentativas)
	fmt.Println("     resultado — Disponibilidade sobre Consistência (Teorema CAP).")
	fmt.Println()
}

func falhar(err error) {
	if err != nil {
		panic(err)
	}
}
