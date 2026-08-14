package ml

import (
	"errors"
	"fmt"
	"sort"

	"ecommerce-neural-recommender/model"
)

// RecommendationEngine orquestra: criação do contexto, codificação de
// features, criação dos dados de treino, treinamento da rede neural e
// geração de recomendações — equivalente a RecommendationEngine.java.
type RecommendationEngine struct {
	model   *NeuralNetwork
	context *ModelContext
}

// Recommendation associa um produto ao score previsto pelo modelo.
type Recommendation struct {
	Product model.Product
	Score   float64
}

func (r Recommendation) String() string {
	return fmt.Sprintf("  [%.4f] %s (R$ %.2f | %s | %s)",
		r.Score, r.Product.Name, r.Product.Price, r.Product.Category, r.Product.Color)
}

// ErrModelNotTrained é retornado por Recommend quando TrainModel ainda não
// foi chamado.
var ErrModelNotTrained = errors.New("modelo nao treinado: chame TrainModel primeiro")

// TrainModel constrói o contexto de codificação e treina a rede neural com
// arquitetura [inputDim, 128, 64, 32, 1] por 100 épocas — mesmos parâmetros
// da versão Java.
func (e *RecommendationEngine) TrainModel(products []model.Product, users []model.User) {
	fmt.Println("Construindo contexto de codificacao...")
	e.context = NewModelContext(products, users)

	fmt.Printf("Categorias encontradas: %v\n", sortedKeys(e.context.CategoriesIndex))
	fmt.Printf("Cores encontradas:      %v\n", sortedKeys(e.context.ColorsIndex))
	fmt.Printf("Dimensoes do vetor:     %d (price + age + %d categorias + %d cores)\n\n",
		e.context.Dimensions, e.context.NumCategories, e.context.NumColors)

	e.context.ProductVectors = make([][]float64, len(products))
	for i, p := range products {
		e.context.ProductVectors[i] = e.EncodeProduct(p)
	}

	inputs, labels := e.createTrainingData()

	inputDim := e.context.Dimensions * 2 // user vector + product vector
	e.model = NewNeuralNetwork([]int{inputDim, 128, 64, 32, 1}, 0.01)

	e.model.Train(inputs, labels, 100)
}

// EncodeProduct codifica um produto como vetor numérico:
// [price_norm*0.2, avgAge_norm*0.1, cat_one_hot*0.4..., color_one_hot*0.3...]
func (e *RecommendationEngine) EncodeProduct(product model.Product) []float64 {
	ctx := e.context
	vector := make([]float64, ctx.Dimensions)
	idx := 0

	vector[idx] = Normalize(product.Price, ctx.MinPrice, ctx.MaxPrice) * WeightPrice
	idx++

	avgAgeNorm, ok := ctx.ProductAvgAgeNorm[product.Name]
	if !ok {
		avgAgeNorm = 0.5
	}
	vector[idx] = avgAgeNorm * WeightAge
	idx++

	catIndex, hasCat := ctx.CategoriesIndex[product.Category]
	for i := 0; i < ctx.NumCategories; i++ {
		if hasCat && i == catIndex {
			vector[idx] = WeightCategory
		}
		idx++
	}

	colorIndex, hasColor := ctx.ColorsIndex[product.Color]
	for i := 0; i < ctx.NumColors; i++ {
		if hasColor && i == colorIndex {
			vector[idx] = WeightColor
		}
		idx++
	}

	return vector
}

// EncodeUser codifica um usuário como vetor numérico: média dos vetores dos
// produtos comprados, ou apenas a idade normalizada se não houver histórico.
func (e *RecommendationEngine) EncodeUser(user model.User) []float64 {
	ctx := e.context

	if len(user.Purchases) == 0 {
		vector := make([]float64, ctx.Dimensions)
		vector[1] = Normalize(float64(user.Age), ctx.MinAge, ctx.MaxAge) * WeightAge
		return vector
	}

	sum := make([]float64, ctx.Dimensions)
	for _, purchase := range user.Purchases {
		prodVector := e.EncodeProduct(purchase)
		for i := range sum {
			sum[i] += prodVector[i]
		}
	}

	n := float64(len(user.Purchases))
	for i := range sum {
		sum[i] /= n
	}

	return sum
}

func (e *RecommendationEngine) createTrainingData() ([][]float64, []float64) {
	ctx := e.context
	var inputs [][]float64
	var labels []float64

	for _, user := range ctx.Users {
		if len(user.Purchases) == 0 {
			continue
		}

		userVector := e.EncodeUser(user)

		for pi, product := range ctx.Products {
			productVector := ctx.ProductVectors[pi]

			purchased := false
			for _, p := range user.Purchases {
				if p.Name == product.Name {
					purchased = true
					break
				}
			}

			combined := make([]float64, ctx.Dimensions*2)
			copy(combined[:ctx.Dimensions], userVector)
			copy(combined[ctx.Dimensions:], productVector)

			inputs = append(inputs, combined)
			if purchased {
				labels = append(labels, 1.0)
			} else {
				labels = append(labels, 0.0)
			}
		}
	}

	fmt.Printf("Dados de treino criados: %d exemplos\n", len(inputs))
	return inputs, labels
}

// Recommend gera recomendações para um usuário, ordenadas do maior para o
// menor score previsto pelo modelo treinado.
func (e *RecommendationEngine) Recommend(user model.User) ([]Recommendation, error) {
	if e.model == nil {
		return nil, ErrModelNotTrained
	}

	ctx := e.context
	userVector := e.EncodeUser(user)

	recommendations := make([]Recommendation, 0, len(ctx.Products))
	for pi, product := range ctx.Products {
		productVector := ctx.ProductVectors[pi]

		combined := make([]float64, ctx.Dimensions*2)
		copy(combined[:ctx.Dimensions], userVector)
		copy(combined[ctx.Dimensions:], productVector)

		score := e.model.Forward(combined)
		recommendations = append(recommendations, Recommendation{Product: product, Score: score})
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations, nil
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
