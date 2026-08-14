package reactagent

import (
	"context"
	"strings"
	"testing"
)

// Cobre os mesmos três comportamentos demonstrados em SimularInteracao() de
// react-agent-prototype.js / .py: convergência normal (com chamada de
// ferramenta), resposta final direta, e não-convergência com escalonamento
// — usando um fakeChatClient no lugar do Ollama real.

func respostaComToolCall(nomeFerramenta string, argumentos map[string]interface{}) ModeloResposta {
	mensagemBruta := Mensagem{
		"role":    "assistant",
		"content": nil,
		"tool_calls": []interface{}{
			map[string]interface{}{
				"id": "call-1",
				"function": map[string]interface{}{
					"name":      nomeFerramenta,
					"arguments": argumentos,
				},
			},
		},
	}
	return ModeloResposta{
		MensagemBruta: mensagemBruta,
		Chamadas:      []ChamadaFerramenta{{ID: "call-1", Nome: nomeFerramenta, Argumentos: argumentos}},
	}
}

func TestAgenteICF_ChamaFerramentaEDepoisRespondeFinal_ConvergeComTrilhaDeDuasVoltas(t *testing.T) {
	fake := (&fakeChatClient{}).
		comResposta(respostaComToolCall(NomeFerramenta, map[string]interface{}{
			"tema": "Assentimento para menores de idade em estudos clínicos", "jurisdicao": "ANVISA",
		})).
		comResposta(respostaFinalFake("Rascunho da seção de assentimento, com a cláusula ANVISA citada."))

	resultado, err := AgenteICF(context.Background(), fake, "Estudo fase II, público-alvo entre 12 e 17 anos.")
	if err != nil {
		t.Fatalf("não esperava erro, obteve %v", err)
	}

	if resultado.EscalarParaAprovacaoHumana {
		t.Error("não esperava escalonamento")
	}
	if resultado.Iteracoes != 2 {
		t.Errorf("esperava 2 iterações, obteve %d", resultado.Iteracoes)
	}
	if !strings.Contains(resultado.Rascunho, "assentimento") {
		t.Errorf("rascunho inesperado: %q", resultado.Rascunho)
	}
	if len(resultado.Trilha) != 2 {
		t.Fatalf("esperava trilha com 2 passos, obteve %d", len(resultado.Trilha))
	}
	if resultado.Trilha[0].Acao != "chamou_ferramenta" {
		t.Errorf("esperava primeiro passo=chamou_ferramenta, obteve %q", resultado.Trilha[0].Acao)
	}
	if resultado.Trilha[1].Acao != "resposta_final" {
		t.Errorf("esperava segundo passo=resposta_final, obteve %q", resultado.Trilha[1].Acao)
	}
	if fake.Chamadas != 2 {
		t.Errorf("esperava 2 chamadas ao client, obteve %d", fake.Chamadas)
	}
}

func TestAgenteICF_RespondeDireto_ConvergeComUmaVoltaSoNaTrilha(t *testing.T) {
	fake := (&fakeChatClient{}).comResposta(respostaFinalFake("Não se aplica assentimento."))

	resultado, err := AgenteICF(context.Background(), fake,
		"Estudo fase III, população adulta, sem envolvimento de menores de idade.")
	if err != nil {
		t.Fatalf("não esperava erro, obteve %v", err)
	}

	if resultado.EscalarParaAprovacaoHumana {
		t.Error("não esperava escalonamento")
	}
	if resultado.Iteracoes != 1 {
		t.Errorf("esperava 1 iteração, obteve %d", resultado.Iteracoes)
	}
	if len(resultado.Trilha) != 1 {
		t.Errorf("esperava trilha com 1 passo, obteve %d", len(resultado.Trilha))
	}
}

func TestAgenteICF_NaoConvergeNoLimiteDeIteracoes_EscalaParaAprovacaoHumana(t *testing.T) {
	fake := (&fakeChatClient{}).comResposta(respostaComToolCall(NomeFerramenta, map[string]interface{}{
		"tema": "Assentimento para menores de idade", "jurisdicao": "ANVISA",
	}))

	resultado, err := AgenteICFComLimite(context.Background(), fake, "Estudo fase III, população adulta.", 1)
	if err != nil {
		t.Fatalf("não esperava erro, obteve %v", err)
	}

	if !resultado.EscalarParaAprovacaoHumana {
		t.Error("esperava escalonamento")
	}
	if resultado.Rascunho != "" {
		t.Errorf("não esperava rascunho, obteve %q", resultado.Rascunho)
	}
	if !strings.Contains(resultado.Motivo, "não convergiu em 1 volta(s)") {
		t.Errorf("motivo inesperado: %q", resultado.Motivo)
	}
	if len(resultado.Trilha) != 1 {
		t.Errorf("esperava trilha com 1 passo, obteve %d", len(resultado.Trilha))
	}
	if fake.Chamadas != 1 {
		t.Errorf("esperava 1 chamada ao client, obteve %d", fake.Chamadas)
	}
}

func TestAgenteICF_FerramentaNaoEncontraClausula_ObservacaoRefleteAviso(t *testing.T) {
	fake := (&fakeChatClient{}).
		comResposta(respostaComToolCall(NomeFerramenta, map[string]interface{}{
			"tema": "Consentimento informado de população adulta", "jurisdicao": "ANVISA",
		})).
		comResposta(respostaFinalFake("Cláusula de assentimento não se aplica a este estudo."))

	resultado, err := AgenteICF(context.Background(), fake, "Estudo fase III, população adulta, diabetes tipo 2.")
	if err != nil {
		t.Fatalf("não esperava erro, obteve %v", err)
	}

	observacao := resultado.Trilha[0].Observacao
	if observacao == nil {
		t.Fatal("esperava observação preenchida no primeiro passo")
	}
	if observacao.Texto != "" {
		t.Errorf("não esperava cláusula encontrada, obteve %q", observacao.Texto)
	}
	if !strings.Contains(observacao.Aviso, "não encontrado") {
		t.Errorf("aviso inesperado: %q", observacao.Aviso)
	}
}
