// Package node implementa os nós do pipeline de análise — equivalente a
// IntentNode.java e ExecutorNode.java.
package node

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/openai"
)

const intentPromptTemplate = `Analise os dados e a pergunta abaixo. Responda APENAS com JSON válido, sem markdown.

Formato esperado:
{
  "dataType": "csv" ou "json",
  "parsedData": "<dados normalizados como string>",
  "question": "<pergunta original>",
  "suggestedTools": ["<tool1>", "<tool2>"]
}

Tools disponíveis: csvToJson, analyzeData, summarize

Pergunta: %s

Dados:
%s
`

// IntentNode extrai a intenção estruturada do pedido do usuário antes de
// executar — equivalente ao IntentNode.java (que por sua vez espelha o
// IntentNode do LangGraph).
type IntentNode struct {
	Client *openai.Client
}

// Extract analisa request e retorna um IntentResult. Qualquer falha na
// chamada ao LLM ou no parse da resposta cai num fallback seguro com o tipo
// de dado detectado heuristicamente — mesmo comportamento da versão Java.
func (n *IntentNode) Extract(ctx context.Context, request model.AnalysisRequest) model.IntentResult {
	dataType := detectDataType(request.Data)

	prompt := fmt.Sprintf(intentPromptTemplate, request.Question, request.Data)

	response, err := n.Client.Chat(ctx, prompt)
	if err != nil {
		return fallbackIntent(request, dataType)
	}

	return parseIntentResponse(response, request, dataType)
}

func detectDataType(data string) string {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return "unknown"
	}
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		return "json"
	}
	return "csv"
}

func parseIntentResponse(response string, request model.AnalysisRequest, fallbackDataType string) model.IntentResult {
	cleanJSON := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(response, "```json", ""), "```", ""))

	var parsed struct {
		DataType       string   `json:"dataType"`
		ParsedData     string   `json:"parsedData"`
		Question       string   `json:"question"`
		SuggestedTools []string `json:"suggestedTools"`
	}
	if err := json.Unmarshal([]byte(cleanJSON), &parsed); err != nil {
		return fallbackIntent(request, fallbackDataType)
	}

	dataType := parsed.DataType
	if dataType == "" {
		dataType = fallbackDataType
	}
	parsedData := parsed.ParsedData
	if parsedData == "" {
		parsedData = request.Data
	}
	question := parsed.Question
	if question == "" {
		question = request.Question
	}

	return model.IntentResult{
		DataType:       dataType,
		ParsedData:     parsedData,
		Question:       question,
		SuggestedTools: parsed.SuggestedTools,
	}
}

func fallbackIntent(request model.AnalysisRequest, dataType string) model.IntentResult {
	return model.IntentResult{
		DataType:       dataType,
		ParsedData:     request.Data,
		Question:       request.Question,
		SuggestedTools: nil,
	}
}
