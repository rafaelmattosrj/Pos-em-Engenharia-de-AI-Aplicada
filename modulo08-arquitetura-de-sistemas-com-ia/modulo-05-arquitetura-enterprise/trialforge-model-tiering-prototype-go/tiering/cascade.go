package tiering

import (
	"context"
	"fmt"
)

const (
	ModeloTier1     = "gemma4:e2b"    // barato — tentado primeiro, sempre
	ModeloTier2     = "gemma4:latest" // caro — escalação ou regra fixa (CSR)
	ModeloEmbedding = "nomic-embed-text"

	LimiarCascataBusca    = 0.75
	LimiarCascataResposta = 0.75

	// Custo estimado por chamada, só pra demonstrar o controle de orçamento —
	// valores ilustrativos, não preço real de nenhum provedor.
	CustoTier1 = 0.001
	CustoTier2 = 0.01
)

// CascadeGateway implementa a cascata de Model Tiering (Módulo 5.4): tenta o
// tier mais barato primeiro, escala pro tier caro se a confiança de BUSCA
// (achou a cláusula certa?) OU a confiança de RESPOSTA (g(pergunta, resposta)
// — a resposta gerada ficou fiel à cláusula que recebeu?) ficar abaixo do
// limiar. Síntese de CSR é regra fixa: pula direto pro tier caro e aciona o
// Approval Gate.
//
// Porte 1:1 de processarComCascata em trialforge-model-tiering-prototype.js / .py.
type CascadeGateway struct {
	Ollama         Gateway
	RagIndex       *RagIndex
	Orcamento      *OrcamentoManager
	ApprovalPrompt ApprovalPrompt
	AuditTrail     *AuditTrail
}

func NewCascadeGateway(ollama Gateway, ragIndex *RagIndex, orcamento *OrcamentoManager,
	approvalPrompt ApprovalPrompt, auditTrail *AuditTrail) *CascadeGateway {
	return &CascadeGateway{
		Ollama:         ollama,
		RagIndex:       ragIndex,
		Orcamento:      orcamento,
		ApprovalPrompt: approvalPrompt,
		AuditTrail:     auditTrail,
	}
}

