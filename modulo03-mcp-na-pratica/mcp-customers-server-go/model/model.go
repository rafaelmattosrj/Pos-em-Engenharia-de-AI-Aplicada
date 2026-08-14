// Package model define os tipos de domínio do servidor MCP — equivalente a
// Customer.java e MutationResult.java.
package model

// Customer representa um cliente da legacy API.
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// MutationResult é o resultado de operações que modificam dados de
// clientes (create, update, delete).
type MutationResult struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Customer *Customer `json:"customer,omitempty"`
}
