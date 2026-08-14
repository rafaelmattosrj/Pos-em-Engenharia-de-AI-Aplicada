// Package ml implementa a rede neural feedforward do zero em Go — porte de
// NeuralNetwork.java, ModelContext.java e RecommendationEngine.java.
package ml

import (
	"fmt"
	"math"
	"math/rand"
)

// NeuralNetwork é uma rede feedforward treinada via SGD + backpropagation,
// com loss de Binary Cross-Entropy — mesma arquitetura da versão Java:
// camadas ocultas com ReLU, camada de saída com Sigmoid.
type NeuralNetwork struct {
	// weights[l][i][j] = peso da camada l, neurônio i, entrada j.
	weights [][][]float64
	biases  [][]float64
	// activations[l] é "relu" para camadas ocultas e "sigmoid" para a última.
	activations  []string
	learningRate float64

	// Cache para backpropagation.
	preActivations  [][]float64 // valor ANTES da função de ativação
	postActivations [][]float64 // valor DEPOIS da função de ativação (saída)

	rng *rand.Rand
}

// NewNeuralNetwork cria a rede com o formato de camadas informado (ex.:
// [20, 128, 64, 32, 1]) e a taxa de aprendizado dada. Usa seed fixo (42) na
// inicialização dos pesos, para reprodutibilidade — igual à versão Java.
func NewNeuralNetwork(layerSizes []int, learningRate float64) *NeuralNetwork {
	numLayers := len(layerSizes) - 1

	nn := &NeuralNetwork{
		weights:         make([][][]float64, numLayers),
		biases:          make([][]float64, numLayers),
		activations:     make([]string, numLayers),
		preActivations:  make([][]float64, numLayers),
		postActivations: make([][]float64, numLayers+1),
		learningRate:    learningRate,
		rng:             rand.New(rand.NewSource(42)),
	}

	for l := 0; l < numLayers; l++ {
		in := layerSizes[l]
		out := layerSizes[l+1]

		nn.weights[l] = make([][]float64, out)
		nn.biases[l] = make([]float64, out)

		// Inicialização He: escala = sqrt(2 / entradas). Funciona bem com
		// ReLU, evita vanishing/exploding gradients.
		scale := math.Sqrt(2.0 / float64(in))
		for i := 0; i < out; i++ {
			nn.weights[l][i] = make([]float64, in)
			for j := 0; j < in; j++ {
				nn.weights[l][i][j] = nn.rng.NormFloat64() * scale
			}
		}

		if l == numLayers-1 {
			nn.activations[l] = "sigmoid"
		} else {
			nn.activations[l] = "relu"
		}
	}

	return nn
}

func relu(x float64) float64 {
	return math.Max(0, x)
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func reluDeriv(preAct float64) float64 {
	if preAct > 0 {
		return 1.0
	}
	return 0.0
}

func sigmoidDeriv(preAct float64) float64 {
	s := sigmoid(preAct)
	return s * (1.0 - s)
}

// Forward propaga a entrada pela rede e retorna a predição (0-1), salvando
// os valores intermediários necessários para o Backward subsequente.
func (nn *NeuralNetwork) Forward(input []float64) float64 {
	nn.postActivations[0] = input
	current := input

	for l := 0; l < len(nn.weights); l++ {
		outSize := len(nn.weights[l])
		pre := make([]float64, outSize)
		post := make([]float64, outSize)

		for i := 0; i < outSize; i++ {
			pre[i] = nn.biases[l][i]
			for j := 0; j < len(current); j++ {
				pre[i] += nn.weights[l][i][j] * current[j]
			}
			if nn.activations[l] == "relu" {
				post[i] = relu(pre[i])
			} else {
				post[i] = sigmoid(pre[i])
			}
		}

		nn.preActivations[l] = pre
		nn.postActivations[l+1] = post
		current = post
	}

	return current[0] // saída binária única
}

// Backward executa backpropagation com SGD. Deve ser chamado sempre APÓS um
// Forward() com o mesmo exemplo.
func (nn *NeuralNetwork) Backward(label float64) {
	numLayers := len(nn.weights)
	deltas := make([][]float64, numLayers)

	// Camada de saída: a derivada da binary cross-entropy + sigmoid juntas
	// simplifica para (pred - label).
	pred := nn.postActivations[numLayers][0]
	preOut := nn.preActivations[numLayers-1][0]
	deltas[numLayers-1] = []float64{(pred - label) * sigmoidDeriv(preOut)}

	// Camadas ocultas, de trás para frente.
	for l := numLayers - 2; l >= 0; l-- {
		size := len(nn.weights[l])
		deltas[l] = make([]float64, size)

		for i := 0; i < size; i++ {
			sum := 0.0
			for k := 0; k < len(deltas[l+1]); k++ {
				sum += nn.weights[l+1][k][i] * deltas[l+1][k]
			}
			deltas[l][i] = sum * reluDeriv(nn.preActivations[l][i])
		}
	}

	// Atualizar pesos e biases.
	for l := 0; l < numLayers; l++ {
		prevOutput := nn.postActivations[l]
		for i := range nn.weights[l] {
			for j := range nn.weights[l][i] {
				nn.weights[l][i][j] -= nn.learningRate * deltas[l][i] * prevOutput[j]
			}
			nn.biases[l][i] -= nn.learningRate * deltas[l][i]
		}
	}
}

// Train treina a rede por `epochs` épocas, embaralhando os índices a cada
// época (equivalente ao shuffle: true do TF.js) e imprimindo loss/acurácia a
// cada 10 épocas — mesmo comportamento observável da versão Java.
func (nn *NeuralNetwork) Train(inputs [][]float64, labels []float64, epochs int) {
	fmt.Println()
	fmt.Println("=== Iniciando Treinamento da Rede Neural ===")
	fmt.Printf("Exemplos: %d | Epocas: %d | Learning Rate: %.3f\n\n", len(inputs), epochs, nn.learningRate)

	n := len(inputs)
	shuffleRng := rand.New(rand.NewSource(rand.Int63()))

	for epoch := 0; epoch < epochs; epoch++ {
		totalLoss := 0.0
		correct := 0

		indices := shuffledIndices(n, shuffleRng)

		for _, idx := range indices {
			pred := nn.Forward(inputs[idx])
			label := labels[idx]

			nn.Backward(label)

			loss := -(label*math.Log(pred+1e-7) + (1-label)*math.Log(1-pred+1e-7))
			totalLoss += loss

			if (pred >= 0.5 && label == 1.0) || (pred < 0.5 && label == 0.0) {
				correct++
			}
		}

		if (epoch+1)%10 == 0 {
			avgLoss := totalLoss / float64(n)
			accuracy := float64(correct) / float64(n) * 100
			fmt.Printf("Epoca %3d | Loss: %.4f | Acuracia: %.1f%%\n", epoch+1, avgLoss, accuracy)
		}
	}

	fmt.Println()
	fmt.Println("=== Treinamento Concluido ===")
	fmt.Println()
}

func shuffledIndices(n int, rng *rand.Rand) []int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	rng.Shuffle(n, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}
