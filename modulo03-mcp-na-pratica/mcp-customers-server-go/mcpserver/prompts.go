package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const searchCustomerPromptTemplate = `Você é um assistente especializado em gerenciamento de clientes.

O usuário quer buscar clientes com o seguinte termo: "%s"

Siga estas instruções:
1. Analise o termo de busca para identificar se é um ID, nome ou telefone
2. Use a tool ` + "`searchCustomer`" + ` com o parâmetro apropriado
3. Se o termo parece ser um ID (ex: UUID ou código alfanumérico), use o parâmetro ` + "`id`" + `
4. Se parece ser um nome, use o parâmetro ` + "`name`" + `
5. Se parece ser um telefone (contém números e/ou caracteres como +, -, (), espaços), use ` + "`phone`" + `
6. Apresente os resultados de forma clara, listando id, nome e telefone de cada cliente encontrado
7. Se nenhum cliente for encontrado, informe ao usuário e sugira refinamentos na busca
`

const createCustomerPromptTemplate = `Você é um assistente especializado em gerenciamento de clientes.

O usuário quer cadastrar um novo cliente com os seguintes dados:
- Nome: %[1]s
- Telefone: %[2]s

Siga estas instruções:
1. Verifique se o nome está preenchido e não é vazio
2. Verifique se o telefone está preenchido e não é vazio
3. Se algum dado estiver ausente, informe o usuário e solicite o dado faltante
4. Se os dados estiverem completos, use a tool ` + "`createCustomer`" + ` com name="%[1]s" e phone="%[2]s"
5. Após criar, confirme o sucesso ao usuário informando o ID gerado para o novo cliente
6. Em caso de erro, informe o usuário com a mensagem de erro retornada
`

// registerPrompts registra os 2 prompts MCP (busca e criação de clientes) —
// equivalente a CustomerPrompts.java.
func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "search-customer-prompt",
		Description: "Template para busca de clientes. Use quando o usuário quiser encontrar um ou mais clientes por nome, telefone ou ID.",
		Arguments: []*mcp.PromptArgument{
			{Name: "query", Description: "Termo de busca — pode ser nome, telefone ou ID do cliente", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]
		promptText := fmt.Sprintf(searchCustomerPromptTemplate, query)

		return &mcp.GetPromptResult{
			Description: "Template para busca de clientes por query",
			Messages: []*mcp.PromptMessage{
				{Role: "user", Content: &mcp.TextContent{Text: promptText}},
			},
		}, nil
	})

	server.AddPrompt(&mcp.Prompt{
		Name:        "create-customer-prompt",
		Description: "Template para criação de clientes. Use quando o usuário quiser cadastrar um novo cliente fornecendo nome e telefone.",
		Arguments: []*mcp.PromptArgument{
			{Name: "name", Description: "Nome completo do cliente a ser cadastrado", Required: true},
			{Name: "phone", Description: "Telefone de contato do cliente", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := req.Params.Arguments["name"]
		phone := req.Params.Arguments["phone"]
		promptText := fmt.Sprintf(createCustomerPromptTemplate, name, phone)

		return &mcp.GetPromptResult{
			Description: "Template para criação de novo cliente",
			Messages: []*mcp.PromptMessage{
				{Role: "user", Content: &mcp.TextContent{Text: promptText}},
			},
		}, nil
	})
}
