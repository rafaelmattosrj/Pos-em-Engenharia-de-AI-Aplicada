package grpo

import (
	"fmt"
	"strconv"
	"strings"
)

// Gabarito é a resposta correta pro documento de sinistro de demo.
var Gabarito = map[string]any{
	"claimant_name":        "Marcos Vinícius Almeida Teixeira",
	"claim_type":           "Colisao veicular",
	"incident_date":        "12/07/2026",
	"estimated_amount_brl": 8450.0,
	"status":               "Em analise pericial",
	"prioridade":           "alta",
}

// RecompensaVerificavel é a fração de campos exatamente corretos -- o "grader"
// que o GRPO e o RFT usam, sem modelo de recompensa aprendido. Equivalente a
// recompensaVerificavel() em JS.
func RecompensaVerificavel(candidato map[string]any) float64 {
	if candidato == nil {
		return 0.0
	}
	acertos := 0
	for campo, esperado := range Gabarito {
		obtido := candidato[campo]
		if esperadoNum, ok := esperado.(float64); ok {
			if numObtido, ok := paraFloat(obtido); ok && abs(numObtido-esperadoNum) < 0.01 {
				acertos++
			}
		} else {
			obtidoStr := ""
			if obtido != nil {
				obtidoStr = fmt.Sprint(obtido)
			}
			if strings.EqualFold(strings.TrimSpace(obtidoStr), strings.TrimSpace(fmt.Sprint(esperado))) {
				acertos++
			}
		}
	}
	return float64(acertos) / float64(len(Gabarito))
}

func paraFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
