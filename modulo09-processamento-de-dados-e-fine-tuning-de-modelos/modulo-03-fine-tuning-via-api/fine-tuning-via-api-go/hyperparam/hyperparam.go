// Package hyperparam valida hiperparametro de fine-tuning ANTES do envio
// (Modulo 3.3/3.4): a Vertex AI aceita epochCount=0 sem erro rapido e
// substitui por um default silencioso, entao a validacao real precisa
// acontecer aqui. Tambem compara o hiperparametro pedido com o aplicado
// (campos int64 voltam como string no JSON da API).
package hyperparam

import (
	"fmt"
	"strconv"
)

type Hiperparametros struct {
	EpochCount             int
	LearningRateMultiplier float64
}

const (
	EpochCountMin = 1
	EpochCountMax = 20
	LRMin         = 0.1
	LRMax         = 10
)

func Validar(h Hiperparametros) error {
	var erros []string
	if h.EpochCount < EpochCountMin || h.EpochCount > EpochCountMax {
		erros = append(erros, fmt.Sprintf("epochCount deve ser inteiro entre %d e %d, recebido: %d", EpochCountMin, EpochCountMax, h.EpochCount))
	}
	if h.LearningRateMultiplier < LRMin || h.LearningRateMultiplier > LRMax {
		erros = append(erros, fmt.Sprintf("learningRateMultiplier deve estar entre %v e %v, recebido: %v", LRMin, LRMax, h.LearningRateMultiplier))
	}
	if len(erros) > 0 {
		msg := "Hiperparâmetro inválido, job não enviado:\n"
		for _, e := range erros {
			msg += "  " + e + "\n"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// ValoresEquivalentes tolera a diferenca de tipo que a Vertex AI introduz de
// verdade: campos int64 (epochCount) voltam como string no JSON, nao como
// numero.
func ValoresEquivalentes(a, b any) bool {
	na, oka := paraNumero(a)
	nb, okb := paraNumero(b)
	if oka && okb {
		return na == nb
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func paraNumero(v any) (float64, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case string:
		n, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

// CompararHiperparametros compara o hiperparametro pedido com o realmente
// aplicado (campo supervisedTuningSpec.hyperParameters do job).
func CompararHiperparametros(pedido, aplicado map[string]any) []string {
	var divergencias []string
	for chave, valorPedido := range pedido {
		valorAplicado, existe := aplicado[chave]
		if !existe || valorAplicado == nil {
			divergencias = append(divergencias, fmt.Sprintf("%s: pedido %v, aplicado ausente (provavelmente default silencioso do provedor)", chave, valorPedido))
		} else if !ValoresEquivalentes(valorAplicado, valorPedido) {
			divergencias = append(divergencias, fmt.Sprintf("%s: pedido %v, aplicado %v", chave, valorPedido, valorAplicado))
		}
	}
	return divergencias
}
