// Package data carrega o catálogo de produtos e usuários de demonstração —
// equivalente ao DataLoader.java (mesmos dados, mesma ordem).
package data

import "ecommerce-neural-recommender/model"

// LoadProducts retorna o catálogo fixo de produtos de demonstração.
func LoadProducts() []model.Product {
	return []model.Product{
		{ID: 1, Name: "Fones de Ouvido Sem Fio", Category: "eletronicos", Price: 129.99, Color: "preto"},
		{ID: 2, Name: "Relogio Inteligente", Category: "eletronicos", Price: 199.99, Color: "prata"},
		{ID: 3, Name: "Caixa de Som Bluetooth", Category: "eletronicos", Price: 89.99, Color: "azul"},
		{ID: 4, Name: "Camiseta Estampada", Category: "vestuario", Price: 49.99, Color: "branco"},
		{ID: 5, Name: "Calca Jeans Slim", Category: "vestuario", Price: 99.99, Color: "azul"},
		{ID: 6, Name: "Tenis Esportivo", Category: "calcados", Price: 149.99, Color: "vermelho"},
		{ID: 7, Name: "Sandalia Casual", Category: "calcados", Price: 69.99, Color: "bege"},
		{ID: 8, Name: "Bone Estiloso", Category: "acessorios", Price: 39.99, Color: "preto"},
		{ID: 9, Name: "Mochila Executiva", Category: "acessorios", Price: 159.99, Color: "cinza"},
		{ID: 10, Name: "Oculos de Sol", Category: "acessorios", Price: 89.99, Color: "marrom"},
	}
}

// LoadUsers retorna os usuários fixos de demonstração, cada um com histórico
// de compras montado a partir de LoadProducts().
func LoadUsers() []model.User {
	products := LoadProducts()

	return []model.User{
		{ID: 1, Name: "Ana Lima", Age: 25, Purchases: []model.Product{products[0], products[1]}},
		{ID: 2, Name: "Bruno Ferreira", Age: 27, Purchases: []model.Product{products[0], products[2]}},
		{ID: 3, Name: "Camila Souza", Age: 30, Purchases: []model.Product{products[3], products[4]}},
		{ID: 4, Name: "Diego Almeida", Age: 22, Purchases: []model.Product{products[1], products[2], products[5]}},
		{ID: 5, Name: "Eduarda Nunes", Age: 28, Purchases: []model.Product{products[0], products[5], products[4]}},
	}
}
