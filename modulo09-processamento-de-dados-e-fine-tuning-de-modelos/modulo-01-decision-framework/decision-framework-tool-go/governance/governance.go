// Package governance implementa o gate de governança/compliance (LGPD), que
// roda ANTES de tudo. Diferente das 4 perguntas ponderadas por AHP, governança
// é binária: reprova o caso ali, sem gastar o resto do pipeline (AHP, NPV,
// Monte Carlo, Real Options). Equivalente a validarGovernancaDado() em
// decision-framework-tool.js.
package governance

import "decision-framework-tool/config"

// Resultado é o veredito do gate de governança.
type Resultado struct {
	Aprovado bool
	Motivos  []string
}

// Validar checa base legal e (quando o dado é sensível) DPA assinado.
func Validar(g config.Governanca) Resultado {
	var motivos []string

	if !g.BaseLegalDefinida {
		motivos = append(motivos, "sem base legal definida pro tratamento do dado (LGPD Art. 7º/11)")
	}
	if g.DadoSensivelLGPD && !g.DpaAssinado {
		motivos = append(motivos, "dado de categoria sensível (LGPD Art. 5º, II) sem DPA assinado com o provedor de fine-tuning")
	}

	return Resultado{Aprovado: len(motivos) == 0, Motivos: motivos}
}
