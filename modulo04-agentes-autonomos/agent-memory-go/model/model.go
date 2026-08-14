// Package model define os tipos de domínio do sistema de memória —
// equivalente a Episode.java, Lesson.java, MemoryFact.java,
// AgentRunRequest.java e AgentRunResult.java.
package model

import "time"

// Episode representa um episódio de execução do agente — uma interação
// completa com input, passos e resultado.
type Episode struct {
	ID               string    `json:"id"`
	Input            string    `json:"input"`
	Steps            []string  `json:"steps"`
	Outcome          string    `json:"outcome"`
	LessonsExtracted []string  `json:"lessonsExtracted"`
	Timestamp        time.Time `json:"timestamp"`
}

// Lesson representa uma lição aprendida extraída pela ReflectionEngine após
// uma execução.
type Lesson struct {
	ID               string    `json:"id"`
	Situation        string    `json:"situation"`
	Action           string    `json:"action"`
	Result           string    `json:"result"`
	Learning         string    `json:"learning"`
	Generalizability string    `json:"generalizability"`
	ExtractedAt      time.Time `json:"extractedAt"`
}

// MemoryFact representa um fato persistido na memória de longo prazo.
type MemoryFact struct {
	ID          string     `json:"id"`
	Content     string     `json:"content"`
	Source      string     `json:"source"`
	ConfirmedAt time.Time  `json:"confirmedAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
}

// AgentRunRequest é a requisição para execução do agente com ou sem
// memória.
type AgentRunRequest struct {
	Input string `json:"input"`
}

// AgentRunResult é o resultado da execução do agente, incluindo métricas de
// uso de memória para comparação.
type AgentRunResult struct {
	Output             string `json:"output"`
	MemoryUsed         bool   `json:"memoryUsed"`
	FactsRetrieved     int    `json:"factsRetrieved"`
	EpisodesConsidered int    `json:"episodesConsidered"`
}
