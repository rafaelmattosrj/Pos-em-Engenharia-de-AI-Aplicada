package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const apiResourceTemplate = `# Legacy Customer API — Documentação

## URL Base
%s

## Autenticação
Todas as requisições requerem header:
  Authorization: Bearer <service_token>

O token é configurado via variável de ambiente API_SERVICE_TOKEN.

## Endpoints Disponíveis

### GET /customers
- Descrição: Lista todos os clientes cadastrados
- Resposta: Array de objetos Customer [ { id, name, phone } ]

### GET /customers/:id
- Descrição: Retorna um cliente específico pelo ID
- Parâmetros: id (path) — identificador único do cliente
- Resposta: Objeto Customer { id, name, phone }
- Erro 404: cliente não encontrado

### POST /customers
- Descrição: Cria um novo cliente
- Corpo: { "name": string, "phone": string }
- Resposta: Objeto Customer criado com ID gerado { id, name, phone }

### PUT /customers/:id
- Descrição: Atualiza dados de um cliente existente
- Parâmetros: id (path) — identificador único do cliente
- Corpo: { "name"?: string, "phone"?: string } (campos opcionais)
- Resposta: Objeto Customer atualizado { id, name, phone }

### DELETE /customers/:id
- Descrição: Remove permanentemente um cliente
- Parâmetros: id (path) — identificador único do cliente
- Resposta: 204 No Content em caso de sucesso

## Formato dos Dados

` + "```json" + `
{
  "id": "uuid-ou-string-gerado-pela-api",
  "name": "Nome Completo do Cliente",
  "phone": "+55 11 99999-9999"
}
` + "```" + `

## Observações
- A API legada NÃO possui endpoint de busca/filtro nativo
- Filtros por nome ou telefone são implementados no servidor MCP em memória
- IDs são gerados automaticamente pela API legada na criação
`

// registerResource registra o resource "info://api" com a documentação da
// legacy customer API — equivalente a ApiResource.java.
func registerResource(server *mcp.Server, baseURL string) {
	resourceContent := fmt.Sprintf(apiResourceTemplate, baseURL)

	server.AddResource(&mcp.Resource{
		URI:         "info://api",
		Name:        "Legacy Customer API Documentation",
		Description: "Documentação completa da legacy customer API: endpoints, autenticação e formato dos dados",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "info://api", MIMEType: "text/markdown", Text: resourceContent},
			},
		}, nil
	})
}
