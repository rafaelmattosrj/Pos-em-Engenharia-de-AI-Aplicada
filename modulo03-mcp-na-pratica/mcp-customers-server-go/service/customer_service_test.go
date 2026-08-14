package service

import (
	"context"
	"testing"

	"mcp-customers-server/model"
)

// fakeHTTPClient é o duplo de teste equivalente ao @Mock CustomerHttpClient
// de CustomerServiceTest.java.
type fakeHTTPClient struct {
	customers []model.Customer
}

func (f *fakeHTTPClient) ListAll(ctx context.Context) ([]model.Customer, error) {
	return f.customers, nil
}
func (f *fakeHTTPClient) FindByID(ctx context.Context, id string) *model.Customer { return nil }
func (f *fakeHTTPClient) Create(ctx context.Context, name, phone string) model.MutationResult {
	return model.MutationResult{}
}
func (f *fakeHTTPClient) Update(ctx context.Context, id, name, phone string) model.MutationResult {
	return model.MutationResult{}
}
func (f *fakeHTTPClient) Delete(ctx context.Context, id string) model.MutationResult {
	return model.MutationResult{}
}

func customersFixture() []model.Customer {
	return []model.Customer{
		{ID: "1", Name: "João Silva", Phone: "+55 11 91111-1111"},
		{ID: "2", Name: "Maria Souza", Phone: "+55 21 92222-2222"},
		{ID: "3", Name: "João Oliveira", Phone: "+55 11 93333-3333"},
	}
}

func TestSearchCustomer_FiltraPorNomeParcialCaseInsensitive(t *testing.T) {
	svc := NewCustomerService(&fakeHTTPClient{customers: customersFixture()})

	result, err := svc.SearchCustomer(context.Background(), "", "joão", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("esperava 2 resultados, obteve %d", len(result))
	}
	names := map[string]bool{result[0].Name: true, result[1].Name: true}
	if !names["João Silva"] || !names["João Oliveira"] {
		t.Errorf("nomes inesperados: %v", names)
	}
}

func TestSearchCustomer_FiltraPorTelefoneParcial(t *testing.T) {
	svc := NewCustomerService(&fakeHTTPClient{customers: customersFixture()})

	result, err := svc.SearchCustomer(context.Background(), "", "", "11")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("esperava 2 resultados, obteve %d", len(result))
	}
	ids := map[string]bool{result[0].ID: true, result[1].ID: true}
	if !ids["1"] || !ids["3"] {
		t.Errorf("ids inesperados: %v", ids)
	}
}

func TestSearchCustomer_SemFiltrosRetornaTodos(t *testing.T) {
	svc := NewCustomerService(&fakeHTTPClient{customers: customersFixture()})

	result, err := svc.SearchCustomer(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("esperava 3 resultados, obteve %d", len(result))
	}
}

func TestSearchCustomer_FiltraPorIDExato(t *testing.T) {
	svc := NewCustomerService(&fakeHTTPClient{customers: customersFixture()})

	result, err := svc.SearchCustomer(context.Background(), "2", "", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Maria Souza" {
		t.Fatalf("resultado inesperado: %+v", result)
	}
}
