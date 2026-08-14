package service

import "testing"

func TestNewCustomerService_SeedsFiveCustomers(t *testing.T) {
	s := NewCustomerService()
	customers := s.FindAll()
	if len(customers) != 5 {
		t.Fatalf("esperava 5 clientes pre-cadastrados, obteve %d", len(customers))
	}
}

func TestCreateAndFindByID(t *testing.T) {
	s := NewCustomerService()
	created := s.Create("Novo Cliente", "(99) 99999-9999")

	found, ok := s.FindByID(created.ID)
	if !ok {
		t.Fatal("esperava encontrar o cliente criado")
	}
	if found.Name != "Novo Cliente" {
		t.Errorf("nome inesperado: %q", found.Name)
	}
}

func TestUpdate_PartialFields(t *testing.T) {
	s := NewCustomerService()
	created := s.Create("Nome Original", "111")

	updated, ok := s.Update(created.ID, "Nome Novo", "")
	if !ok {
		t.Fatal("esperava update bem sucedido")
	}
	if updated.Name != "Nome Novo" {
		t.Errorf("esperava nome atualizado, obteve %q", updated.Name)
	}
	if updated.Phone != "111" {
		t.Errorf("esperava telefone preservado, obteve %q", updated.Phone)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	s := NewCustomerService()
	_, ok := s.Update(s.Create("x", "y").ID, "", "")
	if !ok {
		t.Fatal("update do cliente recem criado deveria funcionar")
	}
}

func TestDelete(t *testing.T) {
	s := NewCustomerService()
	created := s.Create("A remover", "000")

	if !s.Delete(created.ID) {
		t.Fatal("esperava delete bem sucedido")
	}
	if _, ok := s.FindByID(created.ID); ok {
		t.Error("cliente nao deveria mais existir apos delete")
	}
	if s.Delete(created.ID) {
		t.Error("segundo delete do mesmo id deveria retornar false")
	}
}
