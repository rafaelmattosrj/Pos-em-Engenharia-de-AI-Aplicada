// Porte Go de trialforge-model-tiering-prototype.js / trialforge_model_tiering_prototype.py.
// Ver README.md deste projeto para detalhes de paridade e adaptações.
//
// `go run . ` roda a demo gravada (4 chamadas sequenciais).
// `go run . --volume` roda o extra de volume concorrente (Missão Prática) —
// não precisa de aprovação humana (sem CSR na mistura).
package main

import (
	"context"
	"fmt"
	"os"

	"trialforge-model-tiering-prototype/ollama"
	"trialforge-model-tiering-prototype/tiering"
)

func main() {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := ollama.NewClient(baseURL)
	ragIndex := tiering.NewRagIndex(client, tiering.ModeloEmbedding)
	orcamento := tiering.NewOrcamentoManager()
	auditTrail := tiering.NewAuditTrail("audit-trail-tiering.jsonl")
	ctx := context.Background()

	rodarVolume := len(os.Args) > 1 && os.Args[1] == "--volume"

	if err := ragIndex.PrepararIndice(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "[Erro não tratado]", err)
		os.Exit(1)
	}

	var err error
	if rodarVolume {
		err = rodarVolume2(ctx, client, ragIndex, orcamento, auditTrail)
	} else {
		orcamento.DefinirOrcamento("estudo-A", 0.05)
		orcamento.DefinirOrcamento("estudo-B", 0.005)
		gateway := tiering.NewCascadeGateway(client, ragIndex, orcamento,
			tiering.NewStdinApprovalPrompt(os.Stdin, os.Stdout), auditTrail)
		err = rodarDemo(ctx, gateway)
		if err == nil {
			err = tiering.VerificarEImprimir(auditTrail)
		}
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "[Erro não tratado]", err)
		os.Exit(1)
	}
}

func rodarDemo(ctx context.Context, gateway *tiering.CascadeGateway) error {
	// 1) Estudo A, pergunta que o Tier 1 resolve bem sozinho
	if _, err := gateway.ProcessarComCascata(ctx, "Quais são as regras de assentimento pra menores nesse estudo?", "estudo-A"); err != nil {
		return err
	}

	// 2) Estudo A, pergunta fora do assunto do banco de cláusulas — escala pro Tier 2
	if _, err := gateway.ProcessarComCascata(ctx, "Qual é o prazo de validade dos exames laboratoriais desse estudo?", "estudo-A"); err != nil {
		return err
	}

	// 3) Estudo A de novo, mas agora síntese de CSR — regra fixa, direto pro Tier 2
	if _, err := gateway.ProcessarComCascata(ctx, "Preciso da síntese do CSR final desse estudo.", "estudo-A"); err != nil {
		return err
	}

	// 4) Estudo B, com orçamento propositalmente baixo — deve bloquear antes de chamar qualquer modelo
	if _, err := gateway.ProcessarComCascata(ctx, "Quais são as regras de assentimento pra menores nesse estudo?", "estudo-B"); err != nil {
		return err
	}

	return nil
}

// rodarVolume2 (nome evita colidir com a flag --volume) roda a Missão
// Prática de volume concorrente.
func rodarVolume2(ctx context.Context, client *ollama.Client, ragIndex *tiering.RagIndex,
	orcamento *tiering.OrcamentoManager, auditTrail *tiering.AuditTrail) error {
	const perguntaAssentimento = "Quais são as regras de assentimento pra menores nesse estudo?"
	const perguntaRetirada = "O participante pode desistir do estudo a qualquer momento?"
	const perguntaForaDominio = "Qual é o prazo de validade dos exames laboratoriais desse estudo?"

	orcamento.DefinirOrcamento("estudo-C", 0.05)
	orcamento.DefinirOrcamento("estudo-D", 0.05)
	orcamento.DefinirOrcamento("estudo-E", 0.05)
	// Orçamento apertado de propósito: reservarOrcamento reserva o PIOR caso por requisição
	// (CustoTier1+CustoTier2 = 0.011) — esse limite cabe exatamente 2 reservas, com 5
	// requisições concorrentes disputando o mesmo estudo.
	orcamento.DefinirOrcamento("estudo-F", 2*(tiering.CustoTier1+tiering.CustoTier2)+0.0005)

	// Approval Gate não é acionado (sem CSR na mistura), então um ApprovalPrompt que
	// sempre aprova nunca chega a ser usado de fato.
	semprAprova := tiering.ApprovalPromptFunc(func(rascunho string) (bool, error) { return true, nil })
	gateway := tiering.NewCascadeGateway(client, ragIndex, orcamento, semprAprova, auditTrail)

	var requisicoes []func()
	for _, estudoID := range []string{"estudo-C", "estudo-D", "estudo-E"} {
		estudoID := estudoID
		requisicoes = append(requisicoes,
			func() { gateway.ProcessarComCascata(ctx, perguntaAssentimento, estudoID) },
			func() { gateway.ProcessarComCascata(ctx, perguntaRetirada, estudoID) },
			func() { gateway.ProcessarComCascata(ctx, perguntaAssentimento, estudoID) },
			func() { gateway.ProcessarComCascata(ctx, perguntaForaDominio, estudoID) },
		)
	}
	for i := 0; i < 5; i++ {
		requisicoes = append(requisicoes, func() { gateway.ProcessarComCascata(ctx, perguntaAssentimento, "estudo-F") })
	}

	fmt.Printf("\n== Volume concorrente: %d requisições, 4 estudos, disparadas ao mesmo tempo ==\n\n", len(requisicoes))
	resultados := tiering.SimularVolumeConcorrente(orcamento, []string{"estudo-C", "estudo-D", "estudo-E", "estudo-F"}, requisicoes)

	fmt.Println("\n== Resultado por estudo, depois da concorrência ==")
	algumEstourou := false
	for _, r := range resultados {
		status := "OK"
		if r.Estourou {
			status = "ESTOUROU"
			algumEstourou = true
		}
		fmt.Printf("  [%s] %s: gasto %.4f / limite %.4f\n", status, r.EstudoID, r.Gasto, r.Limite)
	}
	if algumEstourou {
		return fmt.Errorf("volume concorrente estourou orçamento de pelo menos um estudo — corrija ReservarOrcamento")
	}
	fmt.Println("\n[OK] Nenhum estudo gastou além do limite, mesmo com requisições simultâneas pro mesmo estudo.")
	return nil
}
