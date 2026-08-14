package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newConnectedClient sobe um servidor MCP completo (contra uma legacy API
// falsa) e um cliente MCP conectados via transporte em memória — permite
// testar o protocolo MCP ponta-a-ponta sem um processo real via stdio.
func newConnectedClient(t *testing.T, legacyAPI *httptest.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := New(legacyAPI.URL, "test-token")

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("erro conectando servidor: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("erro conectando cliente: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return session
}

func newFakeLegacyAPI(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/customers":
			json.NewEncoder(w).Encode([]map[string]string{
				{"id": "1", "name": "João Silva", "phone": "+55 11 91111-1111"},
				{"id": "2", "name": "Maria Souza", "phone": "+55 21 92222-2222"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/customers":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "novo-id", "name": "Novo Cliente", "phone": "123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestListCustomersTool(t *testing.T) {
	session := newConnectedClient(t, newFakeLegacyAPI(t))

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "listCustomers"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool retornou erro: %+v", result.Content)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "João Silva") {
		t.Errorf("esperava conteudo com João Silva, obteve: %s", text)
	}
}

func TestSearchCustomerTool(t *testing.T) {
	session := newConnectedClient(t, newFakeLegacyAPI(t))

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "searchCustomer",
		Arguments: map[string]any{"name": "maria"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool retornou erro: %+v", result.Content)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Maria Souza") || strings.Contains(text, "João") {
		t.Errorf("esperava apenas Maria Souza, obteve: %s", text)
	}
}

func TestCreateCustomerTool(t *testing.T) {
	session := newConnectedClient(t, newFakeLegacyAPI(t))

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "createCustomer",
		Arguments: map[string]any{"name": "Novo Cliente", "phone": "123"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool retornou erro: %+v", result.Content)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "novo-id") {
		t.Errorf("esperava id do cliente criado, obteve: %s", text)
	}
}

func TestApiResource(t *testing.T) {
	legacyAPI := newFakeLegacyAPI(t)
	session := newConnectedClient(t, legacyAPI)

	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "info://api"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("esperava 1 conteudo, obteve %d", len(result.Contents))
	}
	if !strings.Contains(result.Contents[0].Text, legacyAPI.URL) {
		t.Errorf("esperava documentacao contendo a URL base, obteve: %s", result.Contents[0].Text)
	}
}

func TestSearchCustomerPrompt(t *testing.T) {
	session := newConnectedClient(t, newFakeLegacyAPI(t))

	result, err := session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "search-customer-prompt",
		Arguments: map[string]string{"query": "joão"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("esperava 1 mensagem, obteve %d", len(result.Messages))
	}

	text := result.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(text, `"joão"`) || !strings.Contains(text, "searchCustomer") {
		t.Errorf("prompt inesperado: %s", text)
	}
}

func TestCreateCustomerPrompt(t *testing.T) {
	session := newConnectedClient(t, newFakeLegacyAPI(t))

	result, err := session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "create-customer-prompt",
		Arguments: map[string]string{"name": "Fulano", "phone": "555"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	text := result.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(text, "Fulano") || !strings.Contains(text, "555") {
		t.Errorf("prompt inesperado: %s", text)
	}
}
