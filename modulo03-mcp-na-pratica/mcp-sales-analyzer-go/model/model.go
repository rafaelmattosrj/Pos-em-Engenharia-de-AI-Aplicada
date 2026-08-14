// Package model define os tipos de domínio do pipeline de análise —
// equivalente a AnalysisRequest.java, AnalysisResult.java e
// IntentResult.java.
package model

// AnalysisRequest é a requisição recebida pelo endpoint POST /analyze.
type AnalysisRequest struct {
	Question string `json:"question"`
	Data     string `json:"data"`
}

// AnalysisResult é o resultado final retornado pelo endpoint POST /analyze.
type AnalysisResult struct {
	Report          string   `json:"report"`
	ToolsUsed       []string `json:"toolsUsed"`
	ProcessingSteps []string `json:"processingSteps"`
}

// IntentResult é o resultado intermediário produzido pelo IntentNode.
type IntentResult struct {
	DataType       string   `json:"dataType"`
	ParsedData     string   `json:"parsedData"`
	Question       string   `json:"question"`
	SuggestedTools []string `json:"suggestedTools"`
}
