package gateway

import (
	"context"
	"fmt"
	"io"
)

// ChatStreamer é o segundo (e único outro) contrato de rede do gateway:
// gerar uma resposta em streaming, token a token — onChunk é chamado uma vez
// por pedaço de texto recebido, na ordem em que chegam.
type ChatStreamer interface {
	ChatStream(ctx context.Context, model string, messages []Message, onChunk func(content string)) error
}

// Approver é o contrato do Approval Gate (Módulo 4.4) — implementado por
// *ApprovalGate em produção; testes de Processor usam um stub que não
// depende de stdin.
type Approver interface {
	PedirAprovacaoHumana(rascunho string) (bool, error)
}

// Processor é o Gateway propriamente dito: os 4 grupos de padrão do Módulo 4
// encadeados numa requisição só (RAG, Intent-Based Routing + Model Router,
// Semantic Cache + Response Streaming, Confidence Threshold + Approval Gate +
// Audit Trail).
type Processor struct {
	Embedder   Embedder
	Chat       ChatStreamer
	Preparados map[string]IndicePreparado
	Cache      *SemanticCache
	Audit      *AuditTrail
	Approval   Approver
	Out        io.Writer

	ModeloBarato    string
	ModeloCaro      string
	LimiarCache     float64
	LimiarConfianca float64

	proximoID int
}

func (p *Processor) logf(format string, args ...interface{}) {
	if p.Out == nil {
		return
	}
	fmt.Fprintf(p.Out, format+"\n", args...)
}

