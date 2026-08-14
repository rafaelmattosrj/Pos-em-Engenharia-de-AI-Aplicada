// Protótipo: Agente único do TrialForge (geração da seção de assentimento do
// ICF). Padrão: loop ReAct + ferramenta com schema tipado. Sem fila de
// mensagens.
//
// Modelo: Ollama local, rodando gemma4:e2b — gratuito, sem chave de API,
// roda inteiro na máquina. Alternativas pagas (Claude, Gemini, GPT) estão em
// internal/reactagent/provedores_pagos.go — mesmo loop ReAct, só a chamada
// ao modelo muda.
//
// Porte Go de react-agent-prototype.js / react_agent_prototype.py.
// Ver README.md deste projeto para detalhes de paridade e adaptações.
package main

import (
	"context"
	"fmt"
	"os"

	"react-agent-prototype/internal/reactagent"
)

const modelo = "gemma4:e2b"

func main() {
	loadDotEnv(".env")

	if err := rodarTestesFerramenta(); err != nil {
		fmt.Println("[Erro não tratado]", err)
		fmt.Println("[Sistema] Encaminhando ao Approval Gate — falha técnica também é motivo de escalonamento.")
		os.Exit(1)
	}

	baseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	client := reactagent.NewRetryingChatClient(reactagent.NewOllamaReactClient(baseURL, modelo))

	if err := reactagent.SimularInteracao(context.Background(), client); err != nil {
		fmt.Println("[Erro não tratado]", err)
		fmt.Println("[Sistema] Encaminhando ao Approval Gate — falha técnica também é motivo de escalonamento.")
		os.Exit(1)
	}
}

// rodarTestesFerramenta roda os testes automatizados da ferramenta
// (determinístico, sem chamar o modelo) — porte 1:1 de rodarTestesFerramenta
// em react-agent-prototype.js / .py.
func rodarTestesFerramenta() error {
	fmt.Println("== Testes: executarBuscaClausula (determinístico, 6 casos + 1 de parâmetro inválido) ==")
	passou := 0

	for _, caso := range reactagent.CasosTesteFerramenta {
		resultado := reactagent.ExecutarBuscaClausula(map[string]interface{}{
			"tema": caso.Tema, "jurisdicao": caso.Jurisdicao,
		})
		achou := resultado.Texto != ""
		ok := achou == caso.EsperaAchar
		fmt.Printf("  [%s] tema=\"%s\", jurisdicao=%s -> achou=%v (esperado=%v)\n",
			okFalhou(ok), caso.Tema, caso.Jurisdicao, achou, caso.EsperaAchar)
		if ok {
			passou++
		}
	}

	// Reproduz o bug real já corrigido: parâmetro grafado errado ("jurisdicicao") não
	// pode estourar erro pro chamador, tem que virar aviso de falha própria da ferramenta.
	casoInvalido := reactagent.ExecutarBuscaClausula(map[string]interface{}{"tema": "x", "jurisdicicao": "ANVISA"})
	invalidoOk := casoInvalido.Texto == "" && casoInvalido.Aviso != ""
	fmt.Printf("  [%s] parâmetro mal formado (\"jurisdicicao\") tratado como falha própria da ferramenta -> %v\n",
		okFalhou(invalidoOk), invalidoOk)
	if invalidoOk {
		passou++
	}

	total := len(reactagent.CasosTesteFerramenta) + 1
	fmt.Printf("Total: %d teste(s), %d passou(passaram), %d falhou(falharam).\n\n", total, passou, total-passou)
	if passou != total {
		return fmt.Errorf("testes da ferramenta falharam: %d/%d — corrija antes de rodar a simulação", passou, total)
	}
	return nil
}

func okFalhou(ok bool) string {
	if ok {
		return "OK"
	}
	return "FALHOU"
}
