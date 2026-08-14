// Cinco mini-demonstracoes, uma por peca da anatomia de um agente unico
// (Memoria, Planejamento, Ferramentas, Acao, Approval Gate) — cada uma
// isolada e rodavel sozinha, sem o loop ReAct inteiro (isso e
// react-agent-prototype-go) e sem schema formal de ferramenta.
//
// Contexto: TrialForge, Agente ICF (gera a secao de assentimento do Termo
// de Consentimento a partir do protocolo do estudo).
//
// Porte Go de agent-components-demo.js / agent_components_demo.py.
// Ver README.md deste projeto para detalhes de paridade e adaptacoes.
package main

import (
	"context"
	"fmt"

	"agent-components-demo/internal/agente"
	"agent-components-demo/internal/ollama"
)

const modelo = "gemma4:e2b"

// chatClientAdapter adapta *ollama.Client (que fala em []ollama.Message)
// para a interface agente.ChatClient (que fala em []agente.Mensagem).
type chatClientAdapter struct {
	client *ollama.Client
}

func (a chatClientAdapter) Chat(ctx context.Context, mensagens []agente.Mensagem) (string, error) {
	convertidas := make([]ollama.Message, len(mensagens))
	for i, m := range mensagens {
		convertidas[i] = ollama.Message{Role: m.Role, Content: m.Content}
	}
	return a.client.Chat(ctx, convertidas)
}

func main() {
	loadDotEnv(".env")

	rodarTestes()
	demonstrarMemoria()
	demonstrarPlanejamento()
	demonstrarFerramentas()
	demonstrarAcaoEGate()
}

// ============================================================
// Testes automatizados (mesma disciplina do original: so a parte
// deterministica ganha assert; Planejamento fica pra observacao ao vivo)
// ============================================================

func rodarTestes() {
	testarMemoria()
	testarFerramenta()
	testarGate()
}

func testarMemoria() {
	fmt.Println("== Testes: memória de longo prazo persiste entre chamadas ==")
	banco := agente.NovaMemoriaLongoPrazo()
	primeira := banco.Registrar("usuario-teste", "prefere respostas curtas")
	segunda := banco.Registrar("usuario-teste", "prefere tom formal")
	ok := segunda.Interacoes == 2 && len(segunda.Preferencias) == 2 && primeira.Interacoes == 1
	fmt.Printf("  [%s] duas chamadas com o mesmo usuarioId acumulam estado\n", okFalhou(ok))
	if !ok {
		panic("memoriaLongoPrazo não está persistindo corretamente entre chamadas")
	}
	fmt.Println()
}

func testarFerramenta() {
	fmt.Println("== Testes: buscarClausulaAssentimento (determinístico) ==")
	menor := agente.BuscarClausulaAssentimento([]int{12, 15, 17})
	adulto := agente.BuscarClausulaAssentimento([]int{25, 40, 55})
	okMenor := menor.Texto != ""
	okAdulto := adulto.Texto == ""
	fmt.Printf("  [%s] faixa com menor de idade encontra cláusula\n", okFalhou(okMenor))
	fmt.Printf("  [%s] faixa só de adultos não encontra cláusula\n", okFalhou(okAdulto))
	if !okMenor || !okAdulto {
		panic("buscarClausulaAssentimento não está classificando corretamente")
	}
	fmt.Println()
}

func testarGate() {
	fmt.Println("== Testes: executarOuGatear ==")
	semGate := agente.ExecutarOuGatear(agente.AcaoProposta{
		Tipo: "x", RequerAprovacao: false, Executar: func() string { return "feito" },
	})
	comGate := agente.ExecutarOuGatear(agente.AcaoProposta{
		Tipo: "y", RequerAprovacao: true, Executar: func() string { return "nunca deveria rodar" },
	})
	okSemGate := semGate.Status == "executada" && semGate.Resultado == "feito"
	okComGate := comGate.Status == "aguardando_aprovacao" && comGate.Resultado == ""
	fmt.Printf("  [%s] ação sem gate executa direto\n", okFalhou(okSemGate))
	fmt.Printf("  [%s] ação com gate nunca chega a executar\n", okFalhou(okComGate))
	if !okSemGate || !okComGate {
		panic("executarOuGatear não está bloqueando/liberando corretamente")
	}
	fmt.Println()
}

func okFalhou(ok bool) string {
	if ok {
		return "OK"
	}
	return "FALHOU"
}

// ============================================================
// Demonstracoes (saida em console, replicando o original)
// ============================================================

