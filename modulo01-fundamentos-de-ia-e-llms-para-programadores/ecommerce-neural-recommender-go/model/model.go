// Package model define as entidades do domínio — espelho de Product/User da
// versão Java (products.json/users.json do exemplo original em JS).
package model

import "fmt"

// Product representa um produto do catálogo do e-commerce.
type Product struct {
	ID       int
	Name     string
	Category string
	Price    float64
	Color    string
}

func (p Product) String() string {
	return fmt.Sprintf("Product{id=%d, name=%q, category=%q, price=%.2f, color=%q}",
		p.ID, p.Name, p.Category, p.Price, p.Color)
}

// User representa um usuário com seu histórico de compras.
type User struct {
	ID        int
	Name      string
	Age       int
	Purchases []Product
}

func (u User) String() string {
	return fmt.Sprintf("User{id=%d, name=%q, age=%d, purchases=%d}", u.ID, u.Name, u.Age, len(u.Purchases))
}
