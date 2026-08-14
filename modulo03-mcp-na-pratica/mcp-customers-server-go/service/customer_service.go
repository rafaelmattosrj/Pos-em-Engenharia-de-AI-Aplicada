// Package service contém a camada de negócio entre o servidor MCP e o
// cliente HTTP — equivalente a CustomerService.java. Adiciona busca/filtro
// em memória que a API legada não oferece nativamente.
package service

import (
	"context"
	"strings"

	"mcp-customers-server/model"
)

// customerLister é a única operação do CustomerHttpClient de que este
// serviço depende — extraída como interface para permitir um duplo de teste
// simples (equivalente ao @Mock CustomerHttpClient em
// CustomerServiceTest.java, sem precisar de um framework de mocking).
type customerLister interface {
	ListAll(ctx context.Context) ([]model.Customer, error)
	FindByID(ctx context.Context, id string) *model.Customer
	Create(ctx context.Context, name, phone string) model.MutationResult
	Update(ctx context.Context, id, name, phone string) model.MutationResult
	Delete(ctx context.Context, id string) model.MutationResult
}

// CustomerService é o serviço de domínio para operações com clientes.
type CustomerService struct {
	httpClient customerLister
}

// NewCustomerService cria um CustomerService sobre o cliente HTTP informado.
func NewCustomerService(httpClient customerLister) *CustomerService {
	return &CustomerService{httpClient: httpClient}
}

// ListAll retorna todos os clientes cadastrados.
func (s *CustomerService) ListAll(ctx context.Context) ([]model.Customer, error) {
	return s.httpClient.ListAll(ctx)
}

// SearchCustomer busca e filtra clientes em memória por id, nome ou
// telefone. Todos os parâmetros são opcionais (string vazia = sem filtro); o
// filtro usa correspondência parcial case-insensitive para name e phone —
// equivalente a CustomerService.searchCustomer.
func (s *CustomerService) SearchCustomer(ctx context.Context, id, name, phone string) ([]model.Customer, error) {
	all, err := s.httpClient.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []model.Customer
	for _, c := range all {
		if id != "" && id != c.ID {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(c.Name), strings.ToLower(name)) {
			continue
		}
		if phone != "" && !strings.Contains(c.Phone, phone) {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

// CreateCustomer cria um novo cliente.
func (s *CustomerService) CreateCustomer(ctx context.Context, name, phone string) model.MutationResult {
	return s.httpClient.Create(ctx, name, phone)
}

// UpdateCustomer atualiza um cliente existente.
func (s *CustomerService) UpdateCustomer(ctx context.Context, id, name, phone string) model.MutationResult {
	return s.httpClient.Update(ctx, id, name, phone)
}

// DeleteCustomer remove um cliente pelo ID.
func (s *CustomerService) DeleteCustomer(ctx context.Context, id string) model.MutationResult {
	return s.httpClient.Delete(ctx, id)
}
