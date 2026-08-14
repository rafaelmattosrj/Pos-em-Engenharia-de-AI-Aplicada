// Package model define as entidades de domínio da API legada — equivalente
// a Customer.java, AuthRequest.java e AuthResponse.java.
package model

import "github.com/google/uuid"

// Customer é a entidade principal do domínio.
type Customer struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Phone string    `json:"phone"`
}

// NewCustomer cria um Customer com um novo UUID gerado automaticamente —
// equivalente a Customer.create.
func NewCustomer(name, phone string) Customer {
	return Customer{ID: uuid.New(), Name: name, Phone: phone}
}

// WithUpdated retorna uma cópia do Customer com os dados atualizados,
// mantendo o mesmo ID — campos vazios preservam o valor atual, igual à
// semântica de "name != null ? name : this.name" em Java.
func (c Customer) WithUpdated(name, phone string) Customer {
	if name != "" {
		c.Name = name
	}
	if phone != "" {
		c.Phone = phone
	}
	return c
}

// AuthRequest é o payload de requisição de login.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse é o payload de resposta da autenticação.
type AuthResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}
