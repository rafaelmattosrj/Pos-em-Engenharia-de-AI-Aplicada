// Package mcpserver registra as tools, resources e prompts MCP expostos pelo
// servidor — equivalente a CustomerTools.java, ApiResource.java e
// CustomerPrompts.java.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-customers-server/model"
	"mcp-customers-server/service"
)

// registerTools registra as 5 tools de gerenciamento de clientes —
// equivalente a CustomerTools.java.
func registerTools(server *mcp.Server, customerService *service.CustomerService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listCustomers",
		Description: "Lista todos os clientes cadastrados no sistema. Retorna id, nome e telefone de cada cliente.",
	}, listCustomersHandler(customerService))

	mcp.AddTool(server, &mcp.Tool{
		Name: "searchCustomer",
		Description: "Busca clientes por critérios flexíveis. Todos os parâmetros são opcionais.\n" +
			"Use 'id' para busca exata por identificador.\n" +
			"Use 'name' para busca parcial case-insensitive por nome.\n" +
			"Use 'phone' para busca parcial por telefone.\n" +
			"Se nenhum parâmetro for fornecido, retorna todos os clientes.",
	}, searchCustomerHandler(customerService))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "createCustomer",
		Description: "Cria um novo cliente no sistema com nome e telefone. Retorna os dados do cliente criado incluindo o ID gerado.",
	}, createCustomerHandler(customerService))

	mcp.AddTool(server, &mcp.Tool{
		Name: "updateCustomer",
		Description: "Atualiza os dados de um cliente existente identificado pelo ID.\n" +
			"Forneça apenas os campos que deseja alterar (name e/ou phone).\n" +
			"Pelo menos um dos campos name ou phone deve ser informado.",
	}, updateCustomerHandler(customerService))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteCustomer",
		Description: "Remove permanentemente um cliente do sistema pelo ID. Esta operação não pode ser desfeita.",
	}, deleteCustomerHandler(customerService))
}

type emptyArgs struct{}

type searchCustomerArgs struct {
	ID    string `json:"id,omitempty" jsonschema:"ID exato do cliente (opcional)"`
	Name  string `json:"name,omitempty" jsonschema:"Nome ou parte do nome do cliente (opcional)"`
	Phone string `json:"phone,omitempty" jsonschema:"Telefone ou parte do telefone do cliente (opcional)"`
}

type createCustomerArgs struct {
	Name  string `json:"name" jsonschema:"Nome completo do cliente"`
	Phone string `json:"phone" jsonschema:"Telefone de contato do cliente"`
}

type updateCustomerArgs struct {
	ID    string `json:"id" jsonschema:"ID do cliente a ser atualizado"`
	Name  string `json:"name,omitempty" jsonschema:"Novo nome do cliente (opcional)"`
	Phone string `json:"phone,omitempty" jsonschema:"Novo telefone do cliente (opcional)"`
}

type deleteCustomerArgs struct {
	ID string `json:"id" jsonschema:"ID do cliente a ser removido"`
}

// customerListOutput envelopa a lista de clientes num objeto — o SDK Go do
// MCP exige que o schema de saída estruturada seja do tipo "object" no nível
// raiz (arrays soltos não são aceitos), diferente do retorno direto de
// List<Customer> em Java. O conteúdo textual visto pelo LLM continua sendo o
// JSON com os mesmos dados, apenas envelopado neste campo.
type customerListOutput struct {
	Customers []model.Customer `json:"customers"`
}

func listCustomersHandler(s *service.CustomerService) mcp.ToolHandlerFor[emptyArgs, customerListOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, customerListOutput, error) {
		customers, err := s.ListAll(ctx)
		if err != nil {
			return nil, customerListOutput{}, err
		}
		return nil, customerListOutput{Customers: customers}, nil
	}
}

func searchCustomerHandler(s *service.CustomerService) mcp.ToolHandlerFor[searchCustomerArgs, customerListOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, args searchCustomerArgs) (*mcp.CallToolResult, customerListOutput, error) {
		customers, err := s.SearchCustomer(ctx, args.ID, args.Name, args.Phone)
		if err != nil {
			return nil, customerListOutput{}, err
		}
		return nil, customerListOutput{Customers: customers}, nil
	}
}

func createCustomerHandler(s *service.CustomerService) mcp.ToolHandlerFor[createCustomerArgs, model.MutationResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, args createCustomerArgs) (*mcp.CallToolResult, model.MutationResult, error) {
		return nil, s.CreateCustomer(ctx, args.Name, args.Phone), nil
	}
}

func updateCustomerHandler(s *service.CustomerService) mcp.ToolHandlerFor[updateCustomerArgs, model.MutationResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, args updateCustomerArgs) (*mcp.CallToolResult, model.MutationResult, error) {
		return nil, s.UpdateCustomer(ctx, args.ID, args.Name, args.Phone), nil
	}
}

func deleteCustomerHandler(s *service.CustomerService) mcp.ToolHandlerFor[deleteCustomerArgs, model.MutationResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, args deleteCustomerArgs) (*mcp.CallToolResult, model.MutationResult, error) {
		return nil, s.DeleteCustomer(ctx, args.ID), nil
	}
}
