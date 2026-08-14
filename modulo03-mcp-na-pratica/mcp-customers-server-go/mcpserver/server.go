package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-customers-server/client"
	"mcp-customers-server/service"
)

// New constrói o servidor MCP completo (tools + resource + prompts) —
// equivalente à auto-configuração do Spring AI MCP Server somada a
// McpServerApplication.java, ApiResource.java e CustomerPrompts.java.
func New(baseURL, serviceToken string) *mcp.Server {
	httpClient := client.NewCustomerHttpClient(baseURL, serviceToken)
	customerService := service.NewCustomerService(httpClient)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "customers-mcp-server",
		Version: "1.0.0",
	}, nil)

	registerTools(server, customerService)
	registerResource(server, baseURL)
	registerPrompts(server)

	return server
}
