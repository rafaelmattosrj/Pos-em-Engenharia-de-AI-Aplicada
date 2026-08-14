package gateway

import "strings"

// ClassificarIntencao (Módulo 4.2 — Intent-Based Routing): classificador de
// regra determinística, a ponta mais simples da escada descrita no 4.2,
// suficiente pra separar os únicos 2 fluxos que pedem tratamento diferente aqui.
func ClassificarIntencao(pergunta string) string {
	p := strings.ToLower(pergunta)
	if strings.Contains(p, "csr") || strings.Contains(p, "relatório final") ||
		strings.Contains(p, "síntese") || strings.Contains(p, "evento adverso") ||
		strings.Contains(p, "desfecho") {
		return "sintese_csr"
	}
	if strings.Contains(p, "critério") || strings.Contains(p, "inclusão") ||
		strings.Contains(p, "exclusão") || strings.Contains(p, "idade mínima") ||
		strings.Contains(p, "protocolo") {
		return "consulta_protocolo"
	}
	return "consulta_icf"
}
