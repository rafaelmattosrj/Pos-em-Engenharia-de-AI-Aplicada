package tiering

import "strings"

const (
	SinteseCSR       = "sintese_csr"
	ConsultaClausula = "consulta_clausula"
)

// ClassificarIntencao é lógica pura, sem rede.
func ClassificarIntencao(pergunta string) string {
	p := strings.ToLower(pergunta)
	if strings.Contains(p, "csr") || strings.Contains(p, "relatório final") || strings.Contains(p, "síntese") {
		return SinteseCSR
	}
	return ConsultaClausula
}
