package reactagent

import "errors"

// Três alternativas PAGAS ao Ollama local usado em OllamaReactClient: Claude
// (Anthropic), Gemini (Google) e GPT (OpenAI). A lição do diagrama de
// referência do Módulo 1.2 vale aqui: o modelo é a peça que se troca — o
// loop ReAct (AgenteICF), a ferramenta (ExecutarBuscaClausula) e o critério
// de parada não mudam em nenhuma das três.
//
// Repare também que cada provedor declara o schema da MESMA ferramenta num
// formato ligeiramente diferente: Claude e Gemini usam a mesma forma
// achatada (name/description soltos na raiz do schema), o GPT usa uma forma
// aninhada, dentro de "function": {...} (igual ao formato nativo do Ollama,
// usado em BuscarClausulaRegulatoria). É exatamente esse tipo de
// fragmentação que um protocolo padronizado como o MCP existe para resolver.
//
// IMPORTANTE — arquivo de referência, NÃO executado pelo fluxo principal
// (main.go). Nenhum dos três SDKs pagos (Anthropic, Google GenAI, OpenAI)
// está no go.mod deste projeto (só a biblioteca padrão é usada). Adicionar
// um deles exigiria rede + credenciais pagas no momento do build, o que
// quebraria `go build` offline para quem só quer rodar a demo local com
// Ollama. Por isso ChamarClaude/ChamarGemini/ChamarGPT abaixo retornam
// erro "não implementado" se chamadas — o comentário acima de cada uma
// documenta a chamada real que o SDK faria. As três funções SchemaX SÃO
// executáveis: constroem o schema real de cada provedor só com a biblioteca
// padrão (sem SDK nenhum), pra comparar as três formas lado a lado.
//
// Para usar de verdade: rode `go get` da SDK do provedor escolhido, configure
// a variável de ambiente da chave (ver .env.example e o README.md deste
// projeto) e substitua o corpo da função pela chamada real documentada no
// comentário.
//
// Porte de referência de provedores-pagos.js / provedores_pagos.py.

// ChamadaResultado unifica o retorno das três chamadas: ou uma chamada de
// ferramenta, ou uma resposta final.
type ChamadaResultado struct {
	Tipo  string // "chamada_ferramenta" ou "resposta_final"
	Nome  string
	Args  map[string]interface{}
	Texto string
}

// ---------- 1. Claude (Anthropic) — go get github.com/anthropics/anthropic-sdk-go ----------

// SchemaClaude monta o schema da ferramenta no formato do Claude:
// name/description soltos na raiz, "input_schema".
func SchemaClaude() map[string]interface{} {
	return map[string]interface{}{
		"name":        NomeFerramenta,
		"description": descricaoFerramenta,
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tema":       map[string]interface{}{"type": "string"},
				"jurisdicao": map[string]interface{}{"type": "string", "enum": []string{"ANVISA", "FDA"}},
			},
			"required": []string{"tema", "jurisdicao"},
		},
	}
}

// ChamarClaude é referência (não executada). Com o SDK instalado, o corpo
// seria algo como:
//
//	client := anthropic.NewClient() // lê ANTHROPIC_API_KEY do ambiente
//	resposta, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
//	    Model:     "claude-sonnet-5",
//	    MaxTokens: 1024,
//	    Tools:     []anthropic.ToolParam{ /* SchemaClaude() convertido */ },
//	    Messages:  historico,
//	})
//	for _, bloco := range resposta.Content {
//	    if bloco.Type == "tool_use" {
//	        return ChamadaResultado{Tipo: "chamada_ferramenta", Nome: bloco.Name, Args: bloco.Input}, nil
//	    }
//	}
//	return ChamadaResultado{Tipo: "resposta_final", Texto: resposta.Content[0].Text}, nil
func ChamarClaude(historico []Mensagem) (ChamadaResultado, error) {
	return ChamadaResultado{}, errors.New(
		"referência não executável: rode 'go get github.com/anthropics/anthropic-sdk-go' e configure " +
			"ANTHROPIC_API_KEY (ver .env.example e README.md)")
}

