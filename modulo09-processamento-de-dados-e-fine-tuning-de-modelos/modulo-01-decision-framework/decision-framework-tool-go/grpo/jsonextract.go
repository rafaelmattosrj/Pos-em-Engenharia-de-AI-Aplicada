// Package grpo reproduz o mecanismo real do GRPO (Group Relative Policy
// Optimization) até o ponto do sinal de aprendizado (amostragem de grupo ->
// recompensa verificável -> vantagem relativa ao grupo) -- porte de
// grpo-verifiable-reward-demo.js. Não treina nada: a atualização de peso via
// gradiente de política fica fora de escopo, de propósito.
package grpo

import (
	"encoding/json"
	"regexp"
)

var objetoJSON = regexp.MustCompile(`(?s)\{.*\}`)

// ExtrairJSON extrai o primeiro objeto JSON solto num texto, ou nil se não
// achar/não for válido -- equivalente a extrairJson() em JS.
func ExtrairJSON(texto string) map[string]any {
	match := objetoJSON.FindString(texto)
	if match == "" {
		return nil
	}
	var candidato map[string]any
	if err := json.Unmarshal([]byte(match), &candidato); err != nil {
		return nil
	}
	return candidato
}
