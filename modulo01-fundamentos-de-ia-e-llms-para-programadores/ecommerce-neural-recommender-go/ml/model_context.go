package ml

import (
	"sort"

	"ecommerce-neural-recommender/model"
)

// Pesos de cada feature na composição do vetor — espelho do objeto WEIGHTS
// da versão Java/JS.
const (
	WeightCategory = 0.4
	WeightColor    = 0.3
	WeightPrice    = 0.2
	WeightAge      = 0.1
)

// ModelContext contém os metadados de codificação calculados a partir dos
// dados de treino: limites de normalização, índices de one-hot encoding e
// idade média normalizada por produto — equivalente a ModelContext.java.
type ModelContext struct {
	Products []model.Product
	Users    []model.User

	MinAge, MaxAge     float64
	MinPrice, MaxPrice float64

	CategoriesIndex map[string]int
	ColorsIndex     map[string]int

	NumCategories int
	NumColors     int

	// Dimensions = price(1) + age(1) + categorias + cores.
	Dimensions int

	// Idade média normalizada de compradores por produto.
	ProductAvgAgeNorm map[string]float64

	// Vetores pré-calculados de cada produto (preenchidos pelo
	// RecommendationEngine para evitar reprocessamento).
	ProductVectors [][]float64
}

// NewModelContext calcula todos os metadados de codificação a partir dos
// produtos e usuários informados.
func NewModelContext(products []model.Product, users []model.User) *ModelContext {
	ctx := &ModelContext{Products: products, Users: users}

	ctx.MinAge, ctx.MaxAge = ageBounds(users)
	ctx.MinPrice, ctx.MaxPrice = priceBounds(products)

	categories := distinctSorted(products, func(p model.Product) string { return p.Category })
	colors := distinctSorted(products, func(p model.Product) string { return p.Color })

	ctx.CategoriesIndex = indexOf(categories)
	ctx.ColorsIndex = indexOf(colors)
	ctx.NumCategories = len(categories)
	ctx.NumColors = len(colors)
	ctx.Dimensions = 2 + ctx.NumCategories + ctx.NumColors

	ctx.ProductAvgAgeNorm = productAvgAgeNorm(products, users, ctx.MinAge, ctx.MaxAge)

	return ctx
}

// Normalize normaliza um valor para o intervalo [0, 1]: (val - min) / (max - min).
func Normalize(value, min, max float64) float64 {
	rangeVal := max - min
	if rangeVal == 0 {
		return 0
	}
	return (value - min) / rangeVal
}

func ageBounds(users []model.User) (min, max float64) {
	if len(users) == 0 {
		return 0, 100
	}
	min, max = float64(users[0].Age), float64(users[0].Age)
	for _, u := range users[1:] {
		age := float64(u.Age)
		if age < min {
			min = age
		}
		if age > max {
			max = age
		}
	}
	return min, max
}

func priceBounds(products []model.Product) (min, max float64) {
	if len(products) == 0 {
		return 0, 1000
	}
	min, max = products[0].Price, products[0].Price
	for _, p := range products[1:] {
		if p.Price < min {
			min = p.Price
		}
		if p.Price > max {
			max = p.Price
		}
	}
	return min, max
}

func distinctSorted(products []model.Product, key func(model.Product) string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range products {
		k := key(p)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func indexOf(values []string) map[string]int {
	idx := make(map[string]int, len(values))
	for i, v := range values {
		idx[v] = i
	}
	return idx
}

func productAvgAgeNorm(products []model.Product, users []model.User, minAge, maxAge float64) map[string]float64 {
	ageSums := map[string]float64{}
	ageCounts := map[string]int{}

	for _, user := range users {
		for _, p := range user.Purchases {
			ageSums[p.Name] += float64(user.Age)
			ageCounts[p.Name]++
		}
	}

	midAge := (minAge + maxAge) / 2.0
	result := make(map[string]float64, len(products))
	for _, product := range products {
		var avg float64
		if count, ok := ageCounts[product.Name]; ok {
			avg = ageSums[product.Name] / float64(count)
		} else {
			avg = midAge
		}
		result[product.Name] = Normalize(avg, minAge, maxAge)
	}
	return result
}
