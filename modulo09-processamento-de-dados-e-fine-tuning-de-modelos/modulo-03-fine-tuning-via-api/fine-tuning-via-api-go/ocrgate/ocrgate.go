// Package ocrgate implementa o gate de confianca de OCR (Modulo 3.2):
// exemplo cujo metadata.confiancaOcr esta abaixo do limiar e sinalizado pra
// revisao humana; exemplo sem confiancaOcr (texto sintetico, nunca
// escaneado) segue aprovado sem passar pelo gate.
package ocrgate

const LimiarPadrao = 0.85

type Exemplo struct {
	ID            string
	ConfiancaOcr  *float64
}

type Resultado struct {
	Aprovados             []Exemplo
	AprovadosPorOcr       []Exemplo
	SemConfianca          []Exemplo
	SinalizadosParaRevisao []Exemplo
	Limiar                float64
}

func Filtrar(exemplos []Exemplo, limiar ...float64) Resultado {
	l := LimiarPadrao
	if len(limiar) > 0 {
		l = limiar[0]
	}
	var semConfianca, aprovadosPorOcr, sinalizados []Exemplo
	for _, e := range exemplos {
		switch {
		case e.ConfiancaOcr == nil:
			semConfianca = append(semConfianca, e)
		case *e.ConfiancaOcr >= l:
			aprovadosPorOcr = append(aprovadosPorOcr, e)
		default:
			sinalizados = append(sinalizados, e)
		}
	}
	aprovados := append(append([]Exemplo{}, aprovadosPorOcr...), semConfianca...)
	return Resultado{
		Aprovados:              aprovados,
		AprovadosPorOcr:        aprovadosPorOcr,
		SemConfianca:           semConfianca,
		SinalizadosParaRevisao: sinalizados,
		Limiar:                 l,
	}
}
