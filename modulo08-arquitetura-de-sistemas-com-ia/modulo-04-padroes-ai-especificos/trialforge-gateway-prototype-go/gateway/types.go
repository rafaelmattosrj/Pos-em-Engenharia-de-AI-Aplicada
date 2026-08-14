// Package gateway implementa a lógica de negócio do Gateway do TrialForge —
// porte de trialforge-gateway-prototype.js / trialforge_gateway_prototype.py.
// Reúne os 4 grupos de padrão do Módulo 4: RAG (Multi-Index + Hybrid Search +
// Agentic RAG), Intent-Based Routing + Model Router, Semantic Cache + Response
// Streaming, Confidence Threshold + Approval Gate + Audit Trail.
package gateway

// Clausula é uma cláusula regulatória de um dos índices do RAG.
type Clausula struct {
	Tema  string
	Texto string
	Fonte string
}

// Message é uma mensagem de chat (role/content) no formato aceito pelo Ollama.
type Message struct {
	Role    string
	Content string
}
