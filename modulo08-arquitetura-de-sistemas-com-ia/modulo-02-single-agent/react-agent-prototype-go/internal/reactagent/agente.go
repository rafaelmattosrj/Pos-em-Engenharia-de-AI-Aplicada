package reactagent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MaxIteracoesPadrao e o limite calibrado do loop ReAct: 4 voltas.
const MaxIteracoesPadrao = 4

// PassoTrilha e uma entrada da trilha de DESENVOLVIMENTO (uma por volta do
// loop, com o que o modelo decidiu e quanto tempo levou) — nao e
// observabilidade de producao, so pra debugar enquanto se constroi.
type PassoTrilha struct {
	Volta      int
	Acao       string // "resposta_final" ou "chamou_ferramenta"
	Ferramenta string
	Argumentos map[string]interface{}
	Observacao *ResultadoBusca
	DuracaoMs  int64
}

// ResultadoAgente e o retorno do loop ReAct: ou convergiu (Rascunho +
// Iteracoes preenchidos), ou nao convergiu no limite de voltas
// (EscalarParaAprovacaoHumana=true + Motivo preenchido) — nunca falha
// silenciosamente.
type ResultadoAgente struct {
	Rascunho                   string
	Iteracoes                  int
	EscalarParaAprovacaoHumana bool
	Motivo                     string
	Trilha                     []PassoTrilha
}

// AgenteICF roda o loop ReAct com o limite de iteracoes padrao (4).
func AgenteICF(ctx context.Context, client ChatClient, protocolo string) (ResultadoAgente, error) {
	return AgenteICFComLimite(ctx, client, protocolo, MaxIteracoesPadrao)
}

// AgenteICFComLimite roda o loop ReAct: Pensamento -> Acao -> Observacao ->
// Resposta Final. Criterio de parada explicito no orquestrador (maxIteracoes),
// nunca deixado para o modelo.
//
// O unico chamador que sobrescreve maxIteracoes e o cenario de demonstracao
// do escalonamento em SimularInteracao, igual ao original.
//
// Porte 1:1 de agenteICF em react-agent-prototype.js / react_agent_prototype.py.
func AgenteICFComLimite(ctx context.Context, client ChatClient, protocolo string, maxIteracoes int) (ResultadoAgente, error) {
	historico := []Mensagem{
		{
			"role": "system",
			"content": "Você redige seções de ICF para estudos clínicos. Sempre que precisar de uma cláusula " +
				"regulatória, chame a ferramenta em vez de perguntar ao usuário ou de escrever de memória. " +
				"Nunca peça esclarecimento ao usuário: decida os parâmetros da ferramenta a partir do " +
				"protocolo fornecido.",
		},
		{
			"role": "user",
			"content": "Protocolo do estudo: " + protocolo + "\n\n" +
				"Jurisdição regulatória deste estudo: ANVISA.\n\n" +
				"Gere a seção de assentimento do ICF (Termo de Consentimento) para este estudo, se aplicável. " +
				"Use a ferramenta disponível para buscar a cláusula regulatória correta antes de escrever o texto.",
		},
	}

	tools := []Mensagem{BuscarClausulaRegulatoria()}

	var trilha []PassoTrilha

	for volta := 1; volta <= maxIteracoes; volta++ {
		inicioVolta := time.Now()

		// Pensamento: o modelo decide, com base no historico acumulado, se ja sabe o suficiente
		resposta, err := client.Chat(ctx, historico, tools)
		duracaoMs := time.Since(inicioVolta).Milliseconds()
		if err != nil {
			return ResultadoAgente{}, err
		}

		if !resposta.TemChamadaFerramenta() {
			// Resposta Final: o modelo decidiu que ja tem o que precisa
			trilha = append(trilha, PassoTrilha{Volta: volta, Acao: "resposta_final", DuracaoMs: duracaoMs})
			return ResultadoAgente{Rascunho: resposta.Content, Iteracoes: volta, Trilha: trilha}, nil
		}

		// Acao: executa a ferramenta fora do modelo (codigo deterministico)
		chamada := resposta.Chamadas[0]
		observacao := ExecutarBuscaClausula(chamada.Argumentos)
		trilha = append(trilha, PassoTrilha{
			Volta: volta, Acao: "chamou_ferramenta", Ferramenta: chamada.Nome,
			Argumentos: chamada.Argumentos, Observacao: &observacao, DuracaoMs: duracaoMs,
		})

		// Observacao: volta como novo contexto para o proximo Pensamento
		historico = append(historico, resposta.MensagemBruta)
		conteudoJSON, err := json.Marshal(observacao)
		if err != nil {
			return ResultadoAgente{}, fmt.Errorf("react-agent: falha ao serializar observação da ferramenta: %w", err)
		}
		historico = append(historico, Mensagem{
			"role":      "tool",
			"tool_name": chamada.Nome,
			"content":   string(conteudoJSON),
		})
	}

	// Limite de iteracoes atingido sem convergencia: nunca falha silenciosamente
	return ResultadoAgente{
		EscalarParaAprovacaoHumana: true,
		Motivo:                     fmt.Sprintf("não convergiu em %d volta(s)", maxIteracoes),
		Trilha:                     trilha,
	}, nil
}
