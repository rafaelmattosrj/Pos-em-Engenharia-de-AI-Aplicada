// Servidor MCP que encapsula a legacy customer API — equivalente ao
// index.ts dos projetos MCP do curso / McpServerApplication.java. Expõe
// tools, um resource de documentação e prompts via protocolo MCP sobre
// stdio, para integração com Claude Desktop e outros clientes MCP.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-customers-server/mcpserver"
)

func main() {
	loadDotEnv(".env")

	baseURL := getEnv("LEGACY_API_URL", "http://localhost:3000")
	serviceToken := getEnv("API_SERVICE_TOKEN", "")

	server := mcpserver.New(baseURL, serviceToken)

	// Logs de diagnóstico vão para stderr — stdout é reservado para o
	// protocolo JSON-RPC do transporte stdio.
	log.SetOutput(os.Stderr)
	log.Printf("mcp-customers-server iniciado (legacy API: %s)\n", baseURL)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("servidor MCP encerrado com erro: %v", err)
	}
}
