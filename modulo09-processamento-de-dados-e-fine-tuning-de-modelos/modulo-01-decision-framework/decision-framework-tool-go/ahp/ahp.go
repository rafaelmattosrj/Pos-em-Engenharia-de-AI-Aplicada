// Package ahp implementa o Analytic Hierarchy Process (Saaty, 1980): deriva
// pesos de uma matriz de comparação pareada pela média geométrica das linhas,
// calcula a Razão de Consistência (CR) do julgamento, e agrega o julgamento de
// vários avaliadores (comitê) numa única matriz. Equivalente a derivarPesosAHP,
// calcularConsistenciaAHP e agregarMatrizesComite em decision-framework-tool.js.
package ahp

import "math"

// RandomIndexN4 é o Random Index de Saaty (n=4), usado pra normalizar o
// Índice de Consistência em Razão de Consistência.
const RandomIndexN4 = 0.90

// Consistencia é o resultado de calcularConsistencia.
type Consistencia struct {
	LambdaMax   float64
	CI          float64
	CR          float64
	Consistente bool
}

// DerivarPesos retorna o vetor de prioridades (pesos, soma 1.0) pelo método da
// média geométrica das linhas (row geometric mean method).
func DerivarPesos(matriz [][]float64) []float64 {
	n := len(matriz)
	mediasGeometricas := make([]float64, n)
	for i, linha := range matriz {
		produto := 1.0
		for _, v := range linha {
			produto *= v
		}
		mediasGeometricas[i] = math.Pow(produto, 1.0/float64(n))
	}
	soma := 0.0
	for _, v := range mediasGeometricas {
		soma += v
	}
	pesos := make([]float64, n)
	for i, v := range mediasGeometricas {
		pesos[i] = v / soma
	}
	return pesos
}

// CalcularConsistencia calcula a Razão de Consistência (CR); CR < 0.10 é o
// limiar padrão de Saaty pra considerar o julgamento consistente.
func CalcularConsistencia(matriz [][]float64, pesos []float64) Consistencia {
	n := len(matriz)
	aw := make([]float64, n)
	for i, linha := range matriz {
		soma := 0.0
		for j, v := range linha {
			soma += v * pesos[j]
		}
		aw[i] = soma
	}
	somaRazoes := 0.0
	for i := range aw {
		somaRazoes += aw[i] / pesos[i]
	}
	lambdaMax := somaRazoes / float64(n)
	ci := (lambdaMax - float64(n)) / float64(n-1)
	cr := ci / RandomIndexN4
	return Consistencia{LambdaMax: lambdaMax, CI: ci, CR: cr, Consistente: cr < 0.10}
}

// AgregarMatrizesComite agrega várias matrizes de comparação pareada (uma por
// avaliador) pela média geométrica célula a célula (método AIJ) -- preserva a
// propriedade recíproca (a[j][i] = 1/a[i][j]), então o resultado ainda é uma
// matriz de comparação pareada válida. Um comitê de 1 avaliador reduz ao caso
// original.
func AgregarMatrizesComite(matrizes [][][]float64) [][]float64 {
	nAvaliadores := len(matrizes)
	n := len(matrizes[0])
	agregada := make([][]float64, n)
	for i := 0; i < n; i++ {
		agregada[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			produto := 1.0
			for _, matriz := range matrizes {
				produto *= matriz[i][j]
			}
			agregada[i][j] = math.Pow(produto, 1.0/float64(nAvaliadores))
		}
	}
	return agregada
}
