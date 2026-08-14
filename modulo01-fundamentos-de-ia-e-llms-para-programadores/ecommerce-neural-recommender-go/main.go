// Sistema de recomendação de e-commerce — porte Go do projeto Java
// equivalente (ecommerce-neural-recommender-java), que por sua vez traduz o
// Exemplo 01 original (JavaScript/TensorFlow.js). Rede neural feedforward
// implementada do zero, sem nenhuma biblioteca externa de ML.
package main

import (
	"fmt"
	"strings"

	"ecommerce-neural-recommender/data"
	"ecommerce-neural-recommender/ml"
	"ecommerce-neural-recommender/model"
)

func main() {
	fmt.Println(strings.Repeat("=", 53))
	fmt.Println("  Sistema de Recomendacao de E-commerce - Go")
	fmt.Println("  (Traducao do Exemplo 01 - JavaScript/TensorFlow.js)")
	fmt.Println(strings.Repeat("=", 53))
	fmt.Println()

	products := data.LoadProducts()
	users := data.LoadUsers()

	fmt.Printf("Produtos carregados: %d\n", len(products))
	fmt.Printf("Usuarios carregados: %d\n\n", len(users))

	engine := &ml.RecommendationEngine{}
	engine.TrainModel(products, users)

	fmt.Println(strings.Repeat("=", 53))
	fmt.Println("  RECOMENDACOES POR USUARIO")
	fmt.Println(strings.Repeat("=", 53))

	for _, user := range users {
		printRecommendations(engine, user, 3)
	}

	fmt.Println(strings.Repeat("=", 53))
	fmt.Println("  USUARIO NOVO (sem historico de compras)")
	fmt.Println(strings.Repeat("=", 53))

	newUser := model.User{ID: 99, Name: "Rafael Souza", Age: 27}
	printRecommendations(engine, newUser, 3)

	fmt.Println(strings.Repeat("=", 53))
	fmt.Println("  USUARIO COM HISTORICO DE ACESSORIOS")
	fmt.Println(strings.Repeat("=", 53))

	accessoriesUser := model.User{
		ID: 100, Name: "Marina Costa", Age: 29,
		Purchases: []model.Product{products[7], products[8]}, // Bone Estiloso, Mochila Executiva
	}
	printRecommendations(engine, accessoriesUser, 3)
}

func printRecommendations(engine *ml.RecommendationEngine, user model.User, topN int) {
	fmt.Printf("\n Usuario: %s (idade %d)\n", user.Name, user.Age)

	if len(user.Purchases) == 0 {
		fmt.Println("  Historico: nenhuma compra")
	} else {
		names := make([]string, len(user.Purchases))
		for i, p := range user.Purchases {
			names[i] = p.Name
		}
		fmt.Println("  Historico:", strings.Join(names, ", "))
	}

	fmt.Printf("  Top %d Recomendacoes:\n", topN)
	recommendations, err := engine.Recommend(user)
	if err != nil {
		fmt.Println("  Erro:", err)
		return
	}

	if topN > len(recommendations) {
		topN = len(recommendations)
	}
	for _, rec := range recommendations[:topN] {
		fmt.Println(rec)
	}
}