// ---------- 2. Gemini (Google) — go get google.golang.org/genai ----------

// SchemaGemini monta o schema da ferramenta no formato do Gemini: achatado
// como o Claude, mas com "type": "function".
func SchemaGemini() map[string]interface{} {
	return map[string]interface{}{
		"type":        "function",
		"name":        NomeFerramenta,
		"description": descricaoFerramenta,
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tema":       map[string]interface{}{"type": "string"},
				"jurisdicao": map[string]interface{}{"type": "string", "enum": []string{"ANVISA", "FDA"}},
			},
			"required": []string{"tema", "jurisdicao"},
		},
	}
}

// ChamarGemini é referência (não executada).
//
// Nota de assinatura: ChamarClaude/ChamarGPT recebem "historico" (o mesmo
// histórico de mensagens usado no loop ReAct); esta função recebe só
// "protocolo" (string) porque a API de Interactions do Gemini gerencia o
// histórico de conversa do lado do servidor, não como uma lista que o
// chamador monta e reenvia a cada turno — é uma diferença real de formato
// entre provedores, não um descuido. Ao adaptar este sketch pro seu próprio
// protótipo, ajuste o chamador de acordo (não assuma as três funções
// intercambiáveis por assinatura, só por papel na arquitetura).
//
// Com o SDK instalado, o corpo seria algo como:
//
//	client, _ := genai.NewClient(ctx, nil) // lê GEMINI_API_KEY do ambiente
//	interacao, _ := client.Interactions.Create(ctx, &genai.InteractionCreateParams{
//	    Model: "gemini-3.5-flash",
//	    Input: protocolo,
//	    Tools: []genai.Tool{ /* SchemaGemini() convertido */ },
//	})
//	for _, passo := range interacao.Steps {
//	    if passo.Type == "function_call" {
//	        return ChamadaResultado{Tipo: "chamada_ferramenta", Nome: passo.Name, Args: passo.Arguments}, nil
//	    }
//	}
//	return ChamadaResultado{Tipo: "resposta_final", Texto: interacao.OutputText}, nil
func ChamarGemini(protocolo string) (ChamadaResultado, error) {
	return ChamadaResultado{}, errors.New(
		"referência não executável: rode 'go get google.golang.org/genai' e configure GEMINI_API_KEY " +
			"(ver .env.example e README.md)")
}

// ---------- 3. GPT (OpenAI) — go get github.com/openai/openai-go ----------

// SchemaGPT monta o schema da ferramenta no formato do GPT: forma aninhada
// — mesmo formato já usado nativamente pelo Ollama (ver BuscarClausulaRegulatoria).
func SchemaGPT() Mensagem {
	return BuscarClausulaRegulatoria()
}

// ChamarGPT é referência (não executada). Com o SDK instalado, o corpo
// seria algo como:
//
//	client := openai.NewClient() // lê OPENAI_API_KEY do ambiente
//	resposta, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
//	    Model:    "gpt-5.6",
//	    Messages: historico,
//	    Tools:    []openai.ChatCompletionToolParam{ /* SchemaGPT() convertido */ },
//	})
//	toolCalls := resposta.Choices[0].Message.ToolCalls
//	if len(toolCalls) > 0 {
//	    args := parseArguments(toolCalls[0].Function.Arguments) // JSON.parse do original
//	    return ChamadaResultado{Tipo: "chamada_ferramenta", Nome: toolCalls[0].Function.Name, Args: args}, nil
//	}
//	return ChamadaResultado{Tipo: "resposta_final", Texto: resposta.Choices[0].Message.Content}, nil
func ChamarGPT(historico []Mensagem) (ChamadaResultado, error) {
	return ChamadaResultado{}, errors.New(
		"referência não executável: rode 'go get github.com/openai/openai-go' e configure OPENAI_API_KEY " +
			"(ver .env.example e README.md)")
}