// ProcessarRequisicao executa o pipeline completo pra uma pergunta: classifica
// a intenção, consulta o Semantic Cache, roteia pro modelo e índice certos,
// busca a cláusula via RAG agêntico, gera a resposta em streaming e, se
// necessário, escala pro Approval Gate — registrando cada decisão na trilha
// de auditoria. Devolve (nil, nil) quando o rascunho é rejeitado no gate,
// igual ao `return null`/`return None` dos originais.
func (p *Processor) ProcessarRequisicao(ctx context.Context, pergunta string) (*string, error) {
	p.proximoID++
	idRequisicao := fmt.Sprintf("req-%d", p.proximoID)
	p.logf("\n[Gateway] Requisição recebida: %q", pergunta)

	intencao := ClassificarIntencao(pergunta)
	p.logf("[Intent-Based Routing] Intenção classificada: %s", intencao)

	perguntaEmbedding, err := p.Embedder.Embedar(ctx, pergunta)
	if err != nil {
		return nil, fmt.Errorf("embedding da pergunta: %w", err)
	}

	// Semantic Cache: só entra em perguntas de rotina — síntese de CSR nunca
	// usa cache (Módulo 4.3)
	if intencao != "sintese_csr" {
		resultadoCache := p.Cache.Consultar(perguntaEmbedding)
		if resultadoCache.Entrada != nil && resultadoCache.Similaridade >= p.LimiarCache {
			p.logf("[Semantic Cache] HIT (similaridade %.3f) — resposta reaproveitada, sem chamar o modelo.", resultadoCache.Similaridade)
			if err := p.Audit.Registrar(map[string]interface{}{
				"id_requisicao":      idRequisicao,
				"pergunta":           pergunta,
				"intencao":           intencao,
				"cache_hit":          true,
				"similaridade_cache": resultadoCache.Similaridade,
				"status_final":       "respondido_via_cache",
			}); err != nil {
				return nil, err
			}
			resposta := resultadoCache.Entrada.Resposta
			return &resposta, nil
		}
		p.logf("[Semantic Cache] MISS (melhor similaridade %.3f) — segue pro modelo.", resultadoCache.Similaridade)
	}

	// Model Router (Módulo 4.2): a intenção decide qual modelo processa
	modelo := p.ModeloBarato
	tipoModelo := "barato/rápido"
	if intencao == "sintese_csr" {
		modelo = p.ModeloCaro
		tipoModelo = "caro/capaz"
	}
	p.logf("[Model Router] Modelo escolhido: %s (%s)", modelo, tipoModelo)

	// RAG (Módulo 4.1): Multi-Index roteia pro domínio certo, Hybrid Search
	// busca dentro dele (BM25 + embedding, fundidos por RRF), Agentic RAG
	// insiste com estratégia mais ampla se a confiança vier baixa.
	indiceInicial := IndicePorIntencao[intencao]
	p.logf("[Multi-Index] Roteando pro índice %q (Módulo 4.1)", indiceInicial)
	resultadoRag := BuscarClausulaAgentica(p.Preparados, pergunta, perguntaEmbedding, indiceInicial, p.LimiarConfianca, p.logf)
	confianca := resultadoRag.SimilaridadeCosseno
	clausula := resultadoRag.Clausula
	p.logf(
		"[RAG] Cláusula final: %q — índice %q, cosseno %.3f, BM25 %.3f, RRF %.4f, %d iteração(ões).",
		clausula.Tema, resultadoRag.Indice, confianca, resultadoRag.ScoreBM25, resultadoRag.ScoreRRF, resultadoRag.IteracoesUsadas,
	)

	// Geração com streaming (Módulo 4.3) — token a token, não espera tudo pronto
	p.logf("[Response Streaming] Gerando resposta:")
	fmt.Fprint(p.Out, "    ")
	rascunho := ""
	mensagens := []Message{
		{
			Role: "system",
			Content: "Você redige respostas curtas e precisas sobre regras de estudos clínicos, " +
				"citando a fonte regulatória fornecida.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Pergunta: %s\n\nCláusula regulatória relevante: %s\nFonte: %s\n\nResponda usando essa cláusula.",
				pergunta, clausula.Texto, clausula.Fonte,
			),
		},
	}
	err = p.Chat.ChatStream(ctx, modelo, mensagens, func(content string) {
		fmt.Fprint(p.Out, content)
		rascunho += content
	})
	fmt.Fprintln(p.Out)
	if err != nil {
		return nil, fmt.Errorf("geração de resposta: %w", err)
	}

	// Confidence Threshold + Approval Gate (Módulo 4.4)
	precisaAprovacao := intencao == "sintese_csr" || confianca < p.LimiarConfianca
	aprovado := true
	if precisaAprovacao {
		var motivo string
		if intencao == "sintese_csr" {
			motivo = "síntese de CSR — erro caro e irreversível, gate sempre obrigatório (Módulo 1.3)"
		} else {
			motivo = fmt.Sprintf("confiança abaixo do limiar (%.3f < %v)", confianca, p.LimiarConfianca)
		}
		p.logf("[Confidence Threshold] Escalando pro Approval Gate — motivo: %s", motivo)
		// Registro gravado ANTES do prompt, não depois: o pedido pendente
		// precisa sobreviver independente de quando (ou se) alguém responde.
		if err := p.Audit.Registrar(map[string]interface{}{
			"id_requisicao": idRequisicao,
			"pergunta":      pergunta,
			"intencao":      intencao,
			"status_final":  "aguardando_aprovacao",
			"motivo_gate":   motivo,
		}); err != nil {
			return nil, err
		}
		aprovado, err = p.Approval.PedirAprovacaoHumana(rascunho)
		if err != nil {
			return nil, fmt.Errorf("approval gate: %w", err)
		}
	} else {
		p.logf("[Confidence Threshold] Confiança %.3f acima do limiar — segue sem Approval Gate.", confianca)
	}

	statusFinal := "aprovado"
	if !aprovado {
		statusFinal = "rejeitado"
	}
	if err := p.Audit.Registrar(map[string]interface{}{
		"id_requisicao":     idRequisicao,
		"pergunta":          pergunta,
		"intencao":          intencao,
		"cache_hit":         false,
		"modelo_usado":      modelo,
		"indice_usado":      resultadoRag.Indice,
		"iteracoes_agentic": resultadoRag.IteracoesUsadas,
		"esgotou_agentic":   resultadoRag.EsgotouLimite,
		"confianca_rag":     confianca,
		"gate_acionado":     precisaAprovacao,
		"aprovado":          aprovado,
		"status_final":      statusFinal,
	}); err != nil {
		return nil, err
	}

	if !aprovado {
		p.logf("[Approval Gate] Rascunho rejeitado — não vira oficial.")
		return nil, nil
	}

	// Alimenta o Semantic Cache com essa pergunta+resposta pra próxima vez (só rotina)
	if intencao != "sintese_csr" {
		p.Cache.Adicionar(pergunta, perguntaEmbedding, rascunho)
	}

	p.logf("[Gateway] Requisição concluída — trilha de auditoria registrada.")
	return &rascunho, nil
}
