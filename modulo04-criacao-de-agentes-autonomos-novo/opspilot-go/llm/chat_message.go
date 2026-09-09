// Package llm contém a abstração sobre "invocar um LLM" (porte de
// OpsChatModel / OpsResilientChatModel, agents/model.ts), desacoplada de
// qualquer provedor/biblioteca específica. Uma implementação real ligaria
// isto a um provedor OpenAI-compatible via net/http -- fora do escopo deste
// porte arquitetural (ver README).
package llm

// Role é o porte do union type de role em ChatMessage.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage é o porte de ChatMessage (agents/model.ts / llm/ChatMessage.java).
type ChatMessage struct {
	Role    Role
	Content string
}

func SystemMessage(content string) ChatMessage {
	return ChatMessage{Role: RoleSystem, Content: content}
}
func UserMessage(content string) ChatMessage { return ChatMessage{Role: RoleUser, Content: content} }
func ToolMessage(content string) ChatMessage { return ChatMessage{Role: RoleTool, Content: content} }
