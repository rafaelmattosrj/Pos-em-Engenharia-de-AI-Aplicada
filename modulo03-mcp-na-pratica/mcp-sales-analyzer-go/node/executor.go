package node

import (
	"context"
	"encoding/json"
	"fmt"

	"mcp-sales-analyzer/csvtool"
	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/openai"
)

const executorPromptTemplate = `Você é um analista de vendas especializado. Analise os dados abaixo e responda à pergunta.

Use as tools disponíveis quando necessário:
- csvToJson: para converter CSV em JSON antes de analisar
- Qualquer tool MCP disponível para análises adicionais

Tipo dos dados: %s

Pergunta: %s

Dados:
%s

Forneça uma análise detalhada e objetiva. Se os dados estiverem em CSV, converta-os primeiro.
`

var csvToJsonTool = openai.Tool{
	Type: "function",
	Function: openai.ToolFunction{
		Name:        "csvToJson",
		Description: "Converte dados em formato CSV para JSON. Use esta tool quando os dados de entrada estiverem em formato CSV com cabeçalho.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"csv": map[string]any{
					"type":        "string",
					"description": "Conteúdo CSV com header na primeira linha",
				},
			},
			"required": []string{"csv"},
		},
	},
}

// ExecutorNode executa a análise usando o LLM com function calling —
// equivalente a ExecutorNode.java. A tool local csvToJson é sempre
// registrada; tools MCP adicionais não são conectadas por padrão, assim como
// na versão Java (nenhum servidor MCP está configurado em
// application.properties — a lista de ToolCallbackProvider fica vazia).
type ExecutorNode struct {
	Client *openai.Client
}

// Execute roda o prompt de execução com a tool csvToJson disponível para o
// modelo, executando o loop de function calling até obter o relatório
// final.
func (n *ExecutorNode) Execute(ctx context.Context, intent model.IntentResult) (model.AnalysisResult, error) {
	var steps []string
	steps = append(steps, fmt.Sprintf("IntentNode concluído: tipo=%s, tools sugeridas=%v", intent.DataType, intent.SuggestedTools))

	prompt := fmt.Sprintf(executorPromptTemplate, intent.DataType, intent.Question, intent.ParsedData)

	steps = append(steps, "ExecutorNode iniciado: enviando prompt ao LLM com tools habilitadas")

	report, err := n.Client.ChatWithTools(ctx, prompt, []openai.Tool{csvToJsonTool}, executeToolCall)
	if err != nil {
		return model.AnalysisResult{}, err
	}

	steps = append(steps, "ExecutorNode concluído: relatório gerado com sucesso")

	toolsUsed := append([]string{}, intent.SuggestedTools...)
	if intent.DataType == "csv" && !contains(toolsUsed, "csvToJson") {
		toolsUsed = append(toolsUsed, "csvToJson")
	}

	return model.AnalysisResult{
		Report:          report,
		ToolsUsed:       toolsUsed,
		ProcessingSteps: steps,
	}, nil
}

func executeToolCall(name, argsJSON string) (string, error) {
	if name != "csvToJson" {
		return "", fmt.Errorf("tool desconhecida: %s", name)
	}

	var args struct {
		CSV string `json:"csv"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	return csvtool.Convert(args.CSV)
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