func (g *CascadeGateway) ProcessarComCascata(ctx context.Context, pergunta, estudoID string) (string, error) {
	fmt.Printf("\n[Gateway] Estudo \"%s\" — requisição: \"%s\"\n", estudoID, pergunta)

	intencao := ClassificarIntencao(pergunta)
	custoMaximoPossivel := CustoTier1 + CustoTier2
	if intencao == SinteseCSR {
		custoMaximoPossivel = CustoTier2
	}

	// Orçamento reservado (não só checado) ANTES de qualquer chamada de modelo.
	reservou, err := g.Orcamento.ReservarOrcamento(estudoID, custoMaximoPossivel)
	if err != nil {
		return "", err
	}
	if !reservou {
		fmt.Printf("[Orçamento] BLOQUEADO — estudo \"%s\" ultrapassaria o limite antes mesmo de chamar o modelo.\n", estudoID)
		if err := g.AuditTrail.Registrar(map[string]any{
			"estudoId":     estudoID,
			"pergunta":     pergunta,
			"status_final": "bloqueado_por_orcamento",
		}); err != nil {
			return "", err
		}
		return "", nil
	}
	custoRealUsado := 0.0

	perguntaEmbedding, err := g.Ollama.Embed(ctx, ModeloEmbedding, pergunta)
	if err != nil {
		return "", err
	}
	resultadoBusca := g.RagIndex.BuscarClausula(perguntaEmbedding)
	confiancaBusca := resultadoBusca.Similaridade
	clausula := resultadoBusca.Clausula
	indiceClausula := resultadoBusca.Indice

	var tierUsado, rascunho string
	escalou := false
	var confiancaResposta *float64

	if intencao == SinteseCSR {
		// Regra fixa (Módulo 1.3): erro caro e irreversível, pula direto pro tier caro.
		fmt.Println("[Model Tiering] Síntese de CSR — regra fixa, direto pro Tier 2 (caro).")
		tierUsado = "Tier 2"
		rascunho, err = g.gerarComTier(ctx, ModeloTier2, pergunta, clausula)
		if err != nil {
			return "", err
		}
		custoRealUsado += CustoTier2
	} else {
		fmt.Printf("[Model Tiering] Tentando Tier 1 (barato) primeiro — cláusula encontrada com confiança de busca %.3f\n", confiancaBusca)
		tierUsado = "Tier 1"
		rascunho, err = g.gerarComTier(ctx, ModeloTier1, pergunta, clausula)
		if err != nil {
			return "", err
		}
		custoRealUsado += CustoTier1

		cr, err := g.RagIndex.CalcularConfiancaResposta(ctx, rascunho, indiceClausula)
		if err != nil {
			return "", err
		}
		confiancaResposta = &cr
		fmt.Printf("[Model Tiering] Confiança da resposta do Tier 1 — g(pergunta, resposta): %.3f\n", cr)

		buscaFalhou := confiancaBusca < LimiarCascataBusca
		respostaFalhou := cr < LimiarCascataResposta

		if buscaFalhou || respostaFalhou {
			motivo := "confiança de busca"
			switch {
			case buscaFalhou && respostaFalhou:
				motivo = "busca e resposta"
			case respostaFalhou:
				motivo = "confiança de resposta"
			}
			fmt.Printf("[Model Tiering] Escalando pro Tier 2 — motivo: %s abaixo do limiar (busca %.3f, resposta %.3f).\n",
				motivo, confiancaBusca, cr)
			escalou = true
			tierUsado = "Tier 2 (escalado)"
			rascunho, err = g.gerarComTier(ctx, ModeloTier2, pergunta, clausula)
			if err != nil {
				return "", err
			}
			custoRealUsado += CustoTier2
			cr2, err := g.RagIndex.CalcularConfiancaResposta(ctx, rascunho, indiceClausula)
			if err != nil {
				return "", err
			}
			confiancaResposta = &cr2
		} else {
			fmt.Println("[Model Tiering] Busca e resposta acima do limiar — Tier 1 resolve, sem escalar.")
		}
	}

	// A reserva cobriu o pior caso; devolve o que sobrou se o custo real ficou menor.
	if err := g.Orcamento.LiberarSobra(estudoID, custoMaximoPossivel-custoRealUsado); err != nil {
		return "", err
	}

	// Estreitamento de escopo deliberado em relação ao Módulo 4.5: aqui só CSR aciona
	// o Approval Gate — uma escalação de cascata por confiança baixa só ajusta o tier.
	precisaAprovacao := intencao == SinteseCSR
	aprovado := true
	if precisaAprovacao {
		aprovado, err = g.ApprovalPrompt.Approve(rascunho)
		if err != nil {
			return "", err
		}
	}

	registro := map[string]any{
		"estudoId":         estudoID,
		"pergunta":         pergunta,
		"intencao":         intencao,
		"tier_usado":       tierUsado,
		"escalou_cascata":  escalou,
		"confianca_busca":  confiancaBusca,
		"gasto_acumulado":  g.Orcamento.GastoAtual(estudoID),
		"orcamento_limite": g.Orcamento.Limite(estudoID),
		"aprovado":         aprovado,
		"status_final":     statusFinal(aprovado),
	}
	if confiancaResposta != nil {
		registro["confianca_resposta"] = *confiancaResposta
	} else {
		registro["confianca_resposta"] = nil
	}
	if err := g.AuditTrail.Registrar(registro); err != nil {
		return "", err
	}

	fmt.Printf("[Orçamento] Estudo \"%s\": gasto acumulado %.4f / limite %v\n",
		estudoID, g.Orcamento.GastoAtual(estudoID), g.Orcamento.Limite(estudoID))
	return rascunho, nil
}

func statusFinal(aprovado bool) string {
	if aprovado {
		return "aprovado"
	}
	return "rejeitado"
}

func (g *CascadeGateway) gerarComTier(ctx context.Context, modelo, pergunta string, clausula Clausula) (string, error) {
	system := "Você redige respostas curtas e precisas sobre regras de estudos clínicos, " +
		"citando a fonte regulatória fornecida."
	user := fmt.Sprintf("Pergunta: %s\n\nCláusula regulatória relevante: %s\nFonte: %s\n\nResponda usando essa cláusula.",
		pergunta, clausula.Texto, clausula.Fonte)
	rascunho, err := g.Ollama.ChatStream(ctx, modelo, system, user, func(pedaco string) {
		fmt.Print(pedaco)
	})
	if err != nil {
		return "", err
	}
	fmt.Println()
	return rascunho, nil
}
