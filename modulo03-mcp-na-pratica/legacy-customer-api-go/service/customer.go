package service

import (
	"sync"

	"github.com/google/uuid"

	"legacy-customer-api/model"
)

// CustomerService gerencia clientes em memória — equivalente a
// CustomerService.java. Protegido por mutex: um servidor Go atende
// requisições HTTP concorrentemente por padrão.
type CustomerService struct {
	mu    sync.RWMutex
	order []uuid.UUID
	store map[uuid.UUID]model.Customer
}

// NewCustomerService cria o serviço já populado com os mesmos 5 clientes de
// demonstração da versão Java.
func NewCustomerService() *CustomerService {
	s := &CustomerService{store: make(map[uuid.UUID]model.Customer)}
	s.seed("Alice Silva", "(11) 91234-5678")
	s.seed("Bruno Oliveira", "(21) 98765-4321")
	s.seed("Carla Souza", "(31) 97654-3210")
	s.seed("Diego Lima", "(41) 96543-2109")
	s.seed("Eva Costa", "(51) 95432-1098")
	return s
}

func (s *CustomerService) seed(name, phone string) {
	c := model.NewCustomer(name, phone)
	s.store[c.ID] = c
	s.order = append(s.order, c.ID)
}

// FindAll retorna todos os clientes na ordem de criação.
func (s *CustomerService) FindAll() []model.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customers := make([]model.Customer, 0, len(s.order))
	for _, id := range s.order {
		customers = append(customers, s.store[id])
	}
	return customers
}

// FindByID retorna um cliente por ID, ou false se não encontrado.
func (s *CustomerService) FindByID(id uuid.UUID) (model.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.store[id]
	return c, ok
}

// Create cria um novo cliente e o armazena em memória.
func (s *CustomerService) Create(name, phone string) model.Customer {
	c := model.NewCustomer(name, phone)
	s.mu.Lock()
	s.store[c.ID] = c
	s.order = append(s.order, c.ID)
	s.mu.Unlock()
	return c
}

// Update atualiza um cliente existente. Retorna false se o ID não existir.
func (s *CustomerService) Update(id uuid.UUID, name, phone string) (model.Customer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.store[id]
	if !ok {
		return model.Customer{}, false
	}
	updated := existing.WithUpdated(name, phone)
	s.store[id] = updated
	return updated, true
}

// Delete remove um cliente por ID. Retorna true se removido.
func (s *CustomerService) Delete(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.store[id]; !ok {
		return false
	}
	delete(s.store, id)
	for i, orderedID := range s.order {
		if orderedID == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}