func demonstrarMemoria() {
	fmt.Println("===== 1. MEMÓRIA =====")

	contexto := agente.MemoriaCurtoPrazo("Estudo fase II, público-alvo 12-17 anos, terapia oncológica experimental.")
	fmt.Printf("[Curto prazo] Contexto construído para esta requisição: %d mensagem(ns).\n", len(contexto))
	fmt.Println("  -> O Agente ICF do TrialForge só precisa disso: cada protocolo é um caso novo,")
	fmt.Println("     memória de longo prazo aqui seria custo sem benefício (canvas do Módulo 2.1).")
	fmt.Println()

	fmt.Println("[Longo prazo] Simulando um assistente diferente, que acompanha o mesmo usuário ao longo do tempo:")
	banco := agente.NovaMemoriaLongoPrazo()
	fmt.Printf("  1ª interação: %+v\n", banco.Registrar("usuario-42", "prefere respostas curtas"))
	fmt.Printf("  2ª interação: %+v\n", banco.Registrar("usuario-42", "prefere tom formal"))
	fmt.Println("  -> Repare: a segunda chamada já sabe da primeira. Isso só vale a pena quando")
	fmt.Println("     a tarefa se repete com o mesmo contexto ao longo do tempo — o oposto do ICF.")
	fmt.Println()
}

func demonstrarPlanejamento() {
	fmt.Println("===== 2. PLANEJAMENTO =====")
	pergunta := "Um estudo com público-alvo de 12 a 17 anos precisa de assentimento do participante, " +
		"além do consentimento do responsável?"

	baseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	client := chatClientAdapter{client: ollama.NewClient(baseURL, modelo)}
	ctx := context.Background()

	cot, err := agente.ChainOfThought(ctx, client, pergunta)
	if err != nil {
		imprimirErroModelo(err)
		fmt.Println()
		return
	}
	fmt.Printf("[Chain-of-thought] %d chamada ao modelo, %dms.\n", cot.Chamadas, cot.DuracaoMs)

	reflexao, err := agente.ChainOfThoughtMaisReflexao(ctx, client, pergunta)
	if err != nil {
		imprimirErroModelo(err)
		fmt.Println()
		return
	}
	fmt.Printf("[+ Reflexão]        %d chamadas ao modelo, %dms.\n", reflexao.Chamadas, reflexao.DuracaoMs)

	razao := 0.0
	if cot.DuracaoMs > 0 {
		razao = float64(reflexao.DuracaoMs) / float64(cot.DuracaoMs)
	}
	fmt.Printf("  -> Reflexão levou %.1fx mais tempo que chain-of-thought sozinho,\n", razao)
	fmt.Println("     porque é uma chamada inteira a mais, não só um raciocínio mais longo na mesma chamada.")
	fmt.Println("     Essa é a distinção de custo que o canvas do Módulo 2.1 registra.")
	fmt.Println()
}

func imprimirErroModelo(err error) {
	fmt.Println("[Erro] Não foi possível chamar o modelo:", err)
	fmt.Printf("Verifique se o Ollama está rodando ('ollama serve') e o modelo baixado ('ollama pull %s').\n", modelo)
}

func demonstrarFerramentas() {
	fmt.Println("===== 3. FERRAMENTAS =====")
	fmt.Println("[Caso 1] Estudo com participantes de 12 a 17 anos:")
	fmt.Println("  ", agente.BuscarClausulaAssentimento([]int{12, 15, 17}))
	fmt.Println("[Caso 2] Estudo só com adultos:")
	fmt.Println("  ", agente.BuscarClausulaAssentimento([]int{25, 40, 55}))
	fmt.Println("  -> O agente não \"sabe\" essa regra de cor: ele delega pra uma função")
	fmt.Println("     determinística, porque é isso que sistemas determinísticos fazem melhor.")
	fmt.Println()
}

func demonstrarAcaoEGate() {
	fmt.Println("===== 4 e 5. AÇÃO + APPROVAL GATE =====")

	gerarRascunho := agente.AcaoProposta{
		Tipo:            "gerar_rascunho_icf",
		RequerAprovacao: false,
		Executar:        func() string { return "Rascunho da seção de assentimento gerado." },
	}
	fmt.Printf("[Ação 1: gerar rascunho]        %+v\n", agente.ExecutarOuGatear(gerarRascunho))

	notificarEventoAdverso := agente.AcaoProposta{
		Tipo:            "notificar_evento_adverso_regulatorio",
		RequerAprovacao: true,
		Executar:        func() string { return "Notificação enviada à ANVISA." },
	}
	fmt.Printf("[Ação 2: notificar evento adverso] %+v\n", agente.ExecutarOuGatear(notificarEventoAdverso))

	fmt.Println("  -> Mesma \"capacidade de ação\" nos dois casos. A diferença que decide se")
	fmt.Println("     executa direto ou pausa pro Approval Gate não é técnica, é a Pergunta 2")
	fmt.Println("     do Módulo 1.3: o erro é caro E irreversível?")
}
