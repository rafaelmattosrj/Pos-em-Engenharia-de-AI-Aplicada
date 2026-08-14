package core

import "sync/atomic"

const circuitBreakerThreshold = 3

// CircuitBreaker evita loops infinitos com respostas inválidas do LLM —
// quebra após 3 respostas inválidas consecutivas. Equivalente a
// CircuitBreaker.java.
type CircuitBreaker struct {
	invalidCount atomic.Int32
}

// NewCircuitBreaker cria um CircuitBreaker zerado.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{}
}

// RecordInvalid registra uma resposta inválida do LLM.
func (c *CircuitBreaker) RecordInvalid() {
	c.invalidCount.Add(1)
}

// RecordValid registra uma resposta válida, resetando o contador.
func (c *CircuitBreaker) RecordValid() {
	c.invalidCount.Store(0)
}

// ShouldBreak retorna true se o número de falhas consecutivas atingiu o
// threshold.
func (c *CircuitBreaker) ShouldBreak() bool {
	return c.invalidCount.Load() >= circuitBreakerThreshold
}

// Reset volta o circuit breaker ao estado inicial.
func (c *CircuitBreaker) Reset() {
	c.invalidCount.Store(0)
}

// InvalidCount retorna o número atual de respostas inválidas consecutivas.
func (c *CircuitBreaker) InvalidCount() int {
	return int(c.invalidCount.Load())
}
