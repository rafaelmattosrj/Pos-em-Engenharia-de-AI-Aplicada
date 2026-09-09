package llm

// ToolCall é o porte de ToolCall (agents/model.ts / llm/ToolCall.java).
type ToolCall struct {
	ToolName string
	Args     map[string]any
}

// ModelResponse é o porte de ModelResponse (agents/model.ts / llm/ModelResponse.java).
type ModelResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// TextResponse é o porte de ModelResponse.text(content) (porte Java).
func TextResponse(content string) ModelResponse {
	return ModelResponse{Content: content}
}
